package metaschema

import (
	"encoding/json"

	openrpc "github.com/open-rpc/meta-schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

var openrpcSchemaRaw = openrpc.RawOpenrpc_document
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

	// The upstream OpenRPC meta-schema embeds a copy of the json-schema-tools meta-schema under
	// `definitions.JSONSchema` and gives it `$id: https://meta.json-schema.tools/`.
	// Remove those identifiers so the embedded definition behaves like a normal local subschema.
	if defs, ok := openrpcSchemaJSON["definitions"].(map[string]any); ok {
		if js, ok := defs["JSONSchema"].(map[string]any); ok {
			delete(js, "$id")
			delete(js, "$schema")
		}
	}

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
