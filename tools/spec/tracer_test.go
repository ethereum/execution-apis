package spec

import (
	"fmt"
	"strings"
	"testing"
)

const txResultSchema = `{
  "anyOf": [
    {"title": "Opcode tracer result", "type": "object", "required": ["structLogs"]},
    {"title": "Call tracer result", "type": "object", "required": ["type", "from"]},
    {"title": "Named tracer result", "description": "any JSON value"}
  ]
}`

const blockResultSchema = `{
  "title": "Array of per-transaction traces",
  "type": "array",
  "items": {
    "anyOf": [
      {"title": "Opcode tracer entry", "type": "object", "required": ["txHash", "result"]},
      {"title": "Call tracer entry", "type": "object", "required": ["txHash"],
       "anyOf": [{"required": ["result"]}, {"required": ["error"]}]},
      {"title": "Named tracer entry", "type": "object"}
    ]
  }
}`

const noCallBranchSchema = `{"anyOf": [
  {"title": "Opcode tracer result", "type": "object"},
  {"title": "Named tracer result"}
]}`

func tracerTestSpec(t *testing.T) *Spec {
	t.Helper()
	method := func(name, result string, params int) string {
		ps := make([]string, params)
		for i := range ps {
			ps[i] = fmt.Sprintf(`{"name": "p%d", "schema": {}}`, i)
		}
		return fmt.Sprintf(`{"name": %q, "params": [%s], "result": {"name": "r", "schema": %s}}`,
			name, strings.Join(ps, ","), result)
	}
	doc := `{"openrpc": "1.2.4", "info": {"title": "t", "version": "1"}, "methods": [` +
		strings.Join([]string{
			method("debug_traceTransaction", txResultSchema, 2),
			method("debug_traceBlockByNumber", blockResultSchema, 2),
			method("debug_traceCall", txResultSchema, 3),
			method("debug_traceBlockByHash", noCallBranchSchema, 2),
			method("eth_call", txResultSchema, 2),
		}, ",") + `]}`
	s, err := LoadString("test.json", doc)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func checkCall(s *Spec, method, params, result string) error {
	c := NewChecker(s)
	req := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":%q,"params":%s}`, method, params)
	if err := c.Request([]byte(req)); err != nil {
		return err
	}
	return c.Response([]byte(fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"result":%s}`, result)))
}

func TestTracerResultSchema(t *testing.T) {
	s := tracerTestSpec(t)
	for _, tc := range []struct {
		name   string
		method string
		params string
		result string
		valid  bool
	}{
		{"callTracer frame", "debug_traceTransaction", `["0xabc",{"tracer":"callTracer"}]`, `{"type":"CALL","from":"0xaa"}`, true},
		{"callTracer frame without type", "debug_traceTransaction", `["0xabc",{"tracer":"callTracer"}]`, `{"from":"0xaa"}`, false},
		{"callTracer given opcode result", "debug_traceTransaction", `["0xabc",{"tracer":"callTracer"}]`, `{"structLogs":[]}`, false},
		{"no tracer selects opcode logger", "debug_traceTransaction", `["0xabc"]`, `{"structLogs":[]}`, true},
		{"no tracer given call frame", "debug_traceTransaction", `["0xabc"]`, `{"type":"CALL","from":"0xaa"}`, false},
		{"unknown tracer keeps whole schema", "debug_traceTransaction", `["0xabc",{"tracer":"prestateTracer"}]`, `42`, true},
		{"traceCall reads the third param", "debug_traceCall", `[{},"latest",{"tracer":"callTracer"}]`, `{"from":"0xaa"}`, false},
		{"block entry with result", "debug_traceBlockByNumber", `["0x1",{"tracer":"callTracer"}]`, `[{"txHash":"0xaa","result":{}}]`, true},
		{"block entry with error", "debug_traceBlockByNumber", `["0x1",{"tracer":"callTracer"}]`, `[{"txHash":"0xaa","error":"boom"}]`, true},
		{"block entry with neither", "debug_traceBlockByNumber", `["0x1",{"tracer":"callTracer"}]`, `[{"txHash":"0xaa"}]`, false},
		{"block entry without txHash", "debug_traceBlockByNumber", `["0x1",{"tracer":"callTracer"}]`, `[{"result":{}}]`, false},
		{"non-trace method keeps whole schema", "eth_call", `[{},"latest"]`, `42`, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkCall(s, tc.method, tc.params, tc.result)
			if tc.valid && err != nil {
				t.Errorf("valid result rejected: %v", err)
			}
			if !tc.valid && err == nil {
				t.Error("invalid result accepted")
			}
		})
	}
}

func TestTracerMissingBranch(t *testing.T) {
	s := tracerTestSpec(t)
	c := NewChecker(s)
	req := `{"jsonrpc":"2.0","id":1,"method":"debug_traceBlockByHash","params":["0xabc",{"tracer":"callTracer"}]}`
	if err := c.Request([]byte(req)); err == nil {
		t.Fatal("expected an error when the tracer is known but no branch matches")
	}
}
