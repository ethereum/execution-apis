package main

import (
	"encoding/json"
	"fmt"
	"os"

	openrpc "github.com/open-rpc/meta-schema"
)

type ContentDescriptor struct {
	name     string
	required bool
	schema   openrpc.JSONSchemaObject
}

// methodSchema stores all the schemas neccessary to validate a request or
// response corresponding to the method.
type methodSchema struct {
	name   string
	params []*ContentDescriptor
	result *ContentDescriptor
}

// parseSpec reads an OpenRPC specification and parses out each
// method's schemas.
func parseSpec(filename string) (map[string]*methodSchema, error) {
	doc, err := readSpec(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read spec: %v", err)
	}

	// Iterate over each method in the OpenRPC spec and pull out the parameter
	// schema and result schema.
	parsed := make(map[string]*methodSchema)
	for _, method := range *doc.Methods {
		if method.ReferenceObject != nil {
			return nil, fmt.Errorf("reference object not supported, %s", *method.ReferenceObject.Ref)
		}
		var (
			method = method.MethodObject
			ms     = methodSchema{name: string(*method.Name)}
		)
		// Add parameter schemas.
		for i, param := range *method.Params {
			if err := checkCDOR(param); err != nil {
				return nil, fmt.Errorf("%s, parameter %d: %v", *method.Name, i, err)
			}
			required := false
			if param.ContentDescriptorObject.Required != nil && *param.ContentDescriptorObject.Required {
				required = true
			}
			cd := &ContentDescriptor{
				name:     string(*param.ContentDescriptorObject.Name),
				required: required,
				schema:   *param.ContentDescriptorObject.Schema.JSONSchemaObject,
			}
			ms.params = append(ms.params, cd)
		}

		// Add result schema.
		if method.Result == nil {
			return nil, fmt.Errorf("%s: missing result", *method.Name)
		}
		cdor := openrpc.ContentDescriptorOrReference{
			ContentDescriptorObject: method.Result.ContentDescriptorObject,
			ReferenceObject:         method.Result.ReferenceObject,
		}
		if err := checkCDOR(cdor); err != nil {
			return nil, fmt.Errorf("%s: %v", *method.Name, err)
		}
		obj := method.Result.ContentDescriptorObject
		required := false
		if obj.Required != nil && *obj.Required {
			required = true
		}
		ms.result = &ContentDescriptor{
			name:     string(*obj.Name),
			required: required,
			schema:   *obj.Schema.JSONSchemaObject,
		}
		parsed[string(*method.Name)] = &ms
	}

	return parsed, nil
}

// parseParamValues parses each parameter out of the raw json value in its own byte
// slice.
func parseParamValues(raw json.RawMessage) ([][]byte, error) {
	if len(raw) == 0 {
		return [][]byte{}, nil
	}
	var params []interface{}
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil, err
	}
	// Iterate over top-level parameter values and re-marshal them to get a
	// list of json-encoded parameter values.
	var out [][]byte
	for _, param := range params {
		buf, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		out = append(out, buf)
	}
	return out, nil
}

func checkCDOR(obj openrpc.ContentDescriptorOrReference) error {
	if obj.ReferenceObject != nil {
		return fmt.Errorf("references not supported")
	}
	if obj.ContentDescriptorObject == nil {
		return fmt.Errorf("missing content descriptor")
	}
	cd := obj.ContentDescriptorObject
	if cd.Name == nil {
		return fmt.Errorf("missing name")
	}
	if cd.Schema == nil || cd.Schema.JSONSchemaObject == nil {
		return fmt.Errorf("missing schema")
	}
	return nil
}

func readSpec(path string) (*openrpc.OpenrpcDocument, error) {
	spec, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc openrpc.OpenrpcDocument
	if err := json.Unmarshal(spec, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}
