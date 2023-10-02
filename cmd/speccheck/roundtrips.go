package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
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

// roundTrip is a single round trip interaction between a certain JSON-RPC
// method.
type roundTrip struct {
	method   string
	name     string
	params   [][]byte
	response []byte
}

// readRtts walks a root directory and parses round trip HTTP exchanges
// from files that match the regular expression.
func readRtts(root string, re *regexp.Regexp) ([]*roundTrip, error) {
	rts := make([]*roundTrip, 0)
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
		test, err := readTest(pathname, path)
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

// readTest reads a single test into a slice of HTTP round trips.
func readTest(testname string, filename string) ([]*roundTrip, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	rts := make([]*roundTrip, 0)
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
			rts = append(rts, &roundTrip{req.Method, testname, params, resp.Result})
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
