package specgen

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

// TestAccountOverride validates the component directly because the simulation
// payload schema does not currently validate blockStateCalls items (issue #924).
func TestAccountOverride(t *testing.T) {
	gen := New()
	files, err := filepath.Glob("../../../src/schemas/*.yaml")
	if err != nil || len(files) == 0 {
		t.Fatalf("schema files: %v", err)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := gen.AddSchemas(content); err != nil {
			t.Fatal(err)
		}
	}
	schema, err := gen.expandSchema(gen.types["AccountOverride"], gen.types)
	if err != nil {
		t.Fatal(err)
	}
	// Match the dialect used by speccheck.
	schema["$schema"] = "https://json-schema.org/draft/2019-09/schema"
	encoded, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := jsonschema.UnmarshalJSON(bytes.NewReader(encoded))
	if err != nil {
		t.Fatal(err)
	}
	compiler := jsonschema.NewCompiler()
	if err := compiler.AddResource("account-override.json", doc); err != nil {
		t.Fatal(err)
	}
	validator, err := compiler.Compile("account-override.json")
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		{"empty", `{}`, true},
		{"balance only", `{"balance":"0x1"}`, true},
		{"code only", `{"code":"0x6000"}`, true},
		{"nonce only", `{"nonce":"0x1"}`, true},
		{"move precompile only", `{"movePrecompileToAddress":"0x0000000000000000000000000000000000001111"}`, true},
		{"combined fields", `{"balance":"0x0","nonce":"0x0","code":"0x"}`, true},
		{"state only", `{"state":{}}`, true},
		{"stateDiff only", `{"stateDiff":{}}`, true},
		{"state with code", `{"state":{},"code":"0x6000"}`, true},
		{"stateDiff with balance", `{"stateDiff":{},"balance":"0x1"}`, true},
		{"both storage overrides", `{"state":{},"stateDiff":{}}`, false},
		{"numeric balance", `{"balance":1}`, false},
		{"balance with leading zero", `{"balance":"0x01"}`, false},
		{"numeric nonce", `{"nonce":1}`, false},
		{"nonce overflow", `{"nonce":"0x10000000000000000"}`, false},
		{"numeric code", `{"code":1}`, false},
		{"nonhex code", `{"code":"0xgg"}`, false},
		{"short move address", `{"movePrecompileToAddress":"0x1"}`, false},
		{"state array", `{"state":[]}`, false},
		{"stateDiff array", `{"stateDiff":[]}`, false},
		{"null state", `{"state":null}`, false},
		{"null stateDiff", `{"stateDiff":null}`, false},
		{"short storage key", `{"state":{"0x1":"0x1"}}`, false},
		{"short storage value", `{"state":{"0x0000000000000000000000000000000000000000000000000000000000000000":"0x1"}}`, false},
		{"short stateDiff value", `{"stateDiff":{"0x0000000000000000000000000000000000000000000000000000000000000000":"0x1"}}`, false},
		{"null", `null`, false},
		{"array", `[]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := jsonschema.UnmarshalJSON(bytes.NewReader([]byte(tt.input)))
			if err != nil {
				t.Fatal(err)
			}
			err = validator.Validate(value)
			if (err == nil) != tt.valid {
				t.Errorf("valid=%t, want %t: %v", err == nil, tt.valid, err)
			}
		})
	}
}
