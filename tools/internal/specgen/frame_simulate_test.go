package specgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFrameSimulateAPIs(t *testing.T) {
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
	for _, file := range []string{"../../../src/eth/execute.yaml", "../../../src/eth/transaction.yaml"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddMethods(data); err != nil {
			t.Fatal(err)
		}
	}

	type testCase struct {
		name, schema string
		value        any
		valid        bool
	}
	var tests []testCase
	for _, variant := range []string{"optional limits", "explicit limits", "placeholder", "witness", "blob hashes", "scalar blob hash", "invalid mode", "invalid limit", "invalid signature", "invalid override", "invalid block entry"} {
		frame := object{"mode": "0x2"}
		tx := object{"type": "0x6", "from": "0x1111111111111111111111111111111111111111", "frames": []any{frame}}
		block := object{"calls": []any{object{"to": tx["from"]}, tx}}
		valid := true
		switch variant {
		case "explicit limits":
			frame["executionGas"], frame["stateGas"] = "0x10000", "0x0"
		case "placeholder":
			tx["signatures"] = []any{object{"scheme": "0x1", "signer": tx["from"], "msg": "0x"}}
		case "witness":
			tx["signatures"] = []any{object{"scheme": "0x0", "signer": "0x", "msg": "0x", "signature": "0xabcd"}}
		case "blob hashes":
			tx["blobVersionedHashes"] = []any{"0x0100000000000000000000000000000000000000000000000000000000000000"}
		case "scalar blob hash":
			tx["blobVersionedHashes"] = "0x0100000000000000000000000000000000000000000000000000000000000000"
			valid = false
		case "invalid mode":
			frame["mode"] = "0x3"
			valid = false
		case "invalid limit":
			frame["executionGas"] = -1
			valid = false
		case "invalid signature":
			tx["signatures"] = []any{object{"scheme": "0x1", "signer": tx["from"], "msg": "0x", "signature": "0x"}}
			valid = false
		case "invalid override":
			block["stateOverrides"] = object{"invalid address": object{}}
			valid = false
		}
		for _, validation := range []bool{false, true} {
			for _, full := range []bool{false, true} {
				var entry any = block
				if variant == "invalid block entry" {
					entry = "0x1"
					valid = false
				}
				payload := object{"blockStateCalls": []any{entry}, "validation": validation, "returnFullTransactions": full}
				tests = append(tests, testCase{variant, "EthSimulatePayload", payload, valid})
			}
		}
	}
	for _, variant := range []string{"success", "failure", "atomic rollback", "missing frame gas", "invalid return bytes", "malformed error"} {
		frame := object{"status": "0x1", "gasUsed": "0x300", "executionGasUsed": "0x100", "stateGasUsed": "0x200", "logs": []any{}, "returnData": "0xabcd"}
		result := object{"status": "0x1", "gasUsed": "0x4000", "logs": []any{}, "returnData": "0xabcd", "payer": "0x3333333333333333333333333333333333333333", "frameResults": []any{frame}}
		valid := true
		switch variant {
		case "failure", "atomic rollback":
			result["status"], result["returnData"] = "0x0", "0xdead"
			result["error"] = object{"code": 3, "message": "execution reverted"}
			frame["status"], frame["gasUsed"], frame["stateGasUsed"], frame["returnData"], frame["error"] = "0x0", "0x100", "0x0", "0xdead", result["error"]
			if variant == "atomic rollback" {
				result["frameResults"] = []any{
					object{"status": "0x1", "gasUsed": "0x100", "executionGasUsed": "0x100", "stateGasUsed": "0x0", "logs": []any{}, "returnData": "0xabcd"},
					frame,
					object{"status": "0x2", "gasUsed": "0x0", "executionGasUsed": "0x0", "stateGasUsed": "0x0", "logs": []any{}, "returnData": "0x"},
				}
			}
		case "missing frame gas":
			delete(frame, "stateGasUsed")
			valid = false
		case "invalid return bytes":
			frame["returnData"] = "not hex"
			valid = false
		case "malformed error":
			frame["error"] = object{"code": 123, "message": "unknown"}
			valid = false
		}
		tests = append(tests, testCase{variant, "CallResults", []any{result}, valid})
	}
	var fullTx object
	for _, raw := range generator.methods["eth_getTransactionByHash"]["examples"].([]any) {
		example := raw.(object)
		if example["name"] == "Frame transaction schema example" {
			fullTx = example["result"].(object)["value"].(object)
		}
	}
	if fullTx == nil {
		t.Fatal("missing full frame transaction example")
	}
	// Existing examples still validate after correcting the block entry schema.
	for _, raw := range generator.methods["eth_simulateV1"]["examples"].([]any) {
		example := raw.(object)
		tests = append(tests, testCase{example["name"].(string), "EthSimulatePayload", example["params"].([]any)[0].(object)["value"], true})
		tests = append(tests, testCase{example["name"].(string), "EthSimulateResult", example["result"].(object)["value"], true})
		for _, full := range []bool{false, true} {
			data, err := json.Marshal(example["result"].(object)["value"])
			if err != nil {
				t.Fatal(err)
			}
			var blocks []any
			if err := json.Unmarshal(data, &blocks); err != nil {
				t.Fatal(err)
			}
			block := blocks[0].(map[string]any)
			var tx any = fullTx["hash"]
			if full {
				tx = fullTx
			}
			block["transactions"] = []any{tx}
			block["calls"] = []any{object{"status": "0x1", "gasUsed": "0x4000", "logs": []any{}, "returnData": "0xabcd", "payer": fullTx["from"], "frameResults": []any{object{"status": "0x1", "gasUsed": "0x100", "executionGasUsed": "0x100", "stateGasUsed": "0x0", "logs": []any{}, "returnData": "0xabcd"}}}}
			tests = append(tests, testCase{"frame block result", "EthSimulateResult", blocks, true})
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
				if err := compiler.AddResource("simulate.json", root); err != nil {
					t.Fatal(err)
				}
				schema, err := compiler.Compile("simulate.json")
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
