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
