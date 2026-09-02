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
const draft07SchemaURL = "http://json-schema.org/draft-07/schema#"

var externalRefReplacements = map[string]string{
	"https://meta.json-schema.tools":                                                draft07SchemaURL,
	"https://meta.json-schema.tools/#/definitions/JSONSchemaObject/properties/$ref": draft07SchemaURL + "/properties/$ref",
}

func init() {
	var openrpcSchemaJSON = map[string]any{}
	err := json.Unmarshal([]byte(openrpcSchemaRaw), &openrpcSchemaJSON)
	if err != nil {
		panic(err)
	}

	// The OpenRPC meta-schema uses the non-standard json-schema-tools dialect identifier in `$schema`.
	// Override it to a supported JSON Schema draft so compilation doesn't require that metaschema.
	openrpcSchemaJSON["$schema"] = "http://json-schema.org/draft-07/schema"

	replaceExternalSchemaRefs(openrpcSchemaJSON)

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

func replaceExternalSchemaRefs(v any) {
	switch x := v.(type) {
	case map[string]any:
		if ref, ok := x["$ref"].(string); ok && strings.Contains(ref, "json-schema.tools") {
			if repl, found := externalRefReplacements[ref]; found {
				x["$ref"] = repl
			} else {
				x["$ref"] = draft07SchemaURL
			}
		}
		for _, vv := range x {
			replaceExternalSchemaRefs(vv)
		}
	case []any:
		for _, vv := range x {
			replaceExternalSchemaRefs(vv)
		}
	}
}
