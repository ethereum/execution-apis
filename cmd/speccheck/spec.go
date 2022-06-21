package main

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/santhosh-tekuri/jsonschema/v5"
)

// MethodSchema stores all the schemas neccessary to validate a request or
// response corresponding to the method.
type MethodSchema struct {
	Params [][]byte
	Result []byte
}

// RoundTrip is a single round trip interaction between a certain JSON-RPC
// method.
type RoundTrip struct {
	Method   string
	Params   [][]byte
	Response []byte
}

// checkSpec reads the schemas from the spec and test files, then validates
// them against each other.
func checkSpec(args *Args) error {
	re, err := regexp.Compile(args.TestsRegex)
	if err != nil {
		return err
	}

	// Read all method schemas (params+result) from the OpenRPC spec.
	methods, err := parseMethodSchemas(args.SpecPath)
	if err != nil {
		return err
	}

	// Read all tests and parse out roundtrip HTTP exchanges so they can be validated.
	rts, err := parseRoundTrips(args.TestsRoot, re)
	if err != nil {
		return err
	}

	for _, rt := range rts {
		methodSchema, ok := methods[rt.Method]
		if !ok {
			return fmt.Errorf("undefined method: %s", rt.Method)
		}
		if len(methodSchema.Params) < len(rt.Params) {
			return fmt.Errorf("too many parameters")
		}
		// Validate each parameter value against their respective schema.
		for i, schema := range methodSchema.Params {
			// If there are not enough parameters in the method invocation, fail.
			if len(rt.Params) <= i {
				return fmt.Errorf("missing required param %d for method %s", i, rt.Method)
			}
			if err := validate(rt.Params[i], schema, fmt.Sprintf("%s.param[%d]", rt.Method, i)); err != nil {
				return err
			}
		}
		if err := validate(rt.Response, methodSchema.Result, fmt.Sprintf("%s.result", rt.Method)); err != nil {
			// Print out the value and schema if there is an error to further debug.
			var schema interface{}
			json.Unmarshal(methodSchema.Result, &schema)
			buf, _ := json.MarshalIndent(schema, "", "  ")
			fmt.Println(string(buf))
			fmt.Println(string(methodSchema.Result))
			fmt.Println(rt.Response)
			return err
		}
	}

	fmt.Println("all passing.")
	return nil
}

// validateParam validates the provided value against schema using the url base.
func validate(val []byte, schema []byte, url string) error {
	// Unmarshal value into interface{} so that validator can properly reflect
	// the contents.
	var x interface{}
	if err := json.Unmarshal(val, &x); err != nil {
		return err
	}
	s, err := jsonschema.CompileString(url, string(schema))
	if err != nil {
		return err
	}
	if err := s.Validate(x); err != nil {
		return err
	}
	return nil
}
