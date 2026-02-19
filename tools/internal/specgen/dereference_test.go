package specgen

import (
	"strings"
	"testing"
)

// repo builds a repository from alternating name/schema pairs.
func repo(pairs ...any) map[string]object {
	r := make(map[string]object, len(pairs)/2)
	for i := 0; i < len(pairs)-1; i += 2 {
		r[pairs[i].(string)] = pairs[i+1].(object)
	}
	return r
}

// verifies that a schema that is itself a $ref gets replaced by the referenced schema
func TestDereference_TopLevelRef(t *testing.T) {
	repository := repo("Foo", object{"type": "string", "title": "Foo"})
	schema := object{"$ref": "#/components/schemas/Foo"}

	got, err := Dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	if got["type"] != "string" {
		t.Errorf("type: want string, got %v", got["type"])
	}
	if got["title"] != "Foo" {
		t.Errorf("title: want Foo, got %v", got["title"])
	}
}

// verifies that a $ref nested inside a property is expanded correctly.
func TestDereference_NestedRef(t *testing.T) {
	repository := repo("Address", object{"type": "object", "title": "Address"})
	schema := object{
		"type": "object",
		"properties": object{
			"addr": object{"$ref": "#/components/schemas/Address"},
		},
	}

	got, err := Dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	props := got["properties"].(object)
	addr := props["addr"].(object)
	if addr["type"] != "object" {
		t.Errorf("addr.type: want object, got %v", addr["type"])
	}
	if addr["title"] != "Address" {
		t.Errorf("addr.title: want Address, got %v", addr["title"])
	}
}

// verifies that a $ref pointing to a schema that itself contains a $ref is expanded transitively
func TestDereference_ChainedRefs(t *testing.T) {
	repository := repo(
		"Inner", object{"type": "boolean"},
		"Outer", object{"$ref": "#/components/schemas/Inner"},
	)
	schema := object{"$ref": "#/components/schemas/Outer"}

	got, err := Dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	if got["type"] != "boolean" {
		t.Errorf("type: want boolean, got %v", got["type"])
	}
}

// checks that $ref entries inside an array (e.g. oneOf / anyOf / allOf) are expanded
func TestDereference_RefInArray(t *testing.T) {
	repository := repo(
		"A", object{"type": "string"},
		"B", object{"type": "integer"},
	)
	schema := object{
		"oneOf": []any{
			object{"$ref": "#/components/schemas/A"},
			object{"$ref": "#/components/schemas/B"},
		},
	}

	got, err := Dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	oneOf := got["oneOf"].([]any)
	if len(oneOf) != 2 {
		t.Fatalf("oneOf: want 2 entries, got %d", len(oneOf))
	}
	if oneOf[0].(object)["type"] != "string" {
		t.Errorf("oneOf[0].type: want string, got %v", oneOf[0].(object)["type"])
	}
	if oneOf[1].(object)["type"] != "integer" {
		t.Errorf("oneOf[1].type: want integer, got %v", oneOf[1].(object)["type"])
	}
}

// checks that a $ref to a missing schema returns an error
func TestDereference_UnknownRef(t *testing.T) {
	schema := object{"$ref": "#/components/schemas/Missing"}
	_, err := Dereference(schema, repo())
	if err == nil {
		t.Fatal("expected error for unknown $ref, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should mention 'not found', got: %v", err)
	}
}

// verfies that a $ref with an unsupported format returns an error
func TestDereference_BadRefFormat(t *testing.T) {
	schema := object{"$ref": "http://example.com/schema"}
	_, err := Dereference(schema, repo())
	if err == nil {
		t.Fatal("expected error for bad $ref format, got nil")
	}
}

// verifies that a cycle in $ref references returns an error rather than looping forever
func TestDereference_CycleDetection(t *testing.T) {
	// A -> B -> A
	repository := repo(
		"A", object{"$ref": "#/components/schemas/B"},
		"B", object{"$ref": "#/components/schemas/A"},
	)
	schema := object{"$ref": "#/components/schemas/A"}

	_, err := Dereference(schema, repository)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should mention 'cycle', got: %v", err)
	}
}

