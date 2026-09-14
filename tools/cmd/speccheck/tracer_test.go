package main

import (
	"encoding/json"
	"testing"

	openrpc "github.com/open-rpc/spec-types/generated/packages/go/v1_4"
)

func schemaFromJSON(t *testing.T, s string) *openrpc.JSONSchemaObject {
	t.Helper()
	var out openrpc.JSONSchemaObject
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatal(err)
	}
	return &out
}

const txResultSchema = `{
  "anyOf": [
    {"title": "Opcode tracer result", "type": "object", "required": ["structLogs"]},
    {"title": "Call tracer result", "type": "object", "required": ["type", "from"]},
    {"title": "Named tracer result", "description": "any JSON value"}
  ]
}`

const blockResultSchema = `{
  "title": "Array of per-transaction traces",
  "type": "array",
  "items": {
    "anyOf": [
      {"title": "Opcode tracer entry", "type": "object", "required": ["txHash", "result"]},
      {"title": "Call tracer entry", "type": "object", "required": ["txHash"],
       "anyOf": [{"required": ["result"]}, {"required": ["error"]}]},
      {"title": "Named tracer entry", "type": "object"}
    ]
  }
}`

func TestSelectTracerSchema_CallTracer(t *testing.T) {
	schema := schemaFromJSON(t, txResultSchema)
	params := [][]byte{[]byte(`"0xabc"`), []byte(`{"tracer":"callTracer"}`)}
	got, err := selectTracerSchema(schema, "debug_traceTransaction", params)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got["title"] != "Call tracer result" {
		t.Errorf("want Call tracer result branch, got %v", got)
	}
}

func TestSelectTracerSchema_NoTracerSelectsOpcode(t *testing.T) {
	schema := schemaFromJSON(t, txResultSchema)
	params := [][]byte{[]byte(`"0xabc"`)}
	got, err := selectTracerSchema(schema, "debug_traceTransaction", params)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got["title"] != "Opcode tracer result" {
		t.Errorf("want Opcode tracer result branch, got %v", got)
	}
}

func TestSelectTracerSchema_UnknownTracerKeepsWholeSchema(t *testing.T) {
	schema := schemaFromJSON(t, txResultSchema)
	params := [][]byte{[]byte(`"0xabc"`), []byte(`{"tracer":"prestateTracer"}`)}
	got, err := selectTracerSchema(schema, "debug_traceTransaction", params)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("unknown tracer should keep the whole schema, got %v", got)
	}
}

func TestSelectTracerSchema_MissingBranchErrors(t *testing.T) {
	schema := schemaFromJSON(t, `{"anyOf": [
		{"title": "Opcode tracer result", "type": "object"},
		{"title": "Named tracer result"}
	]}`)
	params := [][]byte{[]byte(`"0xabc"`), []byte(`{"tracer":"callTracer"}`)}
	if _, err := selectTracerSchema(schema, "debug_traceTransaction", params); err == nil {
		t.Fatal("expected error when the tracer is known but no branch matches")
	}
}

func TestSelectTracerSchema_BlockItemsNarrowed(t *testing.T) {
	schema := schemaFromJSON(t, blockResultSchema)
	params := [][]byte{[]byte(`"0x1"`), []byte(`{"tracer":"callTracer"}`)}
	got, err := selectTracerSchema(schema, "debug_traceBlockByNumber", params)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected narrowed schema")
	}
	items := itemsSchema(got["items"])
	if items == nil {
		t.Fatalf("narrowed schema has no items: %v", got)
	}
	if items["title"] != "Call tracer entry" {
		t.Errorf("want Call tracer entry items, got %v", items["title"])
	}
	if got["type"] != "array" {
		t.Errorf("narrowed schema should stay an array, got %v", got["type"])
	}
}

func TestSelectTracerSchema_TraceCallParamIndex(t *testing.T) {
	schema := schemaFromJSON(t, txResultSchema)
	params := [][]byte{[]byte(`{}`), []byte(`"latest"`), []byte(`{"tracer":"callTracer"}`)}
	got, err := selectTracerSchema(schema, "debug_traceCall", params)
	if err != nil {
		t.Fatal(err)
	}
	if got == nil || got["title"] != "Call tracer result" {
		t.Errorf("want Call tracer result branch, got %v", got)
	}
}

func TestSelectTracerSchema_NonTraceMethodUntouched(t *testing.T) {
	schema := schemaFromJSON(t, txResultSchema)
	got, err := selectTracerSchema(schema, "eth_call", nil)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("non-trace methods should keep the whole schema, got %v", got)
	}
}

func TestValidateRawNarrowedBranch(t *testing.T) {
	schema := schemaFromJSON(t, txResultSchema)
	params := [][]byte{[]byte(`"0xabc"`), []byte(`{"tracer":"callTracer"}`)}
	branch, err := selectTracerSchema(schema, "debug_traceTransaction", params)
	if err != nil {
		t.Fatal(err)
	}
	if branch == nil {
		t.Fatal("expected narrowed schema")
	}
	if err := validateRaw(branch, []byte(`{"type":"CALL","from":"0xaa"}`), "t.result"); err != nil {
		t.Errorf("valid frame rejected: %v", err)
	}
	if err := validateRaw(branch, []byte(`{"from":"0xaa"}`), "t.result"); err == nil {
		t.Error("frame missing required 'type' was accepted")
	}
}

func TestValidateRawNarrowedBlockSchema(t *testing.T) {
	schema := schemaFromJSON(t, blockResultSchema)
	params := [][]byte{[]byte(`"0x1"`), []byte(`{"tracer":"callTracer"}`)}
	narrowed, err := selectTracerSchema(schema, "debug_traceBlockByNumber", params)
	if err != nil {
		t.Fatal(err)
	}
	if narrowed == nil {
		t.Fatal("expected narrowed schema")
	}
	if err := validateRaw(narrowed, []byte(`[{"txHash":"0xaa","result":{}}]`), "t.result"); err != nil {
		t.Errorf("valid entry rejected: %v", err)
	}
	if err := validateRaw(narrowed, []byte(`[{"txHash":"0xaa","error":"boom"}]`), "t.result"); err != nil {
		t.Errorf("error entry rejected: %v", err)
	}
	if err := validateRaw(narrowed, []byte(`[{"txHash":"0xaa"}]`), "t.result"); err == nil {
		t.Error("entry with neither result nor error was accepted")
	}
	if err := validateRaw(narrowed, []byte(`[{"result":{}}]`), "t.result"); err == nil {
		t.Error("entry missing required txHash was accepted")
	}
}
