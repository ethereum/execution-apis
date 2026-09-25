package specgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFrameReceiptAPIs(t *testing.T) {
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
	for _, file := range []string{"../../../src/eth/block.yaml", "../../../src/eth/transaction.yaml", "../../../src/eth/filter.yaml"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddMethods(data); err != nil {
			t.Fatal(err)
		}
	}

	legacy := generator.methods["eth_getTransactionReceipt"]["examples"].([]any)[0].(object)["result"].(object)["value"].(object)
	type testCase struct {
		name, schema string
		value        any
		valid        bool
	}
	var tests []testCase
	for _, scenario := range []string{"success", "independent failure", "atomic rollback", "blob"} {
		data, err := json.Marshal(legacy)
		if err != nil {
			t.Fatal(err)
		}
		var receipt object
		if err := json.Unmarshal(data, &receipt); err != nil {
			t.Fatal(err)
		}
		receipt["type"] = "0x6"
		receipt["payer"] = "0x3333333333333333333333333333333333333333"
		receipt["to"], receipt["contractAddress"] = nil, nil
		delete(receipt, "blobGasUsed")
		delete(receipt, "blobGasPrice")
		if scenario == "blob" {
			receipt["blobGasUsed"], receipt["blobGasPrice"] = "0x20000", "0x3"
		}
		log := object{"transactionHash": receipt["transactionHash"], "blockHash": receipt["blockHash"], "blockNumber": receipt["blockNumber"], "transactionIndex": "0x1", "logIndex": "0x4", "address": receipt["from"], "topics": []any{}, "data": "0x", "removed": false}
		// Schema examples only; gas and rollback behavior require client execution tests.
		frames := []any{
			object{"status": "0x1", "gasUsed": "0x100", "executionGasUsed": "0x100", "stateGasUsed": "0x0", "logs": []any{}},
			object{"status": "0x1", "gasUsed": "0x300", "executionGasUsed": "0x100", "stateGasUsed": "0x200", "logs": []any{log}},
		}
		if scenario == "independent failure" {
			receipt["status"] = "0x0"
			frames = append(frames, object{"status": "0x0", "gasUsed": "0x100", "executionGasUsed": "0x100", "stateGasUsed": "0x0", "logs": []any{}})
		}
		if scenario == "atomic rollback" {
			receipt["status"] = "0x0"
			frames[1] = object{"status": "0x1", "gasUsed": "0x100", "executionGasUsed": "0x100", "stateGasUsed": "0x0", "logs": []any{}}
			frames = append(frames,
				object{"status": "0x0", "gasUsed": "0x100", "executionGasUsed": "0x100", "stateGasUsed": "0x0", "logs": []any{}},
				object{"status": "0x2", "gasUsed": "0x0", "executionGasUsed": "0x0", "stateGasUsed": "0x0", "logs": []any{}},
				object{"status": "0x1", "gasUsed": "0x300", "executionGasUsed": "0x100", "stateGasUsed": "0x200", "logs": []any{log}})
		}
		receipt["frameReceipts"], receipt["logs"] = frames, []any{log}
		for _, method := range []string{"eth_getTransactionReceipt", "eth_getBlockReceipts"} {
			var value any = receipt
			if method == "eth_getBlockReceipts" {
				value = []any{legacy, receipt}
			}
			generator.types[method] = generator.methods[method]["result"].(object)["schema"].(object)
			tests = append(tests, testCase{scenario, method, value, true})
		}
		for _, method := range []string{"eth_getLogs", "eth_getFilterLogs", "eth_getFilterChanges"} {
			generator.types[method] = generator.methods[method]["result"].(object)["schema"].(object)
			tests = append(tests, testCase{scenario, method, []any{log}, true})
		}
		for _, missing := range []string{"status", "gasUsed", "executionGasUsed", "stateGasUsed", "logs"} {
			frame := object{"status": "0x1", "gasUsed": "0x300", "executionGasUsed": "0x100", "stateGasUsed": "0x200", "logs": []any{}}
			delete(frame, missing)
			tests = append(tests, testCase{"missing " + missing, "FrameReceipt", frame, false})
		}
		badLog := object{"transactionHash": receipt["transactionHash"], "frameIndex": "0x1"}
		tests = append(tests, testCase{"no frame index extension", "Log", badLog, false})
	}
	for _, method := range []string{"eth_getTransactionReceipt", "eth_getBlockReceipts"} {
		tests = append(tests, testCase{"not found", method, nil, true})
		files, err := filepath.Glob("../../../tests/" + method + "/*.io")
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
				if result, ok := response["result"]; ok {
					tests = append(tests, testCase{filepath.Base(file), method, result, true})
				}
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
				if err := compiler.AddResource("receipts.json", root); err != nil {
					t.Fatal(err)
				}
				schema, err := compiler.Compile("receipts.json")
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
