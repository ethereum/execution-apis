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
		// skip validator of test if name includes "invalid" as the schema
		// doesn't yet support it.
		// TODO(matt): create error schemas.
		if strings.Contains(rt.name, "invalid") {
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
		if rt.response.Result == nil && rt.response.Error != nil {
			// skip validation of errors, they haven't been standardized
			continue
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
