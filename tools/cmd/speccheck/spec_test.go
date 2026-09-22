package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/execution-apis/tools/internal/specgen"
)

func TestPreserveArraySchemas(t *testing.T) {
	for _, schema := range []string{
		`{"type":"array","items":{"type":"integer"}}`,
		`{"type":"array","items":[{"type":"integer"},{"type":"string"}],"minItems":2,"maxItems":2}`,
	} {
		data := `{"methods":[{"name":"example","params":[],"result":{"name":"Result","schema":` + schema + `}}]}`
		path := filepath.Join(t.TempDir(), "spec.json")
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
		parsed, err := parseSpec(path)
		if err != nil {
			t.Fatal(err)
		}
		got := parsed["example"].result.schema
		if string(got) != schema {
			t.Fatalf("schema changed during parsing: %s", got)
		}
		invalid := []byte(`[1,false]`)
		if err := validate(got, invalid, "https://example.test/array.json"); err == nil {
			t.Fatal("second item escaped validation")
		}
	}
}

func TestSimulationFailureMessages(t *testing.T) {
	// Fix the schema/prose mismatch exposed by validating every array element:
	// error codes are normative, but the schema says messages are suggestions.
	g := specgen.New()
	for _, name := range []string{"base-types.yaml", "execute.yaml"} {
		data, err := os.ReadFile(filepath.Join("../../../src/schemas", name))
		if err != nil {
			t.Fatal(err)
		}
		if err := g.AddSchemas(data); err != nil {
			t.Fatal(err)
		}
	}
	data, err := g.JSON()
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	root := map[string]any{"$ref": "#/components/schemas/CallResultFailure", "components": doc["components"]}
	schema, err := json.Marshal(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		error string
		valid bool
	}{
		{`{"code":-32015,"message":"out of gas"}`, true},
		{`{"code":3,"message":"execution reverted"}`, true},
		{`{"code":3,"message":"custom revert message"}`, true},
		{`{"code":-1,"message":"out of gas"}`, false},
		{`{"message":"out of gas"}`, false},
		{`{"code":-32015}`, false},
	} {
		value := []byte(`{"status":"0x0","returnData":"0x","gasUsed":"0x0","error":` + tc.error + `}`)
		if err := validate(schema, value, "https://example.test/simulation.json"); (err == nil) != tc.valid {
			t.Fatalf("valid=%v for %s: %v", tc.valid, tc.error, err)
		}
	}
}
