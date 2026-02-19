package specgen

import "fmt"

// MergeAllOf recursively walks schema, merging every allOf it encounters into
// its containing object. It returns a deep copy with all allOf keys removed.
// The input schema is not modified.
//
// Merge rules applied at each allOf site:
//   - properties: union; the parent object wins on key conflict.
//   - required:   concatenated and deduplicated.
//   - all other fields: parent object wins.
//
// The input must not contain any $ref entries; if one is found MergeAllOf panics.
func MergeAllOf(schema object) object {
	return mergeObject(schema)
}

func mergeValue(v any) any {
	switch val := v.(type) {
	case map[string]any:
		return mergeObject(val)
	case []any:
		return mergeSlice(val)
	default:
		return v
	}
}

func mergeSlice(arr []any) []any {
	out := make([]any, len(arr))
	for i, item := range arr {
		out[i] = mergeValue(item)
	}
	return out
}

func mergeObject(obj object) object {
	if ref, ok := obj["$ref"].(string); ok {
		panic(fmt.Sprintf("mergeObject: unexpected $ref %q (dereference input first)", ref))
	}

	// Deep-copy and recurse into all non-allOf values first.
	out := make(object, len(obj))
	for k, v := range obj {
		if k == "allOf" {
			continue // handled below
		}
		out[k] = mergeValue(v)
	}

	// Merge allOf entries into out, if present.
	if allOf, ok := obj["allOf"].([]any); ok {
		for i, entry := range allOf {
			sub, ok := entry.(map[string]any)
			if !ok {
				panic(fmt.Sprintf("mergeObject: allOf[%d]: unexpected type %T", i, entry))
			}
			mergeSchemas(out, mergeObject(sub))
		}
	}

	return out
}

// mergeSchemas merges fields from src into dst.
//   - properties: union; dst wins on key conflict.
//   - required: concatenated and deduplicated.
//   - all other fields: dst wins (src value only copied when dst key is absent).
func mergeSchemas(dst, src object) {
	// Merge properties.
	if srcProps, ok := src["properties"].(map[string]any); ok {
		if dst["properties"] == nil {
			dst["properties"] = make(map[string]any)
		}
		dstProps := dst["properties"].(map[string]any)
		for k, v := range srcProps {
			if _, exists := dstProps[k]; !exists {
				dstProps[k] = v
			}
		}
	}
	// Merge required (deduplicated).
	if srcReq, ok := src["required"].([]any); ok {
		existing := make(map[string]bool)
		dstReq, _ := dst["required"].([]any)
		for _, r := range dstReq {
			if s, ok := r.(string); ok {
				existing[s] = true
			}
		}
		for _, r := range srcReq {
			if s, ok := r.(string); ok && !existing[s] {
				dstReq = append(dstReq, s)
				existing[s] = true
			}
		}
		if len(dstReq) > 0 {
			dst["required"] = dstReq
		}
	}
	// All other fields: dst wins.
	for k, v := range src {
		if k == "properties" || k == "required" {
			continue
		}
		if _, hasDst := dst[k]; !hasDst {
			dst[k] = v
		}
	}
}
