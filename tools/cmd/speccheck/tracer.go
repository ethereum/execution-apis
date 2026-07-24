package main

import (
	"encoding/json"
	"fmt"
	"strings"

	openrpc "github.com/open-rpc/spec-types/generated/packages/go/v1_4"
)

// tracerConfigParamIndex maps debug tracing methods to the position of their
// TraceConfig parameter.
var tracerConfigParamIndex = map[string]int{
	"debug_traceTransaction":   1,
	"debug_traceBlockByNumber": 1,
	"debug_traceBlockByHash":   1,
	"debug_traceCall":          2,
}

// tracerBranchTitles maps a requested tracer name to the title prefix of the
// result-schema anyOf branch that defines its output. The empty name is the
// opcode (struct) logger.
var tracerBranchTitles = map[string]string{
	"":           "Opcode tracer",
	"callTracer": "Call tracer",
}

// selectTracerSchema narrows a debug_trace* result schema to the anyOf branch
// matching the tracer requested by the fixture. Without this, the unconstrained
// named-tracer branch accepts any value and the whole-schema validation is
// vacuous. The narrowed schema is returned as a raw map (not the typed struct):
// the generated union types do not round-trip — re-marshaling an already
// normalized "type" array double-wraps it into invalid schema JSON.
//
// Returns (nil, nil) when the method or tracer is not recognized and the
// caller should validate against the full schema. Returns an error when the
// tracer is recognized but no matching branch exists — falling back silently
// would reopen the vacuous validation this narrowing exists to prevent.
func selectTracerSchema(schema *openrpc.JSONSchemaObject, method string, params [][]byte) (map[string]interface{}, error) {
	idx, ok := tracerConfigParamIndex[method]
	if !ok {
		return nil, nil
	}
	tracer := ""
	if len(params) > idx {
		var cfg struct {
			Tracer string `json:"tracer"`
		}
		if err := json.Unmarshal(params[idx], &cfg); err == nil {
			tracer = cfg.Tracer
		}
	}
	wantTitle, ok := tracerBranchTitles[tracer]
	if !ok {
		return nil, nil
	}

	raw, err := schemaToMap(schema)
	if err != nil {
		return nil, err
	}
	if branch := pickBranch(raw, wantTitle); branch != nil {
		return branch, nil
	}
	if items := itemsSchema(raw["items"]); items != nil {
		if branch := pickBranch(items, wantTitle); branch != nil {
			narrowed := map[string]interface{}{"type": "array", "items": branch}
			if title, ok := raw["title"]; ok {
				narrowed["title"] = title
			}
			return narrowed, nil
		}
	}
	return nil, fmt.Errorf("%s: no %q branch in result schema for tracer %q", method, wantTitle, tracer)
}

// itemsSchema unwraps the items value, which the openrpc Items union type
// marshals either as a single schema object or a one-element array.
func itemsSchema(v interface{}) map[string]interface{} {
	switch items := v.(type) {
	case map[string]interface{}:
		return items
	case []interface{}:
		if len(items) == 1 {
			if m, ok := items[0].(map[string]interface{}); ok {
				return m
			}
		}
	}
	return nil
}

// pickBranch returns the anyOf branch of node whose title starts with wantTitle.
func pickBranch(node map[string]interface{}, wantTitle string) map[string]interface{} {
	branches, ok := node["anyOf"].([]interface{})
	if !ok {
		return nil
	}
	for _, b := range branches {
		branch, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		title, _ := branch["title"].(string)
		if strings.HasPrefix(title, wantTitle) {
			return branch
		}
	}
	return nil
}

// schemaToMap converts a typed schema to a raw map via one JSON round-trip.
func schemaToMap(schema *openrpc.JSONSchemaObject) (map[string]interface{}, error) {
	b, err := json.Marshal(schema)
	if err != nil {
		return nil, err
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return nil, err
	}
	return raw, nil
}
