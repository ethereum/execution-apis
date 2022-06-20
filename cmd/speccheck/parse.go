package main

import (
	"encoding/json"
	"fmt"
	"strings"

	openrpc "github.com/open-rpc/meta-schema"
)

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

func parseTest(data []byte) ([]*RoundTrip, error) {
	rts := make([]*RoundTrip, 0)
	var req *jsonrpcMessage
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		switch {
		case len(line) == 0 || strings.HasPrefix(line, "//"):
			// Skip comments, blank lines.
			continue
		case strings.HasPrefix(line, ">> "):
			req = &jsonrpcMessage{}
			if err := json.Unmarshal([]byte(line[3:]), &req); err != nil {
				return nil, err
			}
		case strings.HasPrefix(line, "<< "):
			if req == nil {
				return nil, fmt.Errorf("response w/o corresponding request")
			}
			var resp jsonrpcMessage
			if err := json.Unmarshal([]byte(line[3:]), &resp); err != nil {
				return nil, err
			}
			// Parse parameters into slice of string.
			params, err := parseParams(req.Params)
			if err != nil {
				return nil, fmt.Errorf("unable to parse params: %s %v", err, req.Params)
			}
			rts = append(rts, &RoundTrip{req.Method, params, string(resp.Result)})
			req = nil
		default:
			return nil, fmt.Errorf("invalid line in test: %s", line)
		}
	}
	if req != nil {
		return nil, fmt.Errorf("unhandled request")
	}
	return rts, nil
}

func parseParams(params []byte) ([]string, error) {
	if len(params) == 0 {
		return []string{}, nil
	}
	var abstractParams []interface{}
	if err := json.Unmarshal(params, &abstractParams); err != nil {
		return nil, err
	}
	var out []string
	for _, param := range abstractParams {
		buf, err := json.Marshal(param)
		if err != nil {
			return nil, err
		}
		out = append(out, string(buf))
	}
	return out, nil
}

func parseMethods(spec []byte) (map[string]*MethodSchema, error) {
	var doc openrpc.OpenrpcDocument
	if err := json.Unmarshal(spec, &doc); err != nil {
		fmt.Println(err)
	}

	// Iterate over each method in the openrpc spec and pull out the parameter
	// schema and result schema.
	parsed := make(map[string]*MethodSchema)
	for _, method := range *doc.Methods {
		var schema MethodSchema

		// Read parameter schemas.
		for _, param := range *method.MethodObject.Params {
			if param.ReferenceObject != nil {
				return nil, fmt.Errorf("parameter references not supported")
			}
			buf, err := json.Marshal(param.ContentDescriptorObject.Schema.JSONSchemaObject)
			if err != nil {
				return nil, err
			}
			schema.Params = append(schema.Params, buf)
		}

		// Read result schema.
		buf, err := json.Marshal(method.MethodObject.Result.ContentDescriptorObject.Schema)
		if err != nil {
			return nil, err
		}
		schema.Result = buf
		parsed[string(*method.MethodObject.Name)] = &schema
	}

	return parsed, nil
}
