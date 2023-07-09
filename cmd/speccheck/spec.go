package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	openrpc "github.com/open-rpc/meta-schema"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

// methodSchema stores all the schemas neccessary to validate a request or
// response corresponding to the method.
type methodSchema struct {
	// Schemas
	params []openrpc.ContentDescriptorObject
	result []byte
}

// roundTrip is a single round trip interaction between a certain JSON-RPC
// method.
type roundTrip struct {
	method   string
	name     string
	params   [][]byte
	response []byte
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
		methodSchema, ok := methods[rt.method]
		if !ok {
			return fmt.Errorf("undefined method: %s", rt.method)
		}
		// skip validator of test if name includes "invalid" as the schema
		// doesn't yet support it.
		// TODO(matt): create error schemas.
		if strings.Contains(rt.name, "invalid") {
			continue
		}
		if len(methodSchema.params) < len(rt.params) {
			return fmt.Errorf("too many parameters")
		}
		// Validate each parameter value against their respective schema.
		for i, schema := range methodSchema.params {
			if len(rt.params) <= i {
				if schema.Required == nil || !(*schema.Required) {
					// skip missing optional values
					continue
				}
				return fmt.Errorf("missing required parameter %s.param[%d]", rt.method, i)
			}
			raw, err := json.Marshal(schema.Schema.JSONSchemaObject)
			if err != nil {
				return err
			}
			if err := validate(rt.params[i], raw, fmt.Sprintf("%s.param[%d]", rt.method, i)); err != nil {
				return fmt.Errorf("unable to validate parameter: %s", err)
			}
		}
		if err := validate(rt.response, methodSchema.result, fmt.Sprintf("%s.result", rt.method)); err != nil {
			// Print out the value and schema if there is an error to further debug.
			var schema interface{}
			json.Unmarshal(methodSchema.result, &schema)
			buf, _ := json.MarshalIndent(schema, "", "  ")
			fmt.Println(string(buf))
			fmt.Println(string(methodSchema.result))
			fmt.Println(string(rt.response))
			return fmt.Errorf("invalid result %s: %w", rt.name, err)
		}
	}

	fmt.Println("all passing.")
	return nil
}

// validateParam validates the provided value against schema using the url base.
func validate(val []byte, baseSchema []byte, url string) error {
	// Unmarshal value into interface{} so that validator can properly reflect
	// the contents.
	var x interface{}
	if err := json.Unmarshal(val, &x); len(val) != 0 && err != nil {
		return fmt.Errorf("unable to unmarshal testcase: %w", err)
	}
	// Add $schema explicitly to force jsonschema to use draft 2019-09.
	schema, err := appendDraft201909(baseSchema)
	if err != nil {
		return fmt.Errorf("unable to append draft: %w", err)
	}
	s, err := jsonschema.CompileString(url, string(schema))
	if err != nil {
		return fmt.Errorf("unable to compile schema: %w", err)
	}
	if err := s.Validate(x); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	return nil
}

// appendDraft201909 adds $schema = draft 2019-09 to the schema.
func appendDraft201909(schema []byte) ([]byte, error) {
	var out map[string]interface{}
	if err := json.Unmarshal(schema, &out); err != nil {
		return nil, err
	}
	out["$schema"] = "https://json-schema.org/draft/2019-09/schema"
	return json.Marshal(out)
}
