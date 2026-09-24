package specgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFramePendingAPIs(t *testing.T) {
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
	for _, file := range []string{"../../../src/txpool/pool.yaml", "../../../src/eth/transaction.yaml", "../../../src/eth/subscribe.yaml"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddMethods(data); err != nil {
			t.Fatal(err)
		}
	}
	var pending object
	for _, raw := range generator.methods["eth_getTransactionByHash"]["examples"].([]any) {
		example := raw.(object)
		if example["name"] == "Pending frame transaction schema example" {
			pending = example["result"].(object)["value"].(object)
		}
	}
	if pending == nil {
		t.Fatal("missing pending frame example")
	}
	// These schema examples model separate payment and execution approval frames.
	sponsor := "0x3333333333333333333333333333333333333333"
	pending["frames"] = []any{
		object{"mode": "0x1", "flags": "0x1", "target": sponsor, "executionGas": "0x10000", "stateGas": "0x0", "value": "0x0", "data": "0x"},
		object{"mode": "0x1", "flags": "0x2", "target": nil, "executionGas": "0x10000", "stateGas": "0x0", "value": "0x0", "data": "0x"},
	}
	type testCase struct {
		name, schema string
		value        any
		valid        bool
	}
	var tests []testCase
	for _, variant := range []string{"pending", "queued", "omitted block metadata", "placeholder", "mined metadata", "missing hash", "malformed hash"} {
		data, err := json.Marshal(pending)
		if err != nil {
			t.Fatal(err)
		}
		var tx object
		if err := json.Unmarshal(data, &tx); err != nil {
			t.Fatal(err)
		}
		valid := true
		bucket, nonce := "pending", "0"
		switch variant {
		case "queued":
			bucket, nonce = "queued", "2"
			tx["nonce"] = "0x2"
		case "omitted block metadata":
			for _, key := range []string{"blockHash", "blockNumber", "blockTimestamp", "transactionIndex"} {
				delete(tx, key)
			}
		case "placeholder":
			tx["signatures"] = []any{object{"scheme": "0x1", "signer": "0x", "msg": "0x"}}
			valid = false
		case "mined metadata":
			tx["blockNumber"] = "0x1"
			valid = false
		case "missing hash":
			delete(tx, "hash")
			valid = false
		case "malformed hash":
			tx["hash"] = "0x1234"
			valid = false
		}
		bySender := object{"pending": object{}, "queued": object{}}
		bySender[bucket] = object{nonce: tx}
		pool := object{"pending": object{}, "queued": object{}}
		pool[bucket] = object{tx["from"].(string): object{nonce: tx}}
		tests = append(tests, testCase{variant, "TxpoolContent", pool, valid}, testCase{variant, "TxpoolContentFromResult", bySender, valid})
		if variant != "queued" {
			notification := object{"jsonrpc": "2.0", "method": "eth_subscription", "params": object{"subscription": "0x1", "result": tx}}
			tests = append(tests, testCase{variant, "NewPendingTransactionsNotification", notification, valid})
		}
	}
	for _, result := range []struct {
		name  string
		value any
		valid bool
	}{
		{"hash only", pending["hash"], true}, {"short hash", "0x1234", false}, {"null result", nil, false},
	} {
		notification := object{"jsonrpc": "2.0", "method": "eth_subscription", "params": object{"subscription": "0x1", "result": result.value}}
		tests = append(tests, testCase{result.name, "NewPendingTransactionsNotification", notification, result.valid})
	}
	for _, variant := range []string{"missing subscription", "wrong method", "response id", "wrong version"} {
		notification := object{"jsonrpc": "2.0", "method": "eth_subscription", "params": object{"subscription": "0x1", "result": pending["hash"]}}
		switch variant {
		case "missing subscription":
			delete(notification["params"].(object), "subscription")
		case "wrong method":
			notification["method"] = "eth_subscribe"
		case "response id":
			notification["id"] = 1
		case "wrong version":
			notification["jsonrpc"] = "1.0"
		}
		tests = append(tests, testCase{variant, "NewPendingTransactionsNotification", notification, false})
	}
	for _, file := range []struct{ path, schema string }{
		{"txpool_content/get-content.io", "TxpoolContent"},
		{"txpool_contentFrom/get-content-from-address.io", "TxpoolContentFromResult"},
	} {
		data, err := os.ReadFile("../../../tests/" + file.path)
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
			tests = append(tests, testCase{"existing fixture", file.schema, response["result"], true})
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
				if err := compiler.AddResource("pending.json", root); err != nil {
					t.Fatal(err)
				}
				schema, err := compiler.Compile("pending.json")
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
