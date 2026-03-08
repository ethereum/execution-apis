package specgen

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

const errorGroupRefPrefix = "#/components/error-groups/"

// errorGroups is a collection of error group definitions
type errorGroups []errorGroup

// errorGroup defines a single error group
type errorGroup struct {
	Name   string     `yaml:"group"`
	Range  groupRange `yaml:"range"`
	Errors []Error    `yaml:"errors"`
}

type groupRange struct {
	Min *int `yaml:"min"`
	Max *int `yaml:"max"`
}

// Error is an individual error entry within an error group
type Error struct {
	Code    int    `yaml:"code"`
	Message string `yaml:"message"`
	Data    any    `yaml:"data"`
}

func (groups errorGroups) find(name string) (errorGroup, bool) {
	for _, g := range groups {
		if g.Name == name {
			return g, true
		}
	}
	return errorGroup{}, false
}

// resolveMethodErrors merges a method's existing errors with resolved error-group refs.
// IMPORTANT: Inline errors take precedence: group errors with a matching code are skipped
func (groups errorGroups) resolveMethodErrors(existingErrors []any, errorGroupRefs []any) ([]any, error) {
	var merged []any
	merged = append(merged, existingErrors...)

	// Collect inline error codes so group errors don't override them.
	inlineCodes := make(map[int]bool)
	for _, e := range existingErrors {
		if obj, ok := e.(object); ok {
			if code, ok := obj["code"].(int); ok {
				inlineCodes[code] = true
			}
		}
	}

	for i, ref := range errorGroupRefs {
		obj, ok := ref.(object)
		if !ok {
			return nil, fmt.Errorf("error-groups[%d]: expected object with $ref, got %T", i, ref)
		}
		ref, isRef := obj["$ref"].(string)
		if !isRef {
			return nil, fmt.Errorf("error-groups[%d]: expected $ref string", i)
		}
		name, err := parseErrorGroupRef(ref)
		if err != nil {
			return nil, fmt.Errorf("error-groups[%d]: %w", i, err)
		}
		group, ok := groups.find(name)
		if !ok {
			return nil, fmt.Errorf("error-groups[%d]: $ref %q: error group not found", i, ref)
		}
		for j, e := range group.Errors {
			if err := group.validateErrorCode(e); err != nil {
				return nil, fmt.Errorf("error-groups[%d] (group %s)[%d]: %w", i, name, j, err)
			}
			if inlineCodes[e.Code] {
				continue
			}
			merged = append(merged, e.toObject())
		}
	}
	return merged, nil
}

// validateErrorCode checks that the error code falls within the group's range.
func (g errorGroup) validateErrorCode(err Error) error {
	if g.Range.Min == nil || g.Range.Max == nil {
		return nil
	}
	if err.Code < *g.Range.Min || err.Code > *g.Range.Max {
		return fmt.Errorf("error code %d out of range [%d, %d] for group %s",
			err.Code, *g.Range.Min, *g.Range.Max, g.Name)
	}
	return nil
}

// toObject converts an Error to a generic object for the output spec
func (e Error) toObject() object {
	obj := object{
		"code":    e.Code,
		"message": e.Message,
	}
	if e.Data != nil {
		obj["data"] = e.Data
	}
	return obj
}

// parseErrorGroups converts YAML content to errorGroups
func parseErrorGroups(content []byte) (errorGroups, error) {
	var body map[string]errorGroup
	if err := yaml.Unmarshal(content, &body); err != nil {
		return nil, fmt.Errorf("invalid error group content: %v", err)
	}
	var groups errorGroups
	for name, group := range body {
		group.Name = name
		groups = append(groups, group)
	}
	return groups, nil
}

// parseErrorGroupRef parses a $ref
func parseErrorGroupRef(ref string) (string, error) {
	name, valid := strings.CutPrefix(ref, errorGroupRefPrefix)
	if !valid {
		return "", fmt.Errorf("unsupported error group $ref format: %q", ref)
	}
	if name == "" {
		return "", fmt.Errorf("empty error group name in $ref: %q", ref)
	}
	return name, nil
}
