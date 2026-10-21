package testgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ethereum/execution-apis/tools/iofile"
	"github.com/ethereum/execution-apis/tools/spec"
)

const testsDir = "../../tests"

// Without a script, a speconly trace response is checked only against an anyOf
// that accepts any value.
func TestTraceFixturesHaveScripts(t *testing.T) {
	tests, err := iofile.LoadDirectory(t, testsDir, regexp.MustCompile(`^debug_trace`))
	if err != nil {
		t.Fatal(err)
	}
	if len(tests) == 0 {
		t.Fatal("no debug_trace fixtures found")
	}
	for _, test := range tests {
		var req struct {
			Method string            `json:"method"`
			Params []json.RawMessage `json:"params"`
		}
		var resp struct {
			Result json.RawMessage `json:"result"`
		}
		if err := json.Unmarshal(test.Messages[0].Data, &req); err != nil {
			t.Fatalf("%s: %v", test.Name, err)
		}
		if err := json.Unmarshal(test.Receives()[0], &resp); err != nil {
			t.Fatalf("%s: %v", test.Name, err)
		}
		if resp.Result == nil || !test.SpecOnly {
			continue
		}
		cfgIndex := 1
		if req.Method == "debug_traceCall" {
			cfgIndex = 2
		}
		var cfg struct {
			Tracer string `json:"tracer"`
		}
		if len(req.Params) > cfgIndex {
			json.Unmarshal(req.Params[cfgIndex], &cfg)
		}
		schema := "#/components/schemas/Opcode"
		if cfg.Tracer == "callTracer" {
			schema = "#/components/schemas/Call"
		}
		if !strings.Contains(test.Script, schema) {
			t.Errorf("%s: script does not validate against a %s* schema", test.Name, schema)
		}
		if cfg.Tracer == "callTracer" && !strings.Contains(test.Script, "m.expected") {
			t.Errorf("%s: speconly callTracer fixture does not compare with the recorded response", test.Name)
		}
	}
}

func TestTraceCompareScript(t *testing.T) {
	stub, err := spec.LoadString("stub.json", `{
		"openrpc": "1.2.4", "info": {"title": "stub", "version": "1"}, "methods": [],
		"components": {"schemas": {"CallFrame": {}, "CallTracerBlockEntry": {}}}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	const (
		panicOutput = "0x4e487b710000000000000000000000000000000000000000000000000000000000000001"
		panicFrame  = `{"type":"CALL","from":"0x01","to":"0x02","gas":"0x10","gasUsed":"0x8","input":"0x","error":"execution reverted","output":"` + panicOutput + `","revertReason":"assert(false)"}`
		panicTest   = `>> {"jsonrpc":"2.0","id":1,"method":"debug_traceTransaction","params":["0x01",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":` + panicFrame + `}
--
` + "%s"
		entryTest = `>> {"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByNumber","params":["0x1",{"tracer":"callTracer"}]}
<< {"jsonrpc":"2.0","id":1,"result":[{"txHash":"0x01","error":"tracer timeout"}]}
--
` + "%s"
	)
	panicFixture := loadInline(t, "panic", strings.Replace(panicTest, "%s", callFrameCompareScript, 1))
	entryFixture := loadInline(t, "entry", strings.Replace(entryTest, "%s", callBlockCompareScript, 1))
	nested := loadFixture(t, "debug_traceTransaction/calltracer-nested-calls")
	block := loadFixture(t, "debug_traceBlockByNumber/calltracer-block-with-log")

	for _, tc := range []struct {
		name   string
		test   iofile.Test
		mutate func(resp map[string]any)
		valid  bool
	}{
		{"unchanged frame", nested, func(map[string]any) {}, true},
		{"write-protection error text", nested, func(r map[string]any) { call(r, 3)["error"] = "write protection" }, true},
		{"nested gasUsed", nested, func(r map[string]any) { call(r, 0)["gasUsed"] = "0x1" }, false},
		{"REVERT error text", nested, func(r map[string]any) { call(r, 1)["error"] = "reverted" }, false},
		{"missing error", nested, func(r map[string]any) { delete(call(r, 3), "error") }, false},
		{"Error(string) revertReason", nested, func(r map[string]any) { call(r, 1)["revertReason"] = "other" }, false},
		{"unchanged block", block, func(map[string]any) {}, true},
		{"txHash", block, func(r map[string]any) { entry(r, 0)["txHash"] = entry(r, 1)["txHash"] }, false},
		{"missing entry", block, func(r map[string]any) { r["result"] = r["result"].([]any)[1:] }, false},
		{"Panic revertReason text", panicFixture, func(r map[string]any) { result(r)["revertReason"] = "assertion failed" }, true},
		{"Panic missing revertReason", panicFixture, func(r map[string]any) { delete(result(r), "revertReason") }, false},
		{"Panic output", panicFixture, func(r map[string]any) { result(r)["output"] = "0x" }, false},
		{"block entry error text", entryFixture, func(r map[string]any) { entry(r, 0)["error"] = "timeout" }, true},
		{"block entry missing error", entryFixture, func(r map[string]any) { delete(entry(r, 0), "error") }, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var resp map[string]any
			if err := json.Unmarshal(tc.test.Receives()[0], &resp); err != nil {
				t.Fatal(err)
			}
			tc.mutate(resp)
			mutated, _ := json.Marshal(resp)
			err := tc.test.RunScript(iofile.ScriptConfig{OpenRPC: stub, Log: t}, []json.RawMessage{mutated})
			if tc.valid && err != nil {
				t.Errorf("valid response rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("invalid response accepted")
			}
		})
	}

	errResp := json.RawMessage(`{"jsonrpc":"2.0","id":1,"error":{"code":-32000,"message":"boom"}}`)
	if err := nested.RunScript(iofile.ScriptConfig{OpenRPC: stub, Log: t}, []json.RawMessage{errResp}); err == nil {
		t.Error("error response accepted")
	}
}

func loadFixture(t *testing.T, name string) iofile.Test {
	t.Helper()
	fd, err := os.Open(filepath.Join(testsDir, name+".io"))
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()
	test, err := iofile.Load(name, fd)
	if err != nil {
		t.Fatal(err)
	}
	return test
}

func loadInline(t *testing.T, name, source string) iofile.Test {
	t.Helper()
	test, err := iofile.Load(name, strings.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	return test
}

func result(resp map[string]any) map[string]any {
	return resp["result"].(map[string]any)
}

func call(resp map[string]any, i int) map[string]any {
	return result(resp)["calls"].([]any)[i].(map[string]any)
}

func entry(resp map[string]any, i int) map[string]any {
	return resp["result"].([]any)[i].(map[string]any)
}
