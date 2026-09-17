package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/ethereum/execution-apis/tools/iofile"
	openrpc "github.com/open-rpc/spec-types/generated/packages/go/v1_4"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// checkSpec reads the schemas from the spec and test files, then validates
// them against each other.
func checkSpec(methods map[string]*methodSchema, tests []iofile.Test, re *regexp.Regexp) error {
	for _, test := range tests {
		calls, err := testCalls(test)
		if err != nil {
			return fmt.Errorf("%s: %v", test.Name, err)
		}
		for _, call := range calls {
			if err := checkCall(methods, call, re); err != nil {
				return fmt.Errorf("%s: %v", test.Name, err)
			}
		}
	}
	fmt.Println("all passing.")
	return nil
}

func checkCall(methods map[string]*methodSchema, rt *methodCall, re *regexp.Regexp) error {
	method, ok := methods[rt.method]
	if !ok {
		return fmt.Errorf("method %s is not defined in schema", rt.method)
	}
	// skip validator of test if name includes "invalid" as the schema
	// doesn't yet support it.
	// TODO(matt): create error schemas.
	if strings.Contains(rt.name, "invalid") {
		return nil
	}
	if len(method.params) < len(rt.params) {
		return fmt.Errorf("%s: too many parameters", method.name)
	}
	// Validate each parameter value against their respective schema.
	for i, cd := range method.params {
		if len(rt.params) <= i {
			if !cd.required {
				// skip missing optional values
				continue
			}
			return fmt.Errorf("missing required parameter %s.param[%d]", rt.method, i)
		}
		if err := validate(&method.params[i].schema, rt.params[i], fmt.Sprintf("%s.param[%d]", rt.method, i)); err != nil {
			return fmt.Errorf("unable to validate parameter: %s", err)
		}
	}
	if rt.response.Result == nil && rt.response.Error != nil {
		// skip validation of errors, they haven't been standardized
		return nil
	}
	if err := validate(&method.result.schema, rt.response.Result, fmt.Sprintf("%s.result", rt.method)); err != nil {
		// Print out the value and schema if there is an error to further debug.
		buf, _ := json.Marshal(method.result.schema)
		fmt.Println(string(buf))
		fmt.Println(string(rt.response.Result))
		fmt.Println()
		return fmt.Errorf("invalid result %#v", err)
	}
	return nil
}

// validateParam validates the provided value against schema using the url base.
func validate(schema *openrpc.JSONSchemaObject, val []byte, url string) error {
	// Set $schema explicitly to force jsonschema to use draft 2019-09.
	draft := openrpc.Schema("https://json-schema.org/draft/2019-09/schema")
	schema.Schema = &draft

	// Compile schema.
	b, err := json.Marshal(schema)
	if err != nil {
		return fmt.Errorf("unable to marshal schema to json")
	}
	s, err := jsonschema.CompileString(url, string(b))
	if err != nil {
		return err
	}

	// Validate value
	var x any
	json.Unmarshal(val, &x)
	if err := s.Validate(x); err != nil {
		return err
	}
	return nil
}

type methodCall struct {
	method   string
	name     string
	params   [][]byte
	response *jsonrpcMessage
}

type jsonrpcMessage struct {
	Version string          `json:"jsonrpc,omitempty"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *jsonError      `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

type jsonError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func testCalls(test iofile.Test) ([]*methodCall, error) {
	var calls []*methodCall
	var req *jsonrpcMessage
	for _, msg := range test.Messages {
		if msg.Send {
			req = new(jsonrpcMessage)
			if err := json.Unmarshal(msg.Data, req); err != nil {
				return nil, err
			}
		} else {
			if req == nil {
				return nil, fmt.Errorf("response w/o corresponding request")
			}
			var resp jsonrpcMessage
			if err := json.Unmarshal(msg.Data, &resp); err != nil {
				return nil, err
			}
			// Parse parameters into slice of string.
			params, err := parseParamValues(req.Params)
			if err != nil {
				return nil, fmt.Errorf("unable to parse params: %s %v", err, req.Params)
			}
			calls = append(calls, &methodCall{req.Method, test.Name, params, &resp})
			req = nil
		}
	}
	if req != nil {
		return nil, fmt.Errorf("unhandled request")
	}
	return calls, nil
}
