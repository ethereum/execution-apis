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
func checkSpec(args *Args) error {
	re, err := regexp.Compile(args.TestsRegex)
	if err != nil {
		return err
	}

	// Read all method schemas (params+result) from the OpenRPC spec.
	methods, err := parseSpec(args.SpecPath)
	if err != nil {
		return err
	}

	// Read all tests and parse out roundtrip HTTP exchanges so they can be validated.
	rts, err := readRtts(args.TestsRoot, re)
	if err != nil {
		return err
	}

	for _, rt := range rts {
		fmt.Println(rt.name)
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
			return fmt.Errorf("too many parameters")
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
		if err := validate(&method.result.schema, rt.response, fmt.Sprintf("%s.result", rt.method)); err != nil {
			// Print out the value and schema if there is an error to further debug.
			buf, _ := json.Marshal(method.result.schema)
			fmt.Println(string(buf))
			fmt.Println(string(rt.response))
			return fmt.Errorf("invalid result %s: %w", rt.name, err)
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
	fmt.Println(string(b))
	s, err := jsonschema.CompileString(url, string(b))
	if err != nil {
		return fmt.Errorf("unable to compile schema: %w", err)
	}

	// Validate value
	var x interface{}
	json.Unmarshal(val, &x)
	fmt.Println(x)
	if err := s.Validate(x); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}
