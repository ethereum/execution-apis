package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type ContentDescriptor struct {
	name     string
	required bool
	schema   json.RawMessage
}

// methodSchema stores all the schemas necessary to validate a request or
// response corresponding to the method.
type methodSchema struct {
	name   string
	params []*ContentDescriptor
	result *ContentDescriptor
}

// Keep schemas as raw JSON. The generated OpenRPC union types change `items`
// (both homogeneous arrays and tuples) when marshaled back to JSON.
type specDescriptor struct {
	Name     *string         `json:"name"`
	Required bool            `json:"required"`
	Schema   json.RawMessage `json:"schema"`
	Ref      *string         `json:"$ref"`
}

// parseSpec reads method descriptors without reserializing their schemas.
func parseSpec(filename string) (map[string]*methodSchema, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read spec: %w", err)
	}
	var doc struct {
		Methods []struct {
			Name   *string           `json:"name"`
			Ref    *string           `json:"$ref"`
			Params []*specDescriptor `json:"params"`
			Result *specDescriptor   `json:"result"`
		} `json:"methods"`
	}
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("unable to read spec: %w", err)
	}
	parsed := make(map[string]*methodSchema)
	for _, method := range doc.Methods {
		if method.Ref != nil {
			return nil, fmt.Errorf("reference object not supported, %s", *method.Ref)
		}
		if method.Name == nil {
			return nil, fmt.Errorf("missing method name")
		}
		ms := &methodSchema{name: *method.Name}
		for i, param := range method.Params {
			cd, err := parseDescriptor(param)
			if err != nil {
				return nil, fmt.Errorf("%s, parameter %d: %w", ms.name, i, err)
			}
			ms.params = append(ms.params, cd)
		}
		if method.Result == nil {
			return nil, fmt.Errorf("%s: missing result", ms.name)
		}
		ms.result, err = parseDescriptor(method.Result)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", ms.name, err)
		}
		parsed[ms.name] = ms
	}
	return parsed, nil
}

func parseDescriptor(cd *specDescriptor) (*ContentDescriptor, error) {
	if cd == nil {
		return nil, fmt.Errorf("missing content descriptor")
	}
	if cd.Ref != nil {
		return nil, fmt.Errorf("references not supported")
	}
	if cd.Name == nil {
		return nil, fmt.Errorf("missing name")
	}
	var schema map[string]json.RawMessage
	if err := json.Unmarshal(cd.Schema, &schema); err != nil || schema == nil {
		return nil, fmt.Errorf("missing or invalid schema object")
	}
	return &ContentDescriptor{name: *cd.Name, required: cd.Required, schema: cd.Schema}, nil
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
