package specgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFrameCallAPIs(t *testing.T) {
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
	for _, file := range []string{"../../../src/eth/execute.yaml", "../../../src/eth/fill.yaml", "../../../src/eth/sign.yaml", "../../../src/eth/submit.yaml"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddMethods(data); err != nil {
			t.Fatal(err)
		}
	}

	for _, method := range []string{"eth_call", "eth_estimateGas", "eth_createAccessList", "eth_fillTransaction", "eth_signTransaction", "eth_sendTransaction"} {
		generator.types[method] = generator.methods[method]["params"].([]any)[0].(object)["schema"].(object)
	}
	type testCase struct {
		name, schema string
		value        any
		valid        bool
	}
	var tests []testCase
	for _, variant := range []string{"defaults", "zero limits", "placeholder", "p256 placeholder", "arbitrary witness", "complete signature", "signed incomplete envelope", "signed incomplete frame", "empty signer", "arbitrary placeholder", "empty protocol signature", "missing execution limit", "missing state limit", "empty frames", "outer input", "wrong type"} {
		frame := object{"mode": "0x1", "executionGas": "0x10000", "stateGas": "0x10000"}
		request := object{"type": "0x6", "from": "0x1111111111111111111111111111111111111111", "frames": []any{frame}}
		placeholder := object{"scheme": "0x1", "signer": request["from"], "msg": "0x"}
		valid := true
		switch variant {
		case "zero limits":
			frame["executionGas"], frame["stateGas"] = "0x0", "0x0"
		case "placeholder", "p256 placeholder", "empty signer", "arbitrary placeholder", "empty protocol signature":
			request["signatures"] = []any{placeholder}
			if variant == "p256 placeholder" {
				placeholder["scheme"] = "0x2"
			}
			if variant == "empty signer" {
				placeholder["signer"] = "0x"
				valid = false
			}
			if variant == "arbitrary placeholder" {
				placeholder["scheme"], placeholder["signer"] = "0x0", "0x"
				valid = false
			}
			if variant == "empty protocol signature" {
				placeholder["signature"] = "0x"
				valid = false
			}
		case "arbitrary witness":
			request["signatures"] = []any{object{"scheme": "0x0", "signer": "0x", "msg": "0x", "signature": "0xabcd"}}
		case "complete signature", "signed incomplete envelope", "signed incomplete frame":
			// Structurally valid signature bytes; cryptographic validity requires a client.
			placeholder["signature"] = "0x00" + strings.Repeat("11", 64)
			request["signatures"] = []any{placeholder}
			request["chainId"], request["nonce"] = "0x1", "0x0"
			request["maxPriorityFeePerGas"], request["maxFeePerGas"], request["maxFeePerBlobGas"] = "0x0", "0x0", "0x0"
			request["blobVersionedHashes"] = []any{}
			frame["flags"], frame["target"], frame["value"], frame["data"] = "0x3", nil, "0x0", "0x"
			if variant == "signed incomplete envelope" {
				delete(request, "nonce")
				valid = false
			}
			if variant == "signed incomplete frame" {
				delete(frame, "data")
				valid = false
			}
		case "missing execution limit":
			delete(frame, "executionGas")
			valid = false
		case "missing state limit":
			delete(frame, "stateGas")
			valid = false
		case "empty frames":
			request["frames"] = []any{}
			valid = false
		case "outer input":
			request["input"] = "0x"
			valid = false
		case "wrong type":
			request["type"] = "0x2"
			valid = false
		}
		tests = append(tests, testCase{variant, "eth_call", request, valid})
		for _, method := range []string{"eth_estimateGas", "eth_createAccessList", "eth_fillTransaction", "eth_signTransaction", "eth_sendTransaction"} {
			tests = append(tests, testCase{variant, method, request, valid})
		}
		if variant == "complete signature" {
			data, err := json.Marshal(request)
			if err != nil {
				t.Fatal(err)
			}
			var unsigned object
			if err := json.Unmarshal(data, &unsigned); err != nil {
				t.Fatal(err)
			}
			delete(unsigned["signatures"].([]any)[0].(map[string]any), "signature")
			tests = append(tests, testCase{"complete envelope with placeholder", "eth_call", unsigned, true}, testCase{"placeholder is not signed", "Transaction8141Signed", unsigned, false})
		}
	}
	tests = append(tests, testCase{"frame type without frames", "eth_call", object{"type": "0x6"}, false})
	for _, method := range []string{"eth_call", "eth_estimateGas", "eth_createAccessList", "eth_fillTransaction", "eth_signTransaction", "eth_sendTransaction"} {
		tests = append(tests, testCase{"legacy request", method, object{"to": "0x1111111111111111111111111111111111111111", "input": "0x"}, true})
	}
	files, err = filepath.Glob("../../../tests/eth_call/*.io")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, line := range strings.Split(string(data), "\n") {
			if !strings.HasPrefix(line, ">> ") {
				continue
			}
			var request object
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, ">> ")), &request); err != nil {
				t.Fatal(err)
			}
			if request["method"] == "eth_call" {
				tests = append(tests, testCase{filepath.Base(file), "eth_call", request["params"].([]any)[0], true})
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
				if err := compiler.AddResource("call.json", root); err != nil {
					t.Fatal(err)
				}
				schema, err := compiler.Compile("call.json")
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
