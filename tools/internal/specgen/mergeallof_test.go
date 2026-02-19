package specgen

import (
	"slices"
	"testing"
)

func toStringSlice(v any) []string {
	arr, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, item := range arr {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// tests merging a single inline allOf entry
func TestMergeAllOf_SimpleInline(t *testing.T) {
	schema := object{
		"type":  "object",
		"title": "Child",
		"allOf": []any{
			object{
				"required":   []any{"a"},
				"properties": object{"a": object{"type": "string"}},
			},
		},
	}
	got := mergeAllOf(schema)

	if _, has := got["allOf"]; has {
		t.Error("allOf should be removed")
	}
	if got["type"] != "object" {
		t.Errorf("type: want object, got %v", got["type"])
	}
	if got["title"] != "Child" {
		t.Errorf("title: want Child, got %v", got["title"])
	}
	req := toStringSlice(got["required"])
	if !slices.Contains(req, "a") {
		t.Errorf("required should contain 'a', got %v", req)
	}
	props := got["properties"].(object)
	if _, ok := props["a"]; !ok {
		t.Error("property 'a' should be present")
	}
}

// checks allOf with multiple inline entries.
func TestMergeAllOf_MultipleEntries(t *testing.T) {
	schema := object{
		"type": "object",
		"allOf": []any{
			object{
				"required":   []any{"x"},
				"properties": object{"x": object{"type": "string"}},
			},
			object{
				"required":   []any{"y"},
				"properties": object{"y": object{"type": "integer"}},
			},
		},
	}
	got := mergeAllOf(schema)

	req := toStringSlice(got["required"])
	for _, r := range []string{"x", "y"} {
		if !slices.Contains(req, r) {
			t.Errorf("required should contain %q, got %v", r, req)
		}
	}
	props := got["properties"].(object)
	for _, p := range []string{"x", "y"} {
		if _, ok := props[p]; !ok {
			t.Errorf("property %q should be present", p)
		}
	}
}

// verifies that parent fields take precedence over values from allOf entries
func TestMergeAllOf_ParentWins(t *testing.T) {
	schema := object{
		"type":  "object",
		"title": "Parent Title",
		"allOf": []any{
			object{"type": "string", "title": "Override Title"},
		},
	}
	got := mergeAllOf(schema)
	if got["type"] != "object" {
		t.Errorf("type: parent should win, got %v", got["type"])
	}
	if got["title"] != "Parent Title" {
		t.Errorf("title: parent should win, got %v", got["title"])
	}
}

// verifies that parent properties win over same-named properties in allOf entries
func TestMergeAllOf_PropertyConflict(t *testing.T) {
	schema := object{
		"type": "object",
		"properties": object{
			"shared": object{"type": "integer", "title": "From Parent"},
		},
		"allOf": []any{
			object{
				"properties": object{
					"shared": object{"type": "string", "title": "From AllOf"},
				},
			},
		},
	}
	got := mergeAllOf(schema)
	props := got["properties"].(object)
	shared := props["shared"].(object)
	if shared["type"] != "integer" {
		t.Errorf("type: parent property should win, got %v", shared["type"])
	}
	if shared["title"] != "From Parent" {
		t.Errorf("title: parent property should win, got %v", shared["title"])
	}
}

// verifies that duplicate required fields are deduplicated after merging
func TestMergeAllOf_RequiredDedup(t *testing.T) {
	schema := object{
		"type":     "object",
		"required": []any{"a", "b"},
		"allOf": []any{
			object{"required": []any{"b", "c"}},
		},
	}
	got := mergeAllOf(schema)
	req := toStringSlice(got["required"])
	counts := make(map[string]int)
	for _, r := range req {
		counts[r]++
	}
	if counts["b"] > 1 {
		t.Errorf("'b' should appear once in required, got %d times", counts["b"])
	}
	for _, r := range []string{"a", "b", "c"} {
		if counts[r] == 0 {
			t.Errorf("required should contain %q", r)
		}
	}
}

// verifies that allOf nested inside a property value is also merged
func TestMergeAllOf_NestedAllOf(t *testing.T) {
	schema := object{
		"type": "object",
		"properties": object{
			"inner": object{
				"type": "object",
				"allOf": []any{
					object{"required": []any{"z"}},
				},
			},
		},
	}
	got := mergeAllOf(schema)
	inner := got["properties"].(object)["inner"].(object)
	if _, has := inner["allOf"]; has {
		t.Error("nested allOf should be removed")
	}
	req := toStringSlice(inner["required"])
	if !slices.Contains(req, "z") {
		t.Errorf("nested required should contain 'z', got %v", req)
	}
}
