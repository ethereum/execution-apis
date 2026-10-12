package specgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v5"
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
	validator, err := jsonschema.CompileString("account-override.json", string(encoded))
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
		{"state only", `{"state":{}}`, true},
		{"stateDiff only", `{"stateDiff":{}}`, true},
		{"both storage overrides", `{"state":{},"stateDiff":{}}`, false},
		{"numeric balance", `{"balance":1}`, false},
		{"numeric nonce", `{"nonce":1}`, false},
		{"odd length code", `{"code":"0x1"}`, false},
		{"state array", `{"state":[]}`, false},
		{"stateDiff array", `{"stateDiff":[]}`, false},
		{"short storage key", `{"state":{"0x1":"0x1"}}`, false},
		{"null", `null`, false},
		{"array", `[]`, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var value any
			if err := json.Unmarshal([]byte(tt.input), &value); err != nil {
				t.Fatal(err)
			}
			err := validator.Validate(value)
			if (err == nil) != tt.valid {
				t.Errorf("valid=%t, want %t: %v", err == nil, tt.valid, err)
			}
		})
	}
}
