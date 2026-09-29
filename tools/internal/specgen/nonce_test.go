package specgen

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestGetTransactionCountNonceKeys(t *testing.T) {
	generator := New()
	files, err := filepath.Glob("../../../src/schemas/*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddSchemas(data); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile("../../../src/eth/state.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := generator.AddMethods(data); err != nil {
		t.Fatal(err)
	}
	method := generator.methods["eth_getTransactionCount"]
	params := method["params"].([]any)
	if len(params) != 3 || params[2].(object)["name"] != "nonceKeys" {
		t.Fatalf("expected nonceKeys as the third parameter: %v", params)
	}
	var items []any
	minItems := 0
	for i, raw := range params {
		param := raw.(object)
		items = append(items, param["schema"])
		if param["required"] == true {
			minItems = i + 1
		}
	}
	generator.types["NonceQuery"] = object{"type": "array", "items": items, "minItems": minItems, "additionalItems": false}
	generator.types["NonceResult"] = method["result"].(object)["schema"].(object)
	address := "0xc94770007dda54cF92009BFF0dE90c06F603a09f"
	type testCase struct {
		name, schema string
		value        any
		valid        bool
	}
	tests := []testCase{
		{"default block", "NonceQuery", []any{address}, true},
		{"legacy request", "NonceQuery", []any{address, "latest"}, true},
		{"missing address", "NonceQuery", []any{}, false},
		{"invalid address", "NonceQuery", []any{"0x1", "latest", []any{"0x1"}}, false},
		{"missing positional block", "NonceQuery", []any{address, []any{"0x1"}}, false},
		{"extra argument", "NonceQuery", []any{address, "latest", []any{"0x1"}, "0x0"}, false},
		{"zero sequence", "NonceResult", "0x0", true},
		{"exhausted sequence", "NonceResult", "0xffffffffffffffff", true},
		{"padded sequence", "NonceResult", "0x00", false},
		{"numeric sequence", "NonceResult", 1, false},
	}
	for _, block := range []string{"latest", "pending", "earliest", "safe", "finalized", "0x10", "0x" + strings.Repeat("1", 64)} {
		tests = append(tests, testCase{block, "NonceQuery", []any{address, block, []any{"0x1", "0x2"}}, true})
	}
	for _, tc := range []struct {
		name  string
		keys  any
		valid bool
	}{
		{"legacy key", []any{"0x0"}, true},
		{"single key", []any{"0x1"}, true},
		{"multiple keys", []any{"0x1", "0x2"}, true},
		{"maximum key", []any{"0x" + strings.Repeat("f", 64)}, true},
		{"empty keys", []any{}, false},
		{"duplicate keys", []any{"0x1", "0x1"}, false},
		{"mixed zero key", []any{"0x0", "0x1"}, false},
		{"padded key", []any{"0x01"}, false},
		{"oversized key", []any{"0x1" + strings.Repeat("0", 64)}, false},
		{"numeric key", []any{1}, false},
		{"scalar key", "0x1", false},
		{"null keys", nil, false},
	} {
		tests = append(tests, testCase{tc.name, "NonceQuery", []any{address, "latest", tc.keys}, tc.valid})
	}
	for _, count := range []int{16, 17} {
		keys := make([]any, count)
		for i := range keys {
			keys[i] = fmt.Sprintf("0x%x", i+1)
		}
		tests = append(tests, testCase{fmt.Sprintf("%d keys", count), "NonceQuery", []any{address, "latest", keys}, count == 16})
	}
	for _, raw := range method["examples"].([]any) {
		example := raw.(object)
		var values []any
		for _, param := range example["params"].([]any) {
			values = append(values, param.(object)["value"])
		}
		tests = append(tests,
			testCase{example["name"].(string), "NonceQuery", values, true},
			testCase{example["name"].(string), "NonceResult", example["result"].(object)["value"], true},
		)
	}
	for _, expanded := range []bool{false, true} {
		for _, schemaName := range []string{"NonceQuery", "NonceResult"} {
			root := object{"components": object{"schemas": repo2object(generator.types)}, "$ref": "#/components/schemas/" + schemaName}
			if expanded {
				root, err = generator.expandSchema(generator.types[schemaName], generator.types)
				if err != nil {
					t.Fatal(err)
				}
			}
			compiler := jsonschema.NewCompiler()
			compiler.DefaultDraft(jsonschema.Draft7)
			if err := compiler.AddResource("nonce.json", root); err != nil {
				t.Fatal(err)
			}
			schema, err := compiler.Compile("nonce.json")
			if err != nil {
				t.Fatal(err)
			}
			for _, tc := range tests {
				if tc.schema != schemaName {
					continue
				}
				t.Run(fmt.Sprintf("expanded=%t/%s/%s", expanded, tc.schema, tc.name), func(t *testing.T) {
					data, err := json.Marshal(tc.value)
					if err != nil {
						t.Fatal(err)
					}
					var value any
					if err := json.Unmarshal(data, &value); err != nil {
						t.Fatal(err)
					}
					err = schema.Validate(value)
					if (err == nil) != tc.valid {
						t.Fatalf("valid=%v, validation error: %v", tc.valid, err)
					}
				})
			}
		}
	}
}
