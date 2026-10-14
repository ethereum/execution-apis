package spec

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"
)

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonrpcError   `json:"error,omitempty"`
	Result  any             `json:"result,omitempty"`
}

type jsonrpcError struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Checker processes RPC messages and validates them against the spec.
type Checker struct {
	spec      *Spec
	calls     map[string]*MethodSchema
	history   map[string]struct{} // id of calls already performed
	checkArgs bool
}

func NewChecker(spec *Spec) *Checker {
	return &Checker{
		spec:    spec,
		calls:   make(map[string]*MethodSchema),
		history: make(map[string]struct{}),
	}
}

// SetCheckArgs enables or disables validation of method call arguments.
func (c *Checker) SetCheckArgs(on bool) {
	c.checkArgs = on
}

// Request checks a request message (sent by the client) against the spec.
func (c *Checker) Request(rawmsg json.RawMessage) error {
	var msg jsonrpcMessage
	if err := json.Unmarshal(rawmsg, &msg); err != nil {
		return fmt.Errorf("invalid call message: %v", err)
	}
	if msg.Method == "" {
		return fmt.Errorf(`missing "method" in call message`)
	}
	if len(msg.ID) == 0 {
		return fmt.Errorf(`missing "id" in call message`)
	}
	var arguments []any
	if len(msg.Params) > 0 {
		if err := json.Unmarshal(msg.Params, &arguments); err != nil {
			return fmt.Errorf(`invalid "params" in call message: %v`, err)
		}
	}

	// Track call.
	if _, ok := c.calls[string(msg.ID)]; ok {
		return fmt.Errorf("call with ID %s already in progress", msg.ID)
	}
	m, ok := c.spec.methods[msg.Method]
	if !ok {
		return fmt.Errorf("call with unknown method %s", msg.Method)
	}
	if c.checkArgs {
		if err := c.checkArguments(m, arguments); err != nil {
			return err
		}
	}
	// Store the call to active call map.
	c.calls[string(msg.ID)] = m
	return nil
}

// Response validates a response message from the server.
// For a response to a call which is currently in progress, this checks the
// call result against the response schema of the call's method.
//
// Note Response doesn't support batch requests yet.
func (c *Checker) Response(rawmsg json.RawMessage) error {
	var msg jsonrpcMessage
	err := json.Unmarshal(rawmsg, &msg)
	if err != nil {
		return fmt.Errorf("invalid response message: %v", err)
	}
	if msg.Method != "" {
		return fmt.Errorf("invalid response message: is method call %s", msg.Method)
	}

	// Find matching call.
	callmethod, ok := c.calls[string(msg.ID)]
	if !ok {
		return fmt.Errorf("invalid response message: no matching call for ID %s", msg.ID)
	}
	defer delete(c.calls, string(msg.ID))
	if msg.Error != nil {
		return c.checkError(callmethod, msg.Error)
	}
	url := fmt.Sprintf("#/methods/%d/result/schema", callmethod.index)
	return c.spec.Validate(url, msg.Result)
}

// UnansweredCalls returns the message IDs of all calls that haven't received a response.
func (c *Checker) UnansweredCalls() []string {
	return slices.SortedFunc(maps.Keys(c.calls), strings.Compare)
}

func (c *Checker) checkArguments(m *MethodSchema, arguments []any) error {
	// Validate each parameter value against their respective schema.
	for i, cd := range m.params {
		if len(arguments) <= i {
			if !cd.required {
				// skip missing optional values
				continue
			}
			return fmt.Errorf("missing required parameter %s.param[%d]", m.name, i)
		}
		url := fmt.Sprintf("#/methods/%d/params/%d/schema", m.index, i)
		if err := c.spec.Validate(url, arguments[i]); err != nil {
			return err
		}
	}
	return nil
}

// checkError verifies the error code is permitted by the schema.
// It also checks the error data against the schema.
func (c *Checker) checkError(m *MethodSchema, err *jsonrpcError) error {
	errSchema := m.errors[err.Code]
	if errSchema == nil {
		if isStandardError(err.Code) {
			return nil
		}
		return fmt.Errorf("response uses undefined error code %d for method %s", err.Code, m.name)
	}
	if !errSchema.hasDataSchema {
		if err.Data != nil {
			return errors.New("unexpected error.data field in response")
		}
	} else {
		if err.Data == nil {
			return errors.New("missing error.data field in response")
		}
		url := fmt.Sprintf("#/methods/%d/errors/%d/data", m.index, errSchema.index)
		return c.spec.Validate(url, err.Data)
	}
	return nil
}

// isStandardError says if `code' is an error code defined by the JSON-RPC 2.0 spec.
// The server may always respond with such errors, even if not explicitly defined by
// the OpenRPC specification.
func isStandardError(code int64) bool {
	switch code {
	case -32700 /* Parse Error */, -32600 /* Invalid Request */, -32602 /* Invalid Params */, -32603 /* Internal Error */ :
		return true
	case -32000 /* Server error */ :
		// This case should be removed. A lot of methods use this error code by default,
		// but it is meaningless, so it should be declared explicitly in the spec when it
		// is expected.
		return true
	default:
		return false
	}
}
