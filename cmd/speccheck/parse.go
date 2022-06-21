package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
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

// parseRoundTrips walks a root directory and parses round trip HTTP exchanges
// from files that match the regular expression.
func parseRoundTrips(root string, re *regexp.Regexp) ([]*RoundTrip, error) {
	rts := make([]*RoundTrip, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("unable to walk path: %s\n", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		if fname := info.Name(); !strings.HasSuffix(fname, ".io") {
			return nil
		}
		pathname := strings.TrimSuffix(strings.TrimPrefix(path, root), ".io")
		if !re.MatchString(pathname) {
			fmt.Println("skip", pathname)
			return nil // skip
		}
		// Found a good test, parse it and append to list.
		test, err := parseTest(path)
		if err != nil {
			return err
		}
		rts = append(rts, test...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rts, nil
}

// parseTest parses a single test into a slice of HTTP round trips.
func parseTest(filename string) ([]*RoundTrip, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
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
			params, err := parseParamValues(req.Params)
			if err != nil {
				return nil, fmt.Errorf("unable to parse params: %s %v", err, req.Params)
			}
			rts = append(rts, &RoundTrip{req.Method, params, resp.Result})
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

// parseMethodSchemas reads an OpenRPC specification and parses out each
// method's schemas.
func parseMethodSchemas(filename string) (map[string]*MethodSchema, error) {
	spec, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var doc openrpc.OpenrpcDocument
	if err := json.Unmarshal(spec, &doc); err != nil {
		return nil, err
	}
	// Iterate over each method in the OpenRPC spec and pull out the parameter
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
