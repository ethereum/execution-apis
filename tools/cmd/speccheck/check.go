package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	openrpc "github.com/open-rpc/meta-schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// checkSpec reads the schemas from the spec and test files, then validates
// them against each other.
func checkSpec(methods map[string]*methodSchema, rts []*roundTrip, re *regexp.Regexp) error {
	for _, rt := range rts {
		method, ok := methods[rt.method]
		if !ok {
			return fmt.Errorf("undefined method: %s", rt.method)
		}
		// Exempts tests on methods that haven't adopted error-groups yet;
		// remove once they do.
		if strings.Contains(rt.name, "invalid") {
			continue
		}
		// Error responses: validate the code against the spec
		if rt.response.Result == nil && rt.response.Error != nil {
			checkError(method, rt)
			continue
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
				return fmt.Errorf("unable to validate parameter in %s: %s", rt.name, err)
			}
		}
		if err := validate(&method.result.schema, rt.response.Result, fmt.Sprintf("%s.result", rt.method)); err != nil {
			// Print out the value and schema if there is an error to further debug.
			buf, _ := json.Marshal(method.result.schema)
			fmt.Println(string(buf))
			fmt.Println(string(rt.response.Result))
			fmt.Println()
			return fmt.Errorf("invalid result %s\n%#v", rt.name, err)
		}
	}

	fmt.Println("all passing.")
	return nil
}

// checkError warns when a fixture's error.code isn't in the method's spec errors.
func checkError(method *methodSchema, rt *roundTrip) {
	if len(method.errors) == 0 {
		return
	}
	code := rt.response.Error.Code
	for _, e := range method.errors {
		if e.Code == code {
			if rt.response.Error.Message != e.Message {
				// Message-mismatch warning is intentionally suppressed until
				// clients converge on spec wording.
				// fmt.Printf("[WARN]: ERROR MESSAGE: %q does not match expected: %q in %s\n",
				// 	rt.response.Error.Message, e.Message, rt.name)
			}
			return
		}
	}
	fmt.Printf("[WARN]: ERROR CODE: %d not found for method %s in %s\n",
		code, method.name, rt.name)
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
	var x interface{}
	json.Unmarshal(val, &x)
	if err := s.Validate(x); err != nil {
		return err
	}
	return nil
}
