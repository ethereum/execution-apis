package spec

import "testing"

func TestPreserveArraySchemas(t *testing.T) {
	for _, schema := range []string{
		`{"type":"array","items":{"type":"integer"}}`,
		`{"type":"array","items":[{"type":"integer"},{"type":"string"}],"minItems":2,"maxItems":2}`,
	} {
		data := `{"openrpc":"1.2.4","info":{"title":"t","version":"1"},"methods":[{"name":"example","params":[],"result":{"name":"Result","schema":` + schema + `}}]}`
		s, err := LoadString("spec.json", data)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Validate("#/methods/0/result/schema", []any{1.0, false}); err == nil {
			t.Fatal("second item escaped validation")
		}
	}
}
