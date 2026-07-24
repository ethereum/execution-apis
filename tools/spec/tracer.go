package spec

import (
	"fmt"
	"strings"
)

var tracerConfigParamIndex = map[string]int{
	"debug_traceTransaction":   1,
	"debug_traceBlockByNumber": 1,
	"debug_traceBlockByHash":   1,
	"debug_traceCall":          2,
}

// The empty tracer name is the opcode (struct) logger.
var tracerBranchTitles = map[string]string{
	"":           "Opcode tracer",
	"callTracer": "Call tracer",
}

// resultSchemaURL narrows a debug_trace* result schema to the anyOf branch of the
// requested tracer. The named-tracer branch accepts any value, so the whole anyOf
// would accept any result; for the same reason a missing branch is an error.
func (s *Spec) resultSchemaURL(m *methodSchema, arguments []any) (url, itemURL string, err error) {
	url = fmt.Sprintf("#/methods/%d/result/schema", m.index)
	idx, ok := tracerConfigParamIndex[m.name]
	if !ok {
		return url, "", nil
	}
	tracer := ""
	if len(arguments) > idx {
		if cfg, ok := arguments[idx].(map[string]any); ok {
			tracer, _ = cfg["tracer"].(string)
		}
	}
	wantTitle, ok := tracerBranchTitles[tracer]
	if !ok {
		return url, "", nil
	}
	schema := s.resultSchema(m)
	if i := pickBranch(schema, wantTitle); i >= 0 {
		return fmt.Sprintf("%s/anyOf/%d", url, i), "", nil
	}
	switch items := schema["items"].(type) {
	case map[string]any:
		if i := pickBranch(items, wantTitle); i >= 0 {
			return url, fmt.Sprintf("%s/items/anyOf/%d", url, i), nil
		}
	case []any:
		if len(items) == 1 {
			item, _ := items[0].(map[string]any)
			if i := pickBranch(item, wantTitle); i >= 0 {
				return url, fmt.Sprintf("%s/items/0/anyOf/%d", url, i), nil
			}
		}
	}
	return "", "", fmt.Errorf("%s: no %q branch in result schema for tracer %q", m.name, wantTitle, tracer)
}

func (s *Spec) resultSchema(m *methodSchema) map[string]any {
	doc, _ := s.cdoc.(map[string]any)
	methods, _ := doc["methods"].([]any)
	if m.index >= len(methods) {
		return nil
	}
	method, _ := methods[m.index].(map[string]any)
	result, _ := method["result"].(map[string]any)
	schema, _ := result["schema"].(map[string]any)
	return schema
}

func pickBranch(node map[string]any, wantTitle string) int {
	branches, _ := node["anyOf"].([]any)
	for i, b := range branches {
		branch, _ := b.(map[string]any)
		title, _ := branch["title"].(string)
		if strings.HasPrefix(title, wantTitle) {
			return i
		}
	}
	return -1
}
