package spec

import (
	"errors"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestPreserveArraySchemas(t *testing.T) {
	for _, tc := range []struct {
		schema string
		valid  []any
	}{
		{`{"type":"array","items":{"type":"integer"}}`, []any{1, 2}},
		{`{"type":"array","items":[{"type":"integer"},{"type":"string"}],"minItems":2,"maxItems":2}`, []any{1, "a"}},
	} {
		data := `{"openrpc":"1.2.4","info":{"title":"t","version":"1"},"methods":[{"name":"example","params":[],"result":{"name":"Result","schema":` + tc.schema + `}}]}`
		s, err := LoadString("spec.json", data)
		if err != nil {
			t.Fatal(err)
		}
		if err := s.Validate("#/methods/0/result/schema", tc.valid); err != nil {
			t.Fatalf("%s: valid array rejected: %v", tc.schema, err)
		}
		var invalid *jsonschema.ValidationError
		if err := s.Validate("#/methods/0/result/schema", []any{1, false}); !errors.As(err, &invalid) {
			t.Fatalf("%s: second item escaped validation: %v", tc.schema, err)
		}
	}
}

func TestExactIntegerBounds(t *testing.T) {
	data := `{"openrpc":"1.2.4","info":{"title":"t","version":"1"},"methods":[{"name":"example","params":[{"name":"n","required":true,"schema":{"type":"integer","maximum":18446744073709551615}}],"result":{"name":"Result","schema":{}}}]}`
	s, err := LoadString("spec.json", data)
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		param string
		valid bool
	}{
		{"18446744073709551615", true},
		{"18446744073709551616", false},
	} {
		checker := NewChecker(s)
		checker.SetCheckArgs(true)
		err := checker.Request([]byte(`{"jsonrpc":"2.0","id":1,"method":"example","params":[` + tc.param + `]}`))
		if (err == nil) != tc.valid {
			t.Fatalf("%s: valid=%v, got %v", tc.param, tc.valid, err)
		}
	}
}

func TestCheckerRejectsTrailingData(t *testing.T) {
	data := `{"openrpc":"1.2.4","info":{"title":"t","version":"1"},"methods":[{"name":"example","params":[],"result":{"name":"Result","schema":{}}}]}`
	s, err := LoadString("spec.json", data)
	if err != nil {
		t.Fatal(err)
	}
	checker := NewChecker(s)
	if err := checker.Request([]byte(`{"jsonrpc":"2.0","id":1,"method":"example","params":[]} {}`)); err == nil {
		t.Fatal("request with trailing data accepted")
	}
	if err := checker.Request([]byte(`{"jsonrpc":"2.0","id":2,"method":"example","params":[]}`)); err != nil {
		t.Fatal(err)
	}
	if err := checker.Response([]byte(`{"jsonrpc":"2.0","id":2,"result":1} x`)); err == nil {
		t.Fatal("response with trailing data accepted")
	}
}
