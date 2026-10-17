package iofile

import (
	"bytes"
	"encoding/json"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/execution-apis/tools/spec"
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
openrpc.validate("#/components/schemas/AccessList", messages[0].send.valid);
openrpc.validate("#/components/schemas/AccessList", messages[1].send.invalid);
`
	test, err := Load("schema.io", strings.NewReader(testFile))
	if err != nil {
		t.Fatal("load failed: ", test, err)
	}

	spec, err := spec.LoadString("test", `{
  "openrpc": "1.4.1",
  "methods": [],
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
	if err != nil {
		t.Fatalf("invalid spec: %v", err)
	}

	config := ScriptConfig{OpenRPC: spec}
	err = test.RunScript(config, nil)
	if err == nil {
		t.Fatal("no error from script")
	}
	if !strings.Contains(err.Error(), "jsonschema validation failed with") {
		t.Fatalf("wrong error from schema validation: %v", err)
	}
	if !strings.Contains(err.Error(), "at schema.io:6:17") {
		t.Fatalf("backtrace line not in error from schema validation: %v", err)
	}
}

func TestValidationScriptIsValidException(t *testing.T) {
	testFile := `
>> {"valid": [{"address": "0xa02457e5dfd32bda5fc7e1f1b008aa5979568150", "storageKeys": ["0x0000000000000000000000000000000000000000000000000000000000000081"]}]}
--
let hash = messages[0].send.valid[0].storageKeys[0];
if (!openrpc.isValid("#/components/schemas/hash32", "0x")) {
  console.log("hash-value-invalid");
}
openrpc.isValid("#/components/schemas/unknown", messages[0].send.valid);
`
	test, err := Load("schema.io", strings.NewReader(testFile))
	if err != nil {
		t.Fatal("load failed: ", test, err)
	}

	spec, err := spec.LoadString("test", `{
  "openrpc": "1.4.1",
  "methods": [],
  "components": {
    "schemas": {
      "hash32": {
        "pattern": "^0x[0-9a-f]{64}$",
        "title": "32 byte hex value",
        "type": "string"
      }
    }
  }
}`)
	if err != nil {
		t.Fatalf("invalid spec: %v", err)
	}

	var logbuffer bytes.Buffer
	config := ScriptConfig{OpenRPC: spec, Log: Logger{&logbuffer}}
	err = test.RunScript(config, nil)
	if err == nil {
		t.Fatal("no error from script")
	}
	if !strings.Contains(logbuffer.String(), "hash-value-invalid") {
		t.Errorf("expected output not found in log")
	}
	re := regexp.MustCompile("json-pointer in \".*\" not found")
	if !re.MatchString(err.Error()) {
		t.Fatalf("wrong error from schema validation: %v", err)
	}
}
