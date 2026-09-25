package specgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFrameFillAPIs(t *testing.T) {
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
	for _, file := range []string{"../../../src/eth/fill.yaml"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddMethods(data); err != nil {
			t.Fatal(err)
		}
	}

	method := generator.methods["eth_fillTransaction"]
	fixture := readFrameFixture(t, "frame-fill")
	request := fixture["request"]
	result := fixture["result"]
	generator.types["FillRequest"] = method["params"].([]any)[0].(object)["schema"].(object)
	generator.types["FillResult"] = method["result"].(object)["schema"].(object)
	type testCase struct {
		name, schema string
		value        any
		valid        bool
	}
	tests := []testCase{{"partial request", "FillRequest", request, true}, {"filled envelope", "FillResult", result, true}}
	for _, variant := range []string{"empty signatures", "complete signature", "arbitrary witness", "missing chain", "missing nonce", "missing fee", "missing execution gas", "missing state gas", "scheme-only placeholder", "malformed signature", "missing tx"} {
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatal(err)
		}
		var response object
		if err := json.Unmarshal(data, &response); err != nil {
			t.Fatal(err)
		}
		tx := response["tx"].(map[string]any)
		frame := tx["frames"].([]any)[0].(map[string]any)
		signature := tx["signatures"].([]any)[0].(map[string]any)
		valid := false
		switch variant {
		case "empty signatures":
			tx["signatures"] = []any{}
			valid = true
		case "complete signature":
			signature["signature"] = "0x00" + strings.Repeat("11", 64)
			valid = true
		case "arbitrary witness":
			tx["signatures"] = []any{object{"scheme": "0x0", "signer": "0x", "msg": "0x", "signature": "0xabcd"}}
			valid = true
		case "missing chain":
			delete(tx, "chainId")
		case "missing nonce":
			delete(tx, "nonce")
		case "missing fee":
			delete(tx, "maxFeePerGas")
		case "missing execution gas":
			delete(frame, "executionGas")
		case "missing state gas":
			delete(frame, "stateGas")
		case "scheme-only placeholder":
			delete(signature, "signer")
			delete(signature, "msg")
			valid = true
		case "malformed signature":
			signature["signature"] = "0xabcd"
		case "missing tx":
			delete(response, "tx")
		}
		tests = append(tests, testCase{variant, "FillResult", response, valid})
	}
	// Project recorded transaction envelopes to unsigned fill results.
	files, err = filepath.Glob("../../../tests/eth_getTransactionByHash/*.io")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.HasPrefix(line, "<< ") {
				continue
			}
			var response object
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "<< ")), &response); err != nil {
				t.Fatal(err)
			}
			if value, ok := response["result"].(map[string]any); ok {
				for _, field := range []string{"blockHash", "blockNumber", "blockTimestamp", "transactionIndex", "hash", "from", "v", "r", "s", "yParity"} {
					delete(value, field)
				}
				tests = append(tests, testCase{filepath.Base(file), "FillResult", object{"tx": value}, true})
			}
		}
	}
	for _, expanded := range []bool{false, true} {
		mode := "referenced/"
		if expanded {
			mode = "expanded/"
		}
		for _, tc := range tests {
			t.Run(mode+tc.schema+"/"+tc.name, func(t *testing.T) {
				root := object{"components": object{"schemas": repo2object(generator.types)}, "$ref": "#/components/schemas/" + tc.schema}
				if expanded {
					var err error
					root, err = generator.expandSchema(generator.types[tc.schema], generator.types)
					if err != nil {
						t.Fatal(err)
					}
				}
				compiler := jsonschema.NewCompiler()
				compiler.DefaultDraft(jsonschema.Draft7)
				if err := compiler.AddResource("fill.json", root); err != nil {
					t.Fatal(err)
				}
				schema, err := compiler.Compile("fill.json")
				if err != nil {
					t.Fatal(err)
				}
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
