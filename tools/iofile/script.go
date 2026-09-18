package iofile

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/dop251/goja"
	"github.com/dop251/goja_nodejs/console"
	"github.com/dop251/goja_nodejs/require"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type scriptMessage struct {
	Send     json.RawMessage `json:"send,omitempty"`
	Expected json.RawMessage `json:"expected,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
}

// ScriptConfig holds settings for validation script execution.
type ScriptConfig struct {
	Log           TestLogger      // script console output is written here
	Timeout       time.Duration   // execution timeout
	OpenRPCSchema json.RawMessage // this is accessible to the script for validation purposes
}

func (cfg ScriptConfig) withDefaults() ScriptConfig {
	if cfg.Log == nil {
		cfg.Log = StdoutLogger
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 20 * time.Second
	}
	return cfg
}

// RunScript runs the test's validation script.
//
// `responses` are the response messages from the server.
// There must be one response for each receive (<<) line in the test.
func (t *Test) RunScript(config ScriptConfig, responses []json.RawMessage) error {
	if n := len(t.Receives()); len(responses) != n {
		panic(fmt.Errorf("%d server responses given, but test has %d receive lines", len(responses), n))
	}
	if t.Script == "" {
		return nil
	}
	config = config.withDefaults()

	// Serialize input messages.
	scriptmsg := make([]scriptMessage, len(t.Messages))
	var respIndex int
	for i, msg := range t.Messages {
		if msg.Send {
			scriptmsg[i].Send = msg.Data
		} else {
			scriptmsg[i].Expected = msg.Data
			scriptmsg[i].Response = responses[respIndex]
			respIndex++
		}
	}
	scriptmsgJSON, err := json.Marshal(scriptmsg)
	if err != nil {
		panic("message serialization failed: " + err.Error())
	}

	// Pad the script with newlines so the line numbers reported for
	// goja errors match the .io file.
	source := strings.Repeat("\n", t.scriptStartLine) + t.Script
	return runValidationScript(t.Name, source, scriptmsgJSON, config)
}

var errExecutionTimeout = errors.New("script execution timeout reached")

func runValidationScript(sourceFile string, source string, jsonMessages []byte, config ScriptConfig) error {
	vm := goja.New()

	// Parse messages into goja object.
	jsonObj := vm.Get("JSON").ToObject(vm)
	parse, ok := goja.AssertFunction(jsonObj.Get("parse"))
	if !ok {
		panic("JSON.parse is not a function")
	}
	v, err := parse(jsonObj, vm.ToValue(string(jsonMessages)))
	if err != nil {
		panic(err) // *goja.Exception for a SyntaxError from bad JSON
	}
	vm.Set("messages", v)

	// Enable console output to w.
	reg := new(require.Registry)
	reg.Enable(vm)
	reg.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(scriptPrinter{config.Log}))
	reg.RegisterNativeModule(jsonschemaModuleName, requireJSONSchemaModule)
	console.Enable(vm)
	enableJSONSchemaModule(vm)

	// Store the RPC schema into a global if configured.
	if len(config.OpenRPCSchema) > 0 {
		schema, err := parse(jsonObj, vm.ToValue(string(config.OpenRPCSchema)))
		if err != nil {
			return fmt.Errorf("invalid OpenRPC schema in ScriptConfig: %v", err)
		}
		vm.Set("openrpc", schema)
	}

	// Set up interrupt timeout.
	done := make(chan struct{})
	go func() {
		select {
		case <-time.After(config.Timeout):
			vm.Interrupt(errExecutionTimeout)
		case <-done:
		}
	}()

	_, err = vm.RunScript(sourceFile, source)
	close(done)
	if gerr, ok := err.(*goja.Exception); ok {
		// Using String() to get multi-line stack.
		return errors.New(gerr.String())
	}
	return err
}

// -- Console Logger

// StdoutLogger implements TestLogger, printing to stdout.
var StdoutLogger = Logger{os.Stdout}

// Logger implements TestLogger, writing to an io.Writer.
type Logger struct {
	io.Writer
}

func (l Logger) Logf(format string, args ...any) {
	fmt.Fprintf(l.Writer, format+"\n", args...)
}

type scriptPrinter struct {
	w TestLogger
}

func (p scriptPrinter) Log(msg string) {
	p.w.Logf("%s", msg)
}

func (p scriptPrinter) Warn(msg string) {
	p.w.Logf("warn: %s", msg)
}

func (p scriptPrinter) Error(msg string) {
	p.w.Logf("error: %s", msg)
}

// -- JSON-Schema Module

const jsonschemaModuleName = "jsonschema"

type jsonschemaModule struct {
	vm *goja.Runtime
}

func requireJSONSchemaModule(runtime *goja.Runtime, module *goja.Object) {
	m := jsonschemaModule{vm: runtime}
	o := module.Get("exports").(*goja.Object)
	o.Set("validate", m.validate)
	o.Set("isValid", m.isValid)
}

func enableJSONSchemaModule(runtime *goja.Runtime) {
	runtime.Set("jsonschema", require.Require(runtime, jsonschemaModuleName))
}

// validate(schema, value), throws when invalid.
func (m *jsonschemaModule) validate(call goja.FunctionCall) goja.Value {
	err := m.doValidate(call)
	if err != nil {
		panic(m.vm.NewGoError(err))
	}
	return goja.Undefined()
}

// isValid(schema, value) -> boolean
func (m *jsonschemaModule) isValid(call goja.FunctionCall) goja.Value {
	err := m.doValidate(call)
	return m.vm.ToValue(err == nil)
}

func (m *jsonschemaModule) doValidate(call goja.FunctionCall) error {
	if len(call.Arguments) < 2 || len(call.Arguments) > 3 {
		throw(m.vm, "invalid number of arguments (%d), need (schema, value, [url])", len(call.Arguments))
	}
	schema := call.Arguments[0]
	value := call.Arguments[1]
	url := ""
	if len(call.Arguments) > 2 {
		url = call.Arguments[2].ToString().String()
	}

	schemaJSON, err := schema.ToObject(m.vm).MarshalJSON()
	if err != nil {
		throw(m.vm, "invalid JSON schema: %v", err)
	}
	valueJSON, err := value.ToObject(m.vm).MarshalJSON()
	if err != nil {
		throw(m.vm, "invalid JSON value: %v", err)
	}
	var valueGo any
	if err := json.Unmarshal(valueJSON, &valueGo); err != nil {
		throw(m.vm, "invalid JSON value: %v", err)
	}

	schemaGo, err := jsonschema.CompileString(url, string(schemaJSON))
	if err != nil {
		throw(m.vm, "invalid JSON schema: %v", err)
	}
	return schemaGo.Validate(valueGo)
}

func throw(runtime *goja.Runtime, format string, args ...any) {
	panic(runtime.NewGoError(fmt.Errorf(format, args...)))
}
