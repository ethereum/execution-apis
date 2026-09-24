package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/execution-apis/tools/internal/specgen"
	schema5 "github.com/santhosh-tekuri/jsonschema/v5"
	schema6 "github.com/santhosh-tekuri/jsonschema/v6"
)

type traceObject = map[string]any

// Build from the real YAML, so tests also run on a clean checkout without make.
func traceDocuments(t *testing.T) (traceObject, traceObject, map[string]*methodSchema) {
	t.Helper()
	g := specgen.New()
	paths, err := filepath.Glob("../../../src/schemas/*.yaml")
	if err != nil || len(paths) == 0 {
		t.Fatalf("schema files: %v", err)
	}
	for _, path := range paths {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := g.AddSchemas(data); err != nil {
			t.Fatal(err)
		}
	}
	data, err := os.ReadFile("../../../src/trace/methods.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := g.AddMethods(data); err != nil {
		t.Fatal(err)
	}
	read := func() (traceObject, []byte) {
		t.Helper()
		data, err := g.JSON()
		if err != nil {
			t.Fatal(err)
		}
		var doc traceObject
		if err := json.Unmarshal(data, &doc); err != nil {
			t.Fatal(err)
		}
		return doc, data
	}
	refs, _ := read()
	if err := g.Dereference(); err != nil {
		t.Fatal(err)
	}
	expanded, data := read()
	path := filepath.Join(t.TempDir(), "openrpc.json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	parsed, err := parseSpec(path)
	if err != nil {
		t.Fatal(err)
	}
	return refs, expanded, parsed
}

func traceJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func traceCopy(t *testing.T, value traceObject) traceObject {
	t.Helper()
	var copy traceObject
	if err := json.Unmarshal(traceJSON(t, value), &copy); err != nil {
		t.Fatal(err)
	}
	return copy
}

func TestTraceContracts(t *testing.T) {
	refs, expanded, parsed := traceDocuments(t)
	address := "0x" + strings.Repeat("1", 40)
	hash := "0x" + strings.Repeat("2", 64)
	call := traceObject{"type": "call", "action": traceObject{"callType": "call", "from": address, "to": address, "gas": "0x100", "input": "0x", "value": "0x0"}, "subtraces": 0, "traceAddress": []any{}, "result": traceObject{"gasUsed": "0x0", "output": "0x"}}
	create := traceCopy(t, call)
	create["type"] = "create"
	create["action"] = traceObject{"from": address, "gas": "0x100", "init": "0x00", "value": "0x0", "creationMethod": "create"}
	create["result"] = traceObject{"gasUsed": "0x0", "address": address, "code": "0x"}
	suicide := traceObject{"type": "suicide", "action": traceObject{"address": address, "refundAddress": address, "balance": "0x0"}, "subtraces": 0, "traceAddress": []any{}}
	reward := traceObject{"type": "reward", "action": traceObject{"author": address, "rewardType": "block", "value": "0x1"}, "subtraces": 0, "traceAddress": []any{}}
	local := func(frame traceObject) traceObject {
		x := traceCopy(t, frame)
		x["blockHash"] = hash
		x["blockNumber"] = 1
		x["transactionHash"] = hash
		x["transactionPosition"] = 0
		return x
	}
	historicalReward := local(reward)
	historicalReward["transactionHash"] = nil
	historicalReward["transactionPosition"] = nil
	envelope := func(frame any) traceObject {
		return traceObject{"output": "0x", "trace": []any{frame}, "stateDiff": nil, "vmTrace": nil}
	}
	type probe struct {
		name, method string
		param        int
		value        any
		valid        bool
	}
	var probes []probe
	add := func(name, method string, param int, value any, valid bool) {
		probes = append(probes, probe{name, method, param, value, valid})
	}
	for _, frame := range []traceObject{call, create, suicide} {
		kind := frame["type"].(string)
		for _, method := range []string{"trace_block", "trace_filter", "trace_transaction"} {
			add(kind+" mined", method, -1, []any{local(frame)}, true)
			add(kind+" unlocalized", method, -1, []any{frame}, false)
			for _, field := range []string{"blockHash", "blockNumber", "transactionHash", "transactionPosition"} {
				bad := local(frame)
				bad[field] = nil
				add(kind+" null "+field, method, -1, []any{bad}, false)
				bad = local(frame)
				delete(bad, field)
				add(kind+" missing "+field, method, -1, []any{bad}, false)
			}
		}
		add(kind+" lookup", "trace_get", -1, local(frame), true)
		for _, method := range []string{"trace_call", "trace_rawTransaction"} {
			add(kind+" simulation", method, -1, envelope(frame), true)
			for _, field := range []string{"blockHash", "blockNumber", "transactionHash", "transactionPosition"} {
				bad := traceCopy(t, frame)
				bad[field] = nil
				add(kind+" simulation has "+field, method, -1, envelope(bad), false)
			}
			add(kind+" simulation localized", method, -1, envelope(local(frame)), false)
		}
	}
	for _, method := range []string{"trace_block", "trace_filter"} {
		add("protocol reward", method, -1, []any{historicalReward}, true)
		add("reward with transaction", method, -1, []any{local(reward)}, false)
		for _, field := range []string{"blockHash", "blockNumber"} {
			bad := traceCopy(t, historicalReward)
			bad[field] = nil
			add("reward null "+field, method, -1, []any{bad}, false)
		}
		for _, field := range []string{"transactionHash", "transactionPosition"} {
			bad := traceCopy(t, historicalReward)
			bad[field] = local(reward)[field]
			add("reward non-null "+field, method, -1, []any{bad}, false)
		}
	}
	add("reward excluded", "trace_transaction", -1, []any{historicalReward}, false)
	add("reward excluded", "trace_get", -1, historicalReward, false)
	for _, method := range []string{"trace_call", "trace_rawTransaction"} {
		add("reward excluded", method, -1, envelope(reward), false)
	}
	for _, frame := range []traceObject{call, create} {
		kind := frame["type"].(string)
		reverted := traceCopy(t, frame)
		reverted["error"] = "Reverted"
		reverted["result"] = traceObject{"gasUsed": "0x1", "output": "0xdead"}
		add("revert bytes "+kind, "trace_call", -1, envelope(reverted), true)
		reverted = traceCopy(t, reverted)
		reverted["result"] = nil
		add("revert null result "+kind, "trace_call", -1, envelope(reverted), false)
		reverted = traceCopy(t, reverted)
		delete(reverted, "result")
		add("revert missing result "+kind, "trace_call", -1, envelope(reverted), false)
		failed := traceCopy(t, frame)
		failed["result"] = nil
		add("null result without error "+kind, "trace_call", -1, envelope(failed), false)
		failed = traceCopy(t, failed)
		failed["error"] = "Out of gas"
		add("halt null result "+kind, "trace_call", -1, envelope(failed), true)
		failed = traceCopy(t, failed)
		delete(failed, "result")
		add("halt missing result "+kind, "trace_call", -1, envelope(failed), true)
		failed = traceCopy(t, failed)
		failed["result"] = traceObject{"gasUsed": "0x1", "output": "0x"}
		add("halt with result "+kind, "trace_call", -1, envelope(failed), false)
		failed = traceCopy(t, failed)
		failed["error"] = "vendor detail"
		delete(failed, "result")
		add("extension error label "+kind, "trace_call", -1, envelope(failed), true)
	}
	failedCreate := traceCopy(t, create)
	failedCreate["error"] = "Out of gas"
	add("failed create successful result", "trace_call", -1, envelope(failedCreate), false)
	failedCreate = traceCopy(t, failedCreate)
	failedCreate["error"] = "Reverted"
	add("reverted create successful result", "trace_call", -1, envelope(failedCreate), false)
	failedCreate = traceCopy(t, failedCreate)
	failedCreate["result"] = traceObject{"gasUsed": "0x1", "output": "0xdead", "address": address}
	add("reverted create with address", "trace_call", -1, envelope(failedCreate), false)
	failedCreate = traceCopy(t, failedCreate)
	failedCreate["result"] = traceObject{"gasUsed": "0x1", "output": "0xdead", "address": address, "code": "0x"}
	add("failed create mixed result", "trace_call", -1, envelope(failedCreate), false)
	collision := traceCopy(t, create)
	collision["error"] = "Contract address collision"
	delete(collision, "result")
	add("create collision", "trace_call", -1, envelope(collision), true)
	for _, method := range []any{nil, "create3"} {
		bad := traceCopy(t, create)
		action := bad["action"].(traceObject)
		if method == nil {
			delete(action, "creationMethod")
		} else {
			action["creationMethod"] = method
		}
		add(fmt.Sprintf("creation method %v", method), "trace_call", -1, envelope(bad), false)
	}
	create2 := traceCopy(t, create)
	create2["action"].(traceObject)["creationMethod"] = "create2"
	add("create2", "trace_call", -1, envelope(create2), true)
	for _, value := range []any{traceObject{}, traceObject{"gasPrice": "0x0"}, traceObject{"maxFeePerGas": "0x1", "maxPriorityFeePerGas": "0x0"}} {
		add("valid fees", "trace_call", 0, value, true)
	}
	add("future call field", "trace_call", 0, traceObject{"futureField": traceObject{"x": 1}}, true)
	add("standard call fields", "trace_call", 0, traceObject{"chainId": "0x1", "maxFeePerBlobGas": "0x1", "blobVersionedHashes": []any{hash}, "authorizationList": []any{}}, true)
	add("invalid chain id", "trace_call", 0, traceObject{"chainId": 1}, false)
	add("gas above uint64", "trace_call", 0, traceObject{"gas": "0x10000000000000000"}, false)
	add("uint64 gas", "trace_call", 0, traceObject{"gas": "0xffffffffffffffff"}, true)
	add("block hash", "trace_call", 2, hash, true)
	add("block hash", "trace_callMany", 1, hash, true)
	for _, param := range []struct {
		method   string
		position int
	}{{"trace_call", 3}, {"trace_callMany", 2}} {
		add("null state overrides", param.method, param.position, nil, true)
		add("state overrides", param.method, param.position, traceObject{address: traceObject{"balance": "0x1", "stateDiff": traceObject{}}}, true)
		add("malformed state overrides", param.method, param.position, traceObject{"bad": traceObject{}}, false)
		add("null block overrides", param.method, param.position+1, nil, true)
		add("block overrides", param.method, param.position+1, traceObject{"baseFeePerGas": "0x0"}, true)
	}
	add("invalid blob hash", "trace_call", 0, traceObject{"blobVersionedHashes": []any{"0x01"}}, false)
	for _, value := range []any{nil, []any{}, []any{address}} {
		add("unrestricted or address filter", "trace_filter", 0, traceObject{"fromAddress": value, "toAddress": value}, true)
	}
	add("scalar address filter", "trace_filter", 0, traceObject{"fromAddress": address}, false)
	add("negative count", "trace_filter", 0, traceObject{"count": -1}, false)
	add("uint64 page", "trace_filter", 0, traceObject{"after": uint64(1) << 63, "count": 1}, true)
	for _, tag := range []string{"earliest", "latest", "safe", "finalized"} {
		add("range tag "+tag, "trace_filter", 0, traceObject{"fromBlock": tag, "toBlock": tag}, true)
	}
	add("pending range", "trace_filter", 0, traceObject{"fromBlock": "pending"}, false)
	for _, method := range []string{"trace_block", "trace_replayBlockTransactions"} {
		add("latest block", method, 0, "latest", true)
		add("pending block", method, 0, "pending", false)
	}
	add("hex path", "trace_get", 1, []any{"0x0", "0xa"}, true)
	add("integer path", "trace_get", 1, []any{0, 10}, false)
	for _, field := range []string{"maxFeePerGas", "maxPriorityFeePerGas"} {
		add("conflicting "+field, "trace_call", 0, traceObject{"gasPrice": "0x0", field: "0x1"}, false)
	}
	add("nonempty tuples", "trace_callMany", 0, []any{[]any{traceObject{"to": address}, []any{"trace"}}, []any{traceObject{}, []any{"stateDiff", "vmTrace"}}}, true)
	add("invalid second envelope", "trace_callMany", -1, []any{envelope(call), envelope(reward)}, false)
	add("invalid second tuple", "trace_callMany", 0, []any{[]any{traceObject{}, []any{}}, []any{traceObject{}, []any{"bad"}}}, false)
	add("invalid second selection", "trace_call", 1, []any{"trace", "bad"}, false)
	add("short tuple", "trace_callMany", 0, []any{[]any{traceObject{}}}, false)
	add("long tuple", "trace_callMany", 0, []any{[]any{traceObject{}, []any{}, 0}}, false)
	add("reversed tuple", "trace_callMany", 0, []any{[]any{[]any{}, traceObject{}}}, false)
	add("duplicate selections", "trace_callMany", 0, []any{[]any{traceObject{}, []any{"trace", "trace"}}}, false)
	add("nonempty envelopes", "trace_callMany", -1, []any{envelope(call), envelope(create)}, true)
	add("reward in sequence", "trace_callMany", -1, []any{envelope(reward)}, false)
	diff := envelope(call)
	diff["stateDiff"] = traceObject{address: traceObject{"balance": "=", "nonce": "=", "code": "=", "storage": traceObject{hash: traceObject{"*": traceObject{"from": hash, "to": hash}}}}}
	add("state change", "trace_call", -1, diff, true)
	slot := "0x" + strings.Repeat("0", 64)
	accountDiff := func(balance, nonce, code any, storage traceObject) traceObject {
		e := envelope(call)
		e["stateDiff"] = traceObject{address: traceObject{"balance": balance, "nonce": nonce, "code": code, "storage": storage}}
		return e
	}
	changed := traceObject{"*": traceObject{"from": "0x1", "to": "0x2"}}
	add("changed balance", "trace_call", -1, accountDiff(changed, "=", "=", traceObject{}), true)
	add("unchanged account", "trace_call", -1, accountDiff("=", "=", "=", traceObject{}), false)
	add("zero slot write", "trace_call", -1, accountDiff("=", "=", "=", traceObject{slot: traceObject{"*": traceObject{"from": slot, "to": hash}}}), true)
	add("slot born on existing account", "trace_call", -1, accountDiff("=", "=", "=", traceObject{slot: traceObject{"+": hash}}), false)
	add("slot died on existing account", "trace_call", -1, accountDiff(changed, "=", "=", traceObject{slot: traceObject{"-": hash}}), false)
	add("unchanged slot", "trace_call", -1, accountDiff(changed, "=", "=", traceObject{slot: "="}), false)
	born := func(storage traceObject) traceObject {
		return accountDiff(traceObject{"+": "0x0"}, traceObject{"+": "0x1"}, traceObject{"+": "0x00"}, storage)
	}
	add("born account", "trace_call", -1, born(traceObject{slot: traceObject{"+": hash}}), true)
	add("born account changed slot", "trace_call", -1, born(traceObject{slot: traceObject{"*": traceObject{"from": slot, "to": hash}}}), false)
	add("born account unchanged nonce", "trace_call", -1, accountDiff(traceObject{"+": "0x1"}, "=", traceObject{"+": "0x"}, traceObject{}), false)
	died := func(storage traceObject) traceObject {
		return accountDiff(traceObject{"-": "0x0"}, traceObject{"-": "0x1"}, traceObject{"-": "0x00"}, storage)
	}
	add("deleted account", "trace_call", -1, died(traceObject{}), true)
	add("deleted account storage", "trace_call", -1, died(traceObject{slot: traceObject{"*": traceObject{"from": hash, "to": slot}}}), false)
	add("mixed markers", "trace_call", -1, accountDiff(traceObject{"-": "0x1"}, "=", "=", traceObject{}), false)
	for _, frame := range []traceObject{call, reward, local(call)} {
		replay := envelope(frame)
		replay["transactionHash"] = hash
		add("replay "+frame["type"].(string), "trace_replayTransaction", -1, replay, frame["type"] == "call" && frame["blockHash"] == nil)
		add("block replay "+frame["type"].(string), "trace_replayBlockTransactions", -1, []any{replay}, frame["type"] == "call" && frame["blockHash"] == nil)
	}
	for _, method := range []string{"trace_get", "trace_transaction", "trace_replayTransaction"} {
		add("missing transaction", method, -1, nil, true)
	}
	// A deeply malformed child must fail through the actual speccheck serialization path.
	vm := traceObject{"code": "0x00", "ops": []any{}}
	for range 5 {
		vm = traceObject{"code": "0x00", "ops": []any{traceObject{"pc": 0, "cost": 0, "ex": traceObject{"used": 0, "push": []any{}, "mem": nil, "store": nil}, "sub": vm}}}
	}
	e := envelope(call)
	e["vmTrace"] = vm
	add("nested VM", "trace_call", -1, e, true)
	badVM := traceCopy(t, vm)
	leaf := badVM
	for range 5 {
		leaf = leaf["ops"].([]any)[0].(traceObject)["sub"].(traceObject)
	}
	leaf["code"] = "not hex"
	e = envelope(call)
	e["vmTrace"] = badVM
	add("malformed deep VM", "trace_call", -1, e, false)

	for _, tc := range probes {
		t.Run(tc.method+"/"+tc.name, func(t *testing.T) {
			for _, artifact := range []struct {
				name string
				doc  traceObject
			}{{"refs", refs}, {"expanded", expanded}} {
				t.Run(artifact.name, func(t *testing.T) {
					var schema traceObject
					for _, m := range artifact.doc["methods"].([]any) {
						method := m.(traceObject)
						if method["name"] != tc.method {
							continue
						}
						if tc.param < 0 {
							schema = method["result"].(traceObject)["schema"].(traceObject)
						} else {
							schema = method["params"].([]any)[tc.param].(traceObject)["schema"].(traceObject)
						}
					}
					root := traceCopy(t, schema)
					root["components"] = artifact.doc["components"]
					for _, draft := range []*schema5.Draft{schema5.Draft7, schema5.Draft2019} {
						compiler := schema5.NewCompiler()
						compiler.Draft = draft
						if err := compiler.AddResource("https://example.test/trace.json", bytes.NewReader(traceJSON(t, root))); err != nil {
							t.Fatal(err)
						}
						compiled, err := compiler.Compile("https://example.test/trace.json")
						if err != nil {
							t.Fatal(err)
						}
						err = compiled.Validate(tc.value)
						if (err == nil) != tc.valid {
							t.Fatalf("valid=%v: %v", tc.valid, err)
						}
					}
				})
			}
			t.Run("speccheck", func(t *testing.T) {
				cd := parsed[tc.method].result
				if tc.param >= 0 {
					cd = parsed[tc.method].params[tc.param]
				}
				err := validate(cd.schema, traceJSON(t, tc.value), "https://example.test/speccheck.json")
				var validationErr *schema5.ValidationError
				if err != nil && !errors.As(err, &validationErr) {
					t.Fatalf("compile/serialization error: %v", err)
				}
				if (err == nil) != tc.valid {
					t.Fatalf("valid=%v: %v", tc.valid, err)
				}
			})
		})
	}
}

func TestTraceVmResource(t *testing.T) {
	refs, _, _ := traceDocuments(t)
	vm := refs["components"].(traceObject)["schemas"].(traceObject)["TraceVm"].(traceObject)
	// Compile at its own declared identity, with no surrounding components to hide a scope defect.
	for _, draft := range []*schema6.Draft{schema6.Draft7, schema6.Draft2019} {
		compiler := schema6.NewCompiler()
		compiler.DefaultDraft(draft)
		id := vm["$id"].(string)
		if err := compiler.AddResource(id, vm); err != nil {
			t.Fatal(err)
		}
		compiled, err := compiler.Compile(id)
		if err != nil {
			t.Fatal(err)
		}
		for _, tc := range []struct {
			input string
			valid bool
		}{
			{`{"code":"0x","ops":[]}`, true},
			{`{"code":"0x","ops":[{"pc":0,"cost":0,"ex":{"used":0,"push":["0x1"],"mem":{"off":0,"data":"0x00"},"store":{"key":"0x0","val":"0x1"}},"sub":{"code":"0x","ops":[]}}]}`, true},
			{`{"code":"0x","ops":[{"pc":0,"cost":0,"ex":null,"sub":{"code":"bad","ops":[]}}]}`, false},
			{`{"code":"0x","ops":[{"pc":0,"cost":0,"ex":{"used":0,"push":["0x00"],"mem":null,"store":null},"sub":null}]}`, false},
			{`{"code":"0x","ops":[{"pc":0,"cost":0,"ex":null,"sub":null}]}`, true},
			{`{"code":"0x","ops":[{"pc":0,"cost":0,"ex":null,"sub":{"code":"0x","ops":[]}}]}`, false},
		} {
			var input any
			if err := json.Unmarshal([]byte(tc.input), &input); err != nil {
				t.Fatal(err)
			}
			err := compiled.Validate(input)
			if (err == nil) != tc.valid {
				t.Fatalf("valid=%v: %v", tc.valid, err)
			}
		}
	}
}

func TestTraceExamples(t *testing.T) {
	_, document, methods := traceDocuments(t)
	for _, entry := range document["methods"].([]any) {
		method := entry.(traceObject)
		name := method["name"].(string)
		for _, entry := range method["examples"].([]any) {
			example := entry.(traceObject)
			t.Run(name+"/"+example["name"].(string), func(t *testing.T) {
				for i, entry := range example["params"].([]any) {
					param := entry.(traceObject)
					descriptor := methods[name].params[i]
					if param["name"] != descriptor.name {
						t.Fatalf("example parameter %d has wrong name", i)
					}
					if err := validate(descriptor.schema, traceJSON(t, param["value"]), "https://example.test/parameter.json"); err != nil {
						t.Fatal(err)
					}
				}
				result := example["result"].(traceObject)
				if err := validate(methods[name].result.schema, traceJSON(t, result["value"]), "https://example.test/result.json"); err != nil {
					t.Fatal(err)
				}
			})
		}
	}
}
