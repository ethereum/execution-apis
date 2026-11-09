package specgen

import (
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestDereferenceResourceLocalRecursion(t *testing.T) {
	repo := schemaRepository{"Node": {
		"$id":      "https://example.test/node.json",
		"type":     "object",
		"required": []any{"value"},
		"properties": object{
			"value": object{"type": "integer"},
			"child": object{"$ref": "#"},
		},
	}}
	// Exercise embedding: # must resolve to the node resource, not the envelope.
	input := object{"type": "object", "properties": object{"node": object{"$ref": "#/components/schemas/Node"}}}
	got, err := (&Generator{}).expandSchema(input, repo)
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("https://example.test/root.json", got); err != nil {
		t.Fatal(err)
	}
	compiled, err := compiler.Compile("https://example.test/root.json")
	if err != nil {
		t.Fatal(err)
	}
	good := object{"node": object{"value": 1, "child": object{"value": 2}}}
	if err := compiled.Validate(good); err != nil {
		t.Fatal(err)
	}
	bad := object{"node": object{"value": 1, "child": object{"value": "wrong"}}}
	if err := compiled.Validate(bad); err == nil {
		t.Fatal("nested child escaped validation")
	}
}

func TestDereferenceRejectsUnscopedSelfReference(t *testing.T) {
	if _, err := dereference(object{"$ref": "#"}, schemaRepository{}); err == nil {
		t.Fatal("unscoped self-reference must be rejected")
	}
}
