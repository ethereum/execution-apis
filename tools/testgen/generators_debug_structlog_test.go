package testgen

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ethereum/execution-apis/tools/iofile"
	"github.com/ethereum/execution-apis/tools/spec"
)

func TestStructLogCompareScript(t *testing.T) {
	stub, err := spec.LoadString("stub.json", `{
		"openrpc": "1.2.4", "info": {"title": "stub", "version": "1"}, "methods": [],
		"components": {"schemas": {"OpcodeTransactionTrace": {}}}
	}`)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"trace-nested-revert-and-halt", "trace-state-gas-oog"} {
		fixture := loadStructLogFixture(t, "debug_traceTransaction/"+name)
		var recorded map[string]any
		if err := json.Unmarshal(fixture.Receives()[0], &recorded); err != nil {
			t.Fatal(err)
		}
		failing := -1
		for i, log := range recorded["result"].(map[string]any)["structLogs"].([]any) {
			if _, ok := log.(map[string]any)["error"]; ok {
				failing = i
			}
		}
		if failing < 0 {
			t.Fatalf("%s: no failing step", name)
		}
		for _, tc := range []struct {
			name   string
			mutate func(logs []any)
			valid  bool
		}{
			{"unchanged", func([]any) {}, true},
			{"other error text", func(l []any) { step(l, failing)["error"] = "other client text" }, true},
			{"empty error", func(l []any) { step(l, failing)["error"] = "" }, false},
			{"missing error", func(l []any) { delete(step(l, failing), "error") }, false},
			{"gasCost", func(l []any) { step(l, failing)["gasCost"] = 10100 }, false},
			{"extra field", func(l []any) { step(l, failing)["reason"] = "0x" }, false},
			{"error on a completed step", func(l []any) { step(l, 0)["error"] = "out of gas" }, false},
		} {
			t.Run(name+"/"+tc.name, func(t *testing.T) {
				var resp map[string]any
				if err := json.Unmarshal(fixture.Receives()[0], &resp); err != nil {
					t.Fatal(err)
				}
				tc.mutate(resp["result"].(map[string]any)["structLogs"].([]any))
				mutated, _ := json.Marshal(resp)
				err := fixture.RunScript(iofile.ScriptConfig{OpenRPC: stub, Log: t}, []json.RawMessage{mutated})
				if tc.valid && err != nil {
					t.Errorf("valid response rejected: %v", err)
				}
				if !tc.valid && err == nil {
					t.Error("invalid response accepted")
				}
			})
		}
	}
}

func loadStructLogFixture(t *testing.T, name string) iofile.Test {
	t.Helper()
	fd, err := os.Open(filepath.Join("../../tests", name+".io"))
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

func step(logs []any, i int) map[string]any {
	return logs[i].(map[string]any)
}
