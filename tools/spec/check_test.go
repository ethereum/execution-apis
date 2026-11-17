package spec

import (
	"encoding/json"
	"testing"
)

const errorDataSpec = `{
  "openrpc": "1.4.1",
  "info": {"title": "test", "version": "1"},
  "methods": [
    {
      "name": "eth_call",
      "params": [],
      "result": {"name": "result", "schema": {"type": "string"}},
      "errors": [
        {"code": 3, "message": "execution reverted", "data": {"type": "string", "pattern": "^0x[0-9a-f]*$"}},
        {"code": -32015, "message": "vm error"}
      ]
    },
    {
      "name": "engine_getPayloadV1",
      "params": [],
      "result": {"name": "result", "schema": {"type": "string"}},
      "errors": [
        {"code": -38001, "message": "Unknown payload"}
      ]
    }
  ]
}`

func TestCheckErrorData(t *testing.T) {
	s, err := LoadString("test", errorDataSpec)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		method string
		error  string
		valid  bool
	}{
		{"eth_call", `{"code":3,"message":"execution reverted","data":"0x"}`, true},
		{"eth_call", `{"code":3,"message":"execution reverted","data":"0x08c379a0"}`, true},
		{"eth_call", `{"code":3,"message":"execution reverted"}`, false},
		{"eth_call", `{"code":3,"message":"execution reverted","data":null}`, false},
		{"eth_call", `{"code":3,"message":"execution reverted","data":"zz"}`, false},
		{"eth_call", `{"code":-32015,"message":"vm error"}`, true},
		{"eth_call", `{"code":-32015,"message":"vm error","data":"text"}`, false},
		{"eth_call", `{"code":-32602,"message":"invalid argument 0"}`, true},
		{"eth_call", `{"code":-32602,"message":"Invalid params","data":"text"}`, false},
		{"eth_call", `{"code":-32000,"message":"header not found","data":null}`, false},
		{"eth_call", `{"code":-32001,"message":"Resource not found"}`, false},
		{"engine_getPayloadV1", `{"code":-38001,"message":"Unknown payload","data":null}`, true},
		{"engine_getPayloadV1", `{"code":-38001,"message":"Unknown payload","data":"text"}`, false},
		{"engine_getPayloadV1", `{"code":-32000,"message":"Server error","data":{"err":"Database corrupted"}}`, true},
	}
	for _, test := range tests {
		c := NewChecker(s)
		call := `{"jsonrpc":"2.0","id":1,"method":"` + test.method + `","params":[]}`
		if err := c.Request(json.RawMessage(call)); err != nil {
			t.Fatalf("request %s: %v", test.method, err)
		}
		resp := `{"jsonrpc":"2.0","id":1,"error":` + test.error + `}`
		err := c.Response(json.RawMessage(resp))
		if test.valid && err != nil {
			t.Errorf("%s %s: unexpected error: %v", test.method, test.error, err)
		}
		if !test.valid && err == nil {
			t.Errorf("%s %s: expected error", test.method, test.error)
		}
	}
}
