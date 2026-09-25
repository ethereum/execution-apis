package specgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

func TestFrameReadAPIs(t *testing.T) {
	generator := New()
	files, err := filepath.Glob("../../../src/schemas/*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddSchemas(content); err != nil {
			t.Fatal(err)
		}
	}
	for _, file := range []string{"transaction.yaml", "block.yaml"} {
		content, err := os.ReadFile("../../../src/eth/" + file)
		if err != nil {
			t.Fatal(err)
		}
		if err := generator.AddMethods(content); err != nil {
			t.Fatal(err)
		}
	}
	mined := readFrameFixture(t, "frame-mined")
	pending := readFrameFixture(t, "frame-pending")
	block := readFrameFixture(t, "frame-block")
	type testCase struct {
		name, method string
		value        any
		valid        bool
	}
	var tests []testCase
	for _, method := range []string{"eth_getTransactionByHash", "eth_getTransactionByBlockHashAndIndex", "eth_getTransactionByBlockNumberAndIndex"} {
		tests = append(tests, testCase{"mined", method, mined, true}, testCase{"not found", method, nil, true})
		if method != "eth_getTransactionByBlockHashAndIndex" {
			tests = append(tests, testCase{"pending", method, pending, false})
		}
		for _, mutation := range []string{"placeholder", "empty cryptographic signature", "missing from", "missing hash", "malformed hash"} {
			data, err := json.Marshal(mined)
			if err != nil {
				t.Fatal(err)
			}
			var value object
			if err := json.Unmarshal(data, &value); err != nil {
				t.Fatal(err)
			}
			switch mutation {
			case "placeholder":
				value["signatures"] = []any{object{"scheme": "0x1", "signer": "0x", "msg": "0x"}}
			case "empty cryptographic signature":
				value["signatures"] = []any{object{"scheme": "0x1", "signer": "0x", "msg": "0x", "signature": "0x"}}
			case "missing from":
				delete(value, "from")
			case "missing hash":
				delete(value, "hash")
			case "malformed hash":
				value["hash"] = "0x1234"
			}
			tests = append(tests, testCase{mutation, method, value, false})
		}
		for _, entry := range []struct {
			name       string
			signatures []any
		}{
			{"no protocol signatures", []any{}},
			{"secp256k1", []any{object{"scheme": "0x1", "signer": "0x", "msg": "0x", "signature": "0x00" + strings.Repeat("1", 128)}}},
			{"p256", []any{object{"scheme": "0x2", "signer": "0x", "msg": "0x", "signature": "0x" + strings.Repeat("1", 256)}}},
		} {
			value := make(object, len(mined))
			for key, field := range mined {
				value[key] = field
			}
			value["signatures"] = entry.signatures
			tests = append(tests, testCase{entry.name, method, value, true})
		}
		// Reuse recorded responses to cover every existing transaction variant.
		for _, file := range []string{"get-legacy-tx.io", "get-access-list.io", "get-dynamic-fee.io", "get-blob-tx.io", "get-setcode-tx.io"} {
			content, err := os.ReadFile("../../../tests/eth_getTransactionByHash/" + file)
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range strings.Split(string(content), "\n") {
				if !strings.HasPrefix(line, "<< ") {
					continue
				}
				var response object
				if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "<< ")), &response); err != nil {
					t.Fatal(err)
				}
				tests = append(tests, testCase{file, method, response["result"], true})
			}
		}
	}
	for _, method := range []string{"eth_getBlockByHash", "eth_getBlockByNumber"} {
		tests = append(tests, testCase{"mixed full transactions", method, block, true}, testCase{"not found", method, nil, true})
		for _, variant := range []string{"hashes", "placeholder", "mixed hash and object"} {
			data, err := json.Marshal(block)
			if err != nil {
				t.Fatal(err)
			}
			var value object
			if err := json.Unmarshal(data, &value); err != nil {
				t.Fatal(err)
			}
			txs := value["transactions"].([]any)
			switch variant {
			case "hashes":
				for i, tx := range txs {
					txs[i] = tx.(object)["hash"]
				}
			case "placeholder":
				txs[1].(object)["signatures"] = []any{object{"scheme": "0x2", "signer": "0x", "msg": "0x"}}
			case "mixed hash and object":
				txs[0] = txs[0].(object)["hash"]
			}
			tests = append(tests, testCase{variant, method, value, variant == "hashes"})
		}
	}
	for _, expanded := range []bool{false, true} {
		mode := "referenced/"
		if expanded {
			mode = "expanded/"
		}
		for _, tc := range tests {
			t.Run(mode+tc.method+"/"+tc.name, func(t *testing.T) {
				method := generator.methods[tc.method]
				schema := method["result"].(object)["schema"].(object)
				var root object
				if expanded {
					var err error
					root, err = generator.expandSchema(schema, generator.types)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					root = object{"components": object{"schemas": repo2object(generator.types)}, "allOf": []any{schema}}
				}
				compiler := jsonschema.NewCompiler()
				compiler.DefaultDraft(jsonschema.Draft7)
				if err := compiler.AddResource("frame-read.json", root); err != nil {
					t.Fatal(err)
				}
				compiled, err := compiler.Compile("frame-read.json")
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
				err = compiled.Validate(value)
				if (err == nil) != tc.valid {
					t.Fatalf("valid=%v, validation error: %v", tc.valid, err)
				}
			})
		}
	}
}

func readFrameFixture(t *testing.T, name string) object {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var value object
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
