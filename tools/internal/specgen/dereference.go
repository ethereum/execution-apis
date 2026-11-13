package specgen

import (
	"fmt"
	"strings"

	"github.com/mattn/go-jsonpointer"
)

// dereference recursively expands all fragment '$ref' entries in schema,
// resolving each reference against repository. Fragment '$ref' values must
// have the form "#/components/schemas/<Name>"; absolute-URI refs are kept
// verbatim (see object).
//
// The input schema is not modified; a dereferenced deep copy is returned.
// Cycles are detected and reported as errors.
func dereference(schema object, repository schemaRepository) (object, error) {
	d := &dereferencer{
		repository: repository,
		visiting:   make(map[string]bool),
		ids:        collectIDs(repository),
	}
	return d.object(schema)
}

// dereferencer holds the state needed to recursively expand $ref entries.
type dereferencer struct {
	repository schemaRepository
	visiting   map[string]bool
	ids        map[string]bool
}

// collectIDs gathers every embedded $id in the repository. These are the only
// valid targets for absolute-URI $refs, so typo'd refs fail the build instead
// of shipping as dangling links.
func collectIDs(repository schemaRepository) map[string]bool {
	ids := make(map[string]bool)
	var walk func(v any)
	walk = func(v any) {
		switch val := v.(type) {
		case object:
			if id, ok := val["$id"].(string); ok {
				ids[id] = true
			}
			for _, item := range val {
				walk(item)
			}
		case []any:
			for _, item := range val {
				walk(item)
			}
		}
	}
	for _, schema := range repository {
		walk(schema)
	}
	return ids
}

func (d *dereferencer) value(v any) (any, error) {
	switch val := v.(type) {
	case object:
		return d.object(val)
	case []any:
		return d.slice(val)
	default:
		// Scalar (string, bool, float64, nil, …) – nothing to expand.
		return v, nil
	}
}

func (d *dereferencer) slice(arr []any) ([]any, error) {
	out := make([]any, len(arr))
	for i, item := range arr {
		expanded, err := d.value(item)
		if err != nil {
			return nil, fmt.Errorf("[%d]: %w", i, err)
		}
		out[i] = expanded
	}
	return out, nil
}

// object returns a dereferenced deep copy of obj. If the object contains a
// "$ref" key, the referenced schema is resolved and
// returned as the base, with any sibling keys merged on top (siblings win on
// conflict). This matches the OpenAPI convention where $ref siblings act as
// local overrides.
func (d *dereferencer) object(obj object) (object, error) {
	ref, hasRef := obj["$ref"].(string)

	// Expand all non-$ref values.
	out := make(object, len(obj))
	for k, v := range obj {
		if k == "$ref" {
			continue
		}
		expanded, err := d.value(v)
		if err != nil {
			return nil, fmt.Errorf(".%s%w", k, err)
		}
		out[k] = expanded
	}

	if hasRef {
		// Absolute-URI refs target an embedded $id (used for recursive schemas
		// like CallFrame) and are kept verbatim; they resolve at validation
		// time via JSON Schema $id semantics. Both ref kinds fail the build
		// when the target does not exist.
		if !strings.HasPrefix(ref, "#") {
			if !d.ids[ref] {
				return nil, fmt.Errorf("$ref %q: no schema with a matching $id", ref)
			}
			out["$ref"] = ref
			return out, nil
		}
		// Resolve the reference and use it as the base, letting local siblings
		// (already in out) take precedence over the resolved fields.
		base, err := d.resolveRef(ref)
		if err != nil {
			return nil, err
		}
		for k, v := range base {
			if _, exists := out[k]; !exists {
				out[k] = v
			}
		}
	}

	return out, nil
}

func (d *dereferencer) resolveRef(ref string) (object, error) {
	name, pointer, err := parseSchemaRef(ref)
	if err != nil {
		return nil, err
	}

	// Use name-only key for cycle detection since sub-path refs are not cyclic
	// by themselves — only whole-schema references can form cycles.
	cycleKey := name
	if d.visiting[cycleKey] {
		return nil, fmt.Errorf("$ref %q: cycle detected", ref)
	}
	schema, ok := d.repository[name]
	if !ok {
		return nil, fmt.Errorf("$ref %q: schema not found in repository", ref)
	}

	// Navigate into sub-path if present.
	if pointer != "" {
		val, err := jsonpointer.Get(schema, pointer)
		if err != nil {
			return nil, fmt.Errorf("$ref %q: %w", ref, err)
		}
		node, ok := val.(object)
		if !ok {
			return nil, fmt.Errorf("$ref %q: pointed-to value at %q is not an object", ref, pointer)
		}
		return d.object(node)
	}

	d.visiting[cycleKey] = true
	result, err := d.object(schema)
	d.visiting[cycleKey] = false
	return result, err
}

// parseSchemaRef parses a $ref of the form "#/components/schemas/<Name>[/...]".
// It returns the schema name and any remaining path segments (may be empty).
func parseSchemaRef(ref string) (name, path string, err error) {
	ref, valid := strings.CutPrefix(ref, "#/components/schemas/")
	if !valid {
		return "", "", fmt.Errorf("unsupported $ref format: %q", ref)
	}
	name, path, hasPath := strings.Cut(ref, "/")
	if hasPath {
		path = "/" + path
	}
	return name, path, nil
}
