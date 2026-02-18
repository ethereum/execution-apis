package specgen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"reflect"
	"slices"
	"strings"

	"github.com/ethereum/execution-apis/tools/internal/metaschema"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed base-doc.json
var baseDocJSON []byte

func init() {
	// Ensure basedoc is valid json.
	bd, err := jsonschema.UnmarshalJSON(bytes.NewReader(baseDocJSON))
	if err != nil {
		panic(err)
	}
	bdMap, ok := bd.(map[string]any)
	if !ok {
		panic(fmt.Errorf("base document is not a json object"))
	}
	if _, ok := bdMap["methods"].([]any); !ok {
		panic(fmt.Errorf("base document is missing methods or it isn't a list"))
	}
	bdComponents, ok := bdMap["components"].(map[string]any)
	if !ok {
		panic(fmt.Errorf("base document is missing components or it isn't a map"))
	}
	if _, ok := bdComponents["schemas"].(map[string]any); !ok {
		panic(fmt.Errorf("base document is missing schemas or it isn't a map"))
	}
}

type rawSchema = map[string]any

func parseSchema(content []byte) (rawSchema, error) {
	var body any
	err := yaml.Unmarshal(content, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal schema: %v", err)
	}
	result, ok := body.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("invalid schema format: %s", reflect.TypeOf(body))
	}
	return result, nil
}

func parseMethods(content []byte) ([]any, error) {
	var body any
	err := yaml.Unmarshal(content, &body)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal methods: %v", err)
	}

	items, ok := body.([]any)
	// Result must be a list.
	if !ok {
		return nil, fmt.Errorf("invalid methods format: %s", reflect.TypeOf(body))
	}

	out := make([]any, len(items))

	for i, item := range items {
		// Each element must be a rawSchema.
		element, ok := item.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid method format: %s", reflect.TypeOf(item))
		}
		out[i] = element
	}
	return out, nil
}

type Generator struct {
	baseDoc map[string]any
}

func New() *Generator {
	baseDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(baseDocJSON))
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal base document, despite init check: %v", err))
	}
	baseDocMap, ok := baseDoc.(map[string]any)
	if !ok {
		panic(fmt.Errorf("failed to unmarshal base document, despite init check: %v", err))
	}
	return &Generator{
		baseDoc: baseDocMap,
	}
}

func (s *Generator) AddMethods(methods []byte) error {
	methodSchemas, err := parseMethods(methods)
	if err != nil {
		return fmt.Errorf("failed to parse methods: %v", err)
	}
	s.baseDoc["methods"] = append(s.baseDoc["methods"].([]any), methodSchemas...)
	return nil
}

func (s *Generator) AddSchemas(schemas []byte) error {
	parsedSchema, err := parseSchema(schemas)
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %v", err)
	}
	dst := s.baseDoc["components"].(map[string]any)["schemas"].(map[string]any)
	for key, schema := range parsedSchema {
		dst[key] = schema
	}
	return nil
}

func (s *Generator) Validate() error {

	// Validate base document against the openrpc schema.
	err := metaschema.Validate(s.baseDoc)
	if err != nil {
		// Write basedoc to a tmpfile
		tmpfile, ierr := os.CreateTemp("", "base-doc-*.json")
		if ierr != nil {
			return fmt.Errorf("failed to create tmpfile: %v", ierr)
		}
		defer tmpfile.Close()
		ierr = json.NewEncoder(tmpfile).Encode(s.baseDoc)
		if ierr != nil {
			return fmt.Errorf("failed to encode base document: %v", ierr)
		}
		ierr = tmpfile.Close()
		if ierr != nil {
			return fmt.Errorf("failed to close tmpfile: %v", ierr)
		}
		fmt.Fprintf(os.Stderr, "base document is invalid, writing to %s\n", tmpfile.Name())
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return fmt.Errorf("failed to validate base document: %v", err)
	}

	return nil
}

func (s *Generator) MarshalJSON() ([]byte, error) {
	// Sort the methods by name.
	slices.SortFunc(s.baseDoc["methods"].([]any), func(a, b any) int {
		aMap, ok := a.(map[string]any)
		if !ok {
			panic(fmt.Errorf("method is not a map: %v", a))
		}
		bMap, ok := b.(map[string]any)
		if !ok {
			panic(fmt.Errorf("method is not a map: %v", b))
		}
		return strings.Compare(aMap["name"].(string), bMap["name"].(string))
	})

	// Make sure the spec is valid before marshalling.
	err := s.Validate()
	if err != nil {
		return nil, fmt.Errorf("failed to validate spec: %v", err)
	}

	return json.Marshal(s.baseDoc)
}