// verifies that a schema directly referencing itself returns a cycle error
func TestDereference_SelfCycle(t *testing.T) {
	repository := repo("Self", object{"$ref": "#/components/schemas/Self"})
	schema := object{"$ref": "#/components/schemas/Self"}

	_, err := Dereference(schema, repository)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

// This verifies that a diamond dependency
// (A -> B, A -> C, B -> D, C -> D) does not produce a false cycle error.
// D is referenced by two independent paths but is not actually cyclic.
func TestDereference_DiamondShape(t *testing.T) {
	repository := repo(
		"D", object{"type": "string"},
		"B", object{"items": object{"$ref": "#/components/schemas/D"}},
		"C", object{"items": object{"$ref": "#/components/schemas/D"}},
		"A", object{
			"properties": object{
				"b": object{"$ref": "#/components/schemas/B"},
				"c": object{"$ref": "#/components/schemas/C"},
			},
		},
	)
	schema := object{"$ref": "#/components/schemas/A"}

	got, err := Dereference(schema, repository)
	if err != nil {
		t.Fatalf("unexpected error for diamond shape: %v", err)
	}
	props := got["properties"].(object)
	bItems := props["b"].(object)["items"].(object)
	if bItems["type"] != "string" {
		t.Errorf("b.items.type: want string, got %v", bItems["type"])
	}
	cItems := props["c"].(object)["items"].(object)
	if cItems["type"] != "string" {
		t.Errorf("c.items.type: want string, got %v", cItems["type"])
	}
}

// TestDereference_SubPathRef verifies that a $ref pointing into a sub-path of
// another schema (e.g. "#/components/schemas/Foo/properties/bar") is resolved
// by navigating into that schema's property.
func TestDereference_SubPathRef(t *testing.T) {
	repository := repo(
		"Base", object{
			"type": "object",
			"properties": map[string]any{
				"x": object{"type": "string", "title": "X field"},
				"y": object{"type": "integer", "title": "Y field"},
			},
		},
	)
	// Child reuses Base's properties via sub-path $ref.
	schema := object{
		"type": "object",
		"properties": map[string]any{
			"x": object{"$ref": "#/components/schemas/Base/properties/x"},
			"y": object{"$ref": "#/components/schemas/Base/properties/y"},
		},
	}
	got, err := Dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	props := got["properties"].(map[string]any)
	xProp := props["x"].(map[string]any)
	if xProp["type"] != "string" {
		t.Errorf("x.type: want string, got %v", xProp["type"])
	}
	if xProp["title"] != "X field" {
		t.Errorf("x.title: want 'X field', got %v", xProp["title"])
	}
	yProp := props["y"].(map[string]any)
	if yProp["type"] != "integer" {
		t.Errorf("y.type: want integer, got %v", yProp["type"])
	}
}

// TestDereference_SubPathRefChained verifies that a sub-path ref that itself
// contains a $ref is transitively expanded (matching the PayloadAttributesV* pattern).
func TestDereference_SubPathRefChained(t *testing.T) {
	repository := repo(
		"V1", object{
			"type": "object",
			"properties": map[string]any{
				"val": object{"$ref": "#/components/schemas/Scalar"},
			},
		},
		"Scalar", object{"type": "number"},
		"V2", object{
			"type": "object",
			"properties": map[string]any{
				// reuses V1's property via sub-path ref
				"val": object{"$ref": "#/components/schemas/V1/properties/val"},
			},
		},
	)
	schema := object{"$ref": "#/components/schemas/V2"}
	got, err := Dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	props := got["properties"].(map[string]any)
	val := props["val"].(map[string]any)
	// The sub-path ref pointed to {"$ref": "#/components/schemas/Scalar"},
	// which should itself be expanded to {"type": "number"}.
	if val["type"] != "number" {
		t.Errorf("val.type: want number, got %v", val["type"])
	}
}
