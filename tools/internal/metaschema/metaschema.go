package metaschema

import (
	"encoding/json"
	"strings"

	openrpc "github.com/open-rpc/spec-types/generated/packages/go/v1_4"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var openrpcSchemaRaw = openrpc.RawOpenrpcDocument
var openrpcSchema *jsonschema.Schema

const OpenRpcSchemaURL = "https://meta.open-rpc.org/"

func init() {
	var openrpcSchemaJSON = map[string]any{}
	err := json.Unmarshal([]byte(openrpcSchemaRaw), &openrpcSchemaJSON)
	if err != nil {
		panic(err)
	}

	// The OpenRPC meta-schema uses the non-standard json-schema-tools dialect identifier in `$schema`.
	// Override it to a supported JSON Schema draft so compilation doesn't require that metaschema.
	openrpcSchemaJSON["$schema"] = "http://json-schema.org/draft-07/schema"

	// In spec-types v1_4, several `$ref`s point at the external json-schema-tools meta-schema
	// (`https://meta.json-schema.tools[/#/...]`) — an URL we don't fetch at build time. The old
	// meta-schema package inlined that content; v1_4 refs out instead. Walk the parsed schema and
	// replace each `{$ref: "<json-schema.tools URL>"}` object with an empty (permissive) schema so
	// it resolves locally. The OpenRPC document's structural rules still get validated.
	stripExternalRefs(openrpcSchemaJSON)

	compiler := jsonschema.NewCompiler()
	err = compiler.AddResource(OpenRpcSchemaURL, openrpcSchemaJSON)
	if err != nil {
		panic(err)
	}

	openrpcSchema = compiler.MustCompile(OpenRpcSchemaURL)
}

func Validate(schema map[string]any) error {
	return openrpcSchema.Validate(schema)
}

func stripExternalRefs(v any) {
	switch x := v.(type) {
	case map[string]any:
		if ref, ok := x["$ref"].(string); ok && strings.Contains(ref, "json-schema.tools") {
			for k := range x {
				delete(x, k)
			}
			return
		}
		for _, vv := range x {
			stripExternalRefs(vv)
		}
	case []any:
		for _, vv := range x {
			stripExternalRefs(vv)
		}
	}
}
