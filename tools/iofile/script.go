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
)

type scriptMessage struct {
	Send     json.RawMessage `json:"send,omitempty"`
	Expected json.RawMessage `json:"expected,omitempty"`
	Response json.RawMessage `json:"response,omitempty"`
}

// ScriptConfig holds settings for validation script execution.
type ScriptConfig struct {
	Log     TestLogger
	Timeout time.Duration
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

// Receives returns the data of all receive (<<) lines.
func (t *Test) Receives() (m []json.RawMessage) {
	for _, msg := range t.Messages {
		if !msg.Send {
			m = append(m, msg.Data)
		}
	}
	return m
}

var errExecutionTimeout = errors.New("script execution timeout reached")

func runValidationScript(sourceFile string, source string, jsonMessages []byte, config ScriptConfig) error {
	vm := goja.New()

	// Enable console output to w.
	reg := new(require.Registry)
	reg.Enable(vm)
	reg.RegisterNativeModule(console.ModuleName, console.RequireWithPrinter(scriptPrinter{config.Log}))
	console.Enable(vm)

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
