package specgen

import (
	"strings"
	"testing"
)

// verifies that a schema that is itself a $ref gets replaced by the referenced schema
func TestDereference_TopLevelRef(t *testing.T) {
	repository := schemaRepository{"Foo": {"type": "string", "title": "Foo"}}
	schema := object{"$ref": "#/components/schemas/Foo"}

	got, err := dereference(schema, repository)
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
	repository := schemaRepository{"Address": {"type": "object", "title": "Address"}}
	schema := object{
		"type": "object",
		"properties": object{
			"addr": object{"$ref": "#/components/schemas/Address"},
		},
	}

	got, err := dereference(schema, repository)
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
	repository := schemaRepository{
		"Inner": {"type": "boolean"},
		"Outer": {"$ref": "#/components/schemas/Inner"},
	}
	schema := object{"$ref": "#/components/schemas/Outer"}

	got, err := dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	if got["type"] != "boolean" {
		t.Errorf("type: want boolean, got %v", got["type"])
	}
}

// checks that $ref entries inside an array (e.g. oneOf / anyOf / allOf) are expanded
func TestDereference_RefInArray(t *testing.T) {
	repository := schemaRepository{
		"A": {"type": "string"},
		"B": {"type": "integer"},
	}
	schema := object{
		"oneOf": []any{
			object{"$ref": "#/components/schemas/A"},
			object{"$ref": "#/components/schemas/B"},
		},
	}

	got, err := dereference(schema, repository)
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
	_, err := dereference(schema, schemaRepository{})
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
	_, err := dereference(schema, schemaRepository{})
	if err == nil {
		t.Fatal("expected error for bad $ref format, got nil")
	}
}

// verifies that a cycle in $ref references returns an error rather than looping forever
func TestDereference_CycleDetection(t *testing.T) {
	// A -> B -> A
	repository := schemaRepository{
		"A": {"$ref": "#/components/schemas/B"},
		"B": {"$ref": "#/components/schemas/A"},
	}
	schema := object{"$ref": "#/components/schemas/A"}

	_, err := dereference(schema, repository)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle") {
		t.Errorf("error should mention 'cycle', got: %v", err)
	}
}

// verifies that a schema directly referencing itself returns a cycle error
func TestDereference_SelfCycle(t *testing.T) {
	repository := schemaRepository{"Self": {"$ref": "#/components/schemas/Self"}}
	schema := object{"$ref": "#/components/schemas/Self"}

	_, err := dereference(schema, repository)
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

// This verifies that a diamond dependency
// (A -> B, A -> C, B -> D, C -> D) does not produce a false cycle error.
// D is referenced by two independent paths but is not actually cyclic.
func TestDereference_DiamondShape(t *testing.T) {
	repository := schemaRepository{
		"D": {"type": "string"},
		"B": {"items": object{"$ref": "#/components/schemas/D"}},
		"C": {"items": object{"$ref": "#/components/schemas/D"}},
		"A": {
			"properties": object{
				"b": object{"$ref": "#/components/schemas/B"},
				"c": object{"$ref": "#/components/schemas/C"},
			},
		},
	}
	schema := object{"$ref": "#/components/schemas/A"}

	got, err := dereference(schema, repository)
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
	repository := schemaRepository{
		"Base": {
			"type": "object",
			"properties": object{
				"x": object{"type": "string", "title": "X field"},
				"y": object{"type": "integer", "title": "Y field"},
			},
		},
	}
	// Child reuses Base's properties via sub-path $ref.
	schema := object{
		"type": "object",
		"properties": object{
			"x": object{"$ref": "#/components/schemas/Base/properties/x"},
			"y": object{"$ref": "#/components/schemas/Base/properties/y"},
		},
	}
	got, err := dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	props := got["properties"].(object)
	xProp := props["x"].(object)
	if xProp["type"] != "string" {
		t.Errorf("x.type: want string, got %v", xProp["type"])
	}
	if xProp["title"] != "X field" {
		t.Errorf("x.title: want 'X field', got %v", xProp["title"])
	}
	yProp := props["y"].(object)
	if yProp["type"] != "integer" {
		t.Errorf("y.type: want integer, got %v", yProp["type"])
	}
}

// TestDereference_SubPathRefChained verifies that a sub-path ref that itself
// contains a $ref is transitively expanded (matching the PayloadAttributesV* pattern).
func TestDereference_SubPathRefChained(t *testing.T) {
	repository := schemaRepository{
		"V1": {
			"type": "object",
			"properties": object{
				"val": object{"$ref": "#/components/schemas/Scalar"},
			},
		},
		"Scalar": {"type": "number"},
		"V2": {
			"type": "object",
			"properties": object{
				// reuses V1's property via sub-path ref
				"val": object{"$ref": "#/components/schemas/V1/properties/val"},
			},
		},
	}
	schema := object{"$ref": "#/components/schemas/V2"}
	got, err := dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	props := got["properties"].(object)
	val := props["val"].(object)
	// The sub-path ref pointed to {"$ref": "#/components/schemas/Scalar"},
	// which should itself be expanded to {"type": "number"}.
	if val["type"] != "number" {
		t.Errorf("val.type: want number, got %v", val["type"])
	}
}

// TestDereference_RefSiblingOverride verifies that when a $ref appears
// alongside sibling keys (e.g. title + $ref), the sibling values win over
// the fields of the resolved schema. This matches the OpenAPI convention used
// throughout the spec YAML files, e.g.:
//
//	hash:
//	  title: Hash          ← local sibling, must win
//	  $ref: '#/components/schemas/hash32'
func TestDereference_RefSiblingOverride(t *testing.T) {
	repository := schemaRepository{
		"BaseType": {
			"type":    "string",
			"title":   "generic title from base",
			"pattern": "^0x[0-9a-f]{64}$",
		},
	}
	// Property object with both a local title and a $ref.
	schema := object{
		"properties": object{
			"hash": object{
				"title": "Hash",
				"$ref":  "#/components/schemas/BaseType",
			},
		},
	}

	got, err := dereference(schema, repository)
	if err != nil {
		t.Fatal(err)
	}
	hash := got["properties"].(object)["hash"].(object)

	// Local sibling wins for title.
	if hash["title"] != "Hash" {
		t.Errorf("title: want 'Hash' (local sibling), got %v", hash["title"])
	}
	// Fields from the resolved schema that have no local sibling are present.
	if hash["type"] != "string" {
		t.Errorf("type: want string (from ref), got %v", hash["type"])
	}
	if hash["pattern"] != "^0x[0-9a-f]{64}$" {
		t.Errorf("pattern: want ref pattern, got %v", hash["pattern"])
	}
	// $ref itself must not appear in the output.
	if _, has := hash["$ref"]; has {
		t.Error("$ref should not appear in the expanded output")
	}
}
