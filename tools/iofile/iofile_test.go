package iofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestValidationScript(t *testing.T) {
	testFile := `
// gets the current gas price in wei
// speconly: client response is only checked for schema validity.
>> {"jsonrpc":"2.0","id":1,"method":"eth_gasPrice"}
<< {"jsonrpc":"2.0","id":1,"result":"0x1047435"}
--
console.log("hello")
if (BigInt(messages[1].response.result) <= 0) {
	throw new Error("gasprice too low");
}
`
	test, err := Load("gasprice.io", strings.NewReader(testFile))
	if err != nil {
		t.Fatal("load failed: ", test)
	}
	if !test.SpecOnly {
		t.Errorf("speconly wasn't recognized")
	}
	if test.Script == "" {
		t.Errorf("validation script wasn't recognized")
	}

	validResp := []json.RawMessage{
		json.RawMessage(`{"jsonrpc":"2.0","id":1,"result":"0x133"}`),
	}
	logbuf := new(bytes.Buffer)
	scriptLog := Logger{logbuf}
	if err := test.RunScript(ScriptConfig{Log: scriptLog}, validResp); err != nil {
		t.Fatal("validation script failed:", err)
	}
	if !bytes.Equal(logbuf.Bytes(), []byte("hello\n")) {
		t.Errorf("script console output not captured: %q", logbuf.Bytes())
	}

	invalidResp := []json.RawMessage{
		json.RawMessage(`{"jsonrpc":"2.0","id":1,"result":"0x0"}`),
	}
	if err := test.RunScript(ScriptConfig{Log: scriptLog}, invalidResp); err == nil {
		t.Fatal("validation script didn't fail for invalid response")
	} else if !strings.Contains(err.Error(), "gasprice too low") {
		t.Fatalf("validation script had wrong error message: %q", err)
	}
}

func TestValidationScriptTimeout(t *testing.T) {
	testFile := "--\nwhile (true) {}\n"
	test, err := Load("loop.io", strings.NewReader(testFile))
	if err != nil {
		t.Fatal("load failed: ", test)
	}
	config := ScriptConfig{Timeout: 500 * time.Millisecond}
	err = test.RunScript(config, nil)
	if err == nil {
		t.Fatal("no error from script")
	}
	if !errors.Is(err, errExecutionTimeout) {
		t.Error("script error is not timeout")
	}
}

func TestValidationScriptSchema(t *testing.T) {
	testFile := `
>> {"valid": [{"address": "0xa02457e5dfd32bda5fc7e1f1b008aa5979568150", "storageKeys": ["0x0000000000000000000000000000000000000000000000000000000000000081"]}]}
>> {"invalid": [{"address": "0xa02457e5dfd32bda5fc7e1f1b008aa5979568150", "storageKeys": ["0x00000000000000000000000000000000000000000000000000000000000000"]}]}
--
jsonschema.validate(openrpc, "#/components/schemas/AccessList", messages[0].send.valid);
jsonschema.validate(openrpc, "#/components/schemas/AccessList", messages[1].send.invalid);
`
	test, err := Load("schema.io", strings.NewReader(testFile))
	if err != nil {
		t.Fatal("load failed: ", test, err)
	}

	schema := json.RawMessage(`{
  "openrpc": "1.4.1",
  "components": {
    "schemas": {
      "AccessList": {
        "items": {
          "$ref": "#/components/schemas/AccessListEntry"
        },
        "title": "Access list",
        "type": "array"
      },
      "AccessListEntry": {
        "additionalProperties": false,
        "properties": {
          "address": {
            "$ref": "#/components/schemas/address"
          },
          "storageKeys": {
            "items": {
              "$ref": "#/components/schemas/hash32"
            },
            "type": "array"
          }
        },
        "required": [
          "address",
          "storageKeys"
        ],
        "title": "Access list entry",
        "type": "object"
      },
      "address": {
        "pattern": "^0x[0-9a-fA-F]{40}$",
        "title": "hex encoded address",
        "type": "string"
      },
      "hash32": {
        "pattern": "^0x[0-9a-f]{64}$",
        "title": "32 byte hex value",
        "type": "string"
      }
    }
  }
}`)
	config := ScriptConfig{OpenRPCSchema: schema}
	err = test.RunScript(config, nil)
	if err == nil {
		t.Fatal("no error from script")
	}
	if !strings.Contains(err.Error(), "jsonschema validation failed with") {
		t.Fatalf("wrong error from schema validation: %v", err)
	}
	if !strings.Contains(err.Error(), "at schema.io:6:20") {
		t.Fatalf("backtrace line not in error from schema validation: %v", err)
	}

}
