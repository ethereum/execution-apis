package specgen

import (
	"fmt"
	"strings"

	"github.com/mattn/go-jsonpointer"
)

// Dereference recursively expands all '$ref' entries in schema, resolving each
// reference against repository. Every '$ref' value must have the form
// "#/components/schemas/<Name>".
//
// The input schema is not modified; a fully dereferenced deep copy is returned.
// Cycles are detected and reported as errors.
func Dereference(schema object, repository map[string]object) (object, error) {
	d := &dereferencer{
		repository: repository,
		visiting:   make(map[string]bool),
	}
	return d.object(schema)
}

// dereferencer holds the state needed to recursively expand $ref entries.
type dereferencer struct {
	repository map[string]object
	visiting   map[string]bool
}

func (d *dereferencer) value(v any) (any, error) {
	switch val := v.(type) {
	case map[string]any:
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
// "$ref" key its siblings are ignored (standard JSON Schema $ref semantics)
// and the referenced schema is returned instead, recursively dereferenced.
func (d *dereferencer) object(obj object) (object, error) {
	if ref, ok := obj["$ref"].(string); ok {
		return d.resolveRef(ref)
	}

	out := make(object, len(obj))
	for k, v := range obj {
		expanded, err := d.value(v)
		if err != nil {
			return nil, fmt.Errorf(".%s%w", k, err)
		}
		out[k] = expanded
	}
	return out, nil
}

func (d *dereferencer) resolveRef(ref string) (object, error) {
	name, path, err := parseSchemaRef(ref)
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
	if len(path) > 0 {
		pointer := "/" + strings.Join(path, "/")
		val, err := jsonpointer.Get(schema, pointer)
		if err != nil {
			return nil, fmt.Errorf("$ref %q: %w", ref, err)
		}
		node, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("$ref %q: value at %q is not an object", ref, pointer)
		}
		// Dereference the sub-node without cycle-guard (it's a value, not a
		// whole schema definition, so cycles through it are impossible).
		return d.object(node)
	}

	d.visiting[cycleKey] = true
	result, err := d.object(schema)
	d.visiting[cycleKey] = false
	return result, err
}

// parseSchemaRef parses a $ref of the form "#/components/schemas/<Name>[/...]".
// It returns the schema name and any remaining path segments (may be empty).
func parseSchemaRef(ref string) (name string, path []string, err error) {
	const prefix = "#/components/schemas/"
	if !strings.HasPrefix(ref, prefix) {
		return "", nil, fmt.Errorf("unsupported $ref format: %q", ref)
	}
	rest := strings.TrimPrefix(ref, prefix)
	parts := strings.SplitN(rest, "/", 2)
	name = parts[0]
	if name == "" {
		return "", nil, fmt.Errorf("unsupported $ref format: %q", ref)
	}
	if len(parts) == 2 && parts[1] != "" {
		path = strings.Split(parts[1], "/")
	}
	return name, path, nil
}

// extractSchemaName parses a local $ref like "#/components/schemas/Foo" and
// returns the schema name ("Foo"). Returns an error for unsupported formats or
// if a sub-path is present.
func extractSchemaName(ref string) (string, error) {
	name, path, err := parseSchemaRef(ref)
	if err != nil {
		return "", err
	}
	if len(path) > 0 {
		return "", fmt.Errorf("unsupported $ref format: %q", ref)
	}
	return name, nil
}
