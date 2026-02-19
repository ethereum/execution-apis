package specgen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"slices"

	"github.com/ethereum/execution-apis/tools/internal/metaschema"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"gopkg.in/yaml.v3"
)

//go:embed base-doc.json
var baseDocJSON []byte

type object = map[string]any

type schemaRepository = map[string]object

type Generator struct {
	baseDoc object
	methods map[string]object
	types   schemaRepository
}

func New() *Generator {
	baseDoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(baseDocJSON))
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal base document, despite init check: %v", err))
	}
	baseDocMap, ok := baseDoc.(object)
	if !ok {
		panic(fmt.Errorf("failed to unmarshal base document, despite init check: %v", err))
	}
	return &Generator{
		methods: make(map[string]object),
		types:   make(schemaRepository),
		baseDoc: baseDocMap,
	}
}

// AddMethods parses the given YAML content and adds the methods defined within it to the spec.
//
// The input is assumed to be a list of method objects.
func (s *Generator) AddMethods(methods []byte) error {
	methodSchemas, err := parseMethods(methods)
	if err != nil {
		return fmt.Errorf("failed to parse methods: %v", err)
	}
	for i, method := range methodSchemas {
		nameVal, ok := method["name"]
		if !ok {
			return fmt.Errorf("method %d has no name", i)
		}
		name, ok := nameVal.(string)
		if !ok {
			return fmt.Errorf("method %d name is not a string: %v", i, nameVal)
		}
		if _, exists := s.methods[name]; exists {
			return fmt.Errorf("method %s already defined", name)
		}
		s.methods[name] = method
	}
	return nil
}

func parseMethods(content []byte) ([]object, error) {
	var body []object
	err := yaml.Unmarshal(content, &body)
	if err != nil {
		err = fmt.Errorf("failed to unmarshal methods: %v", err)
	}
	return body, err
}

// AddSchemas parses the given YAML content and adds the given schemas (type definitions)
// to the generator.
//
// The input is assumed to be of object shape, i.e. it contains key-value pairs.
func (s *Generator) AddSchemas(schemas []byte) error {
	parsedSchema, err := parseSchema(schemas)
	if err != nil {
		return fmt.Errorf("failed to parse schemas: %v", err)
	}
	for key, schema := range parsedSchema {
		if _, exists := s.types[key]; exists {
			return fmt.Errorf("duplicate schema %s", key)
		}
		s.types[key] = schema
	}
	return nil
}

func parseSchema(content []byte) (map[string]object, error) {
	var body map[string]object
	err := yaml.Unmarshal(content, &body)
	if err != nil {
		return nil, fmt.Errorf("invalid schema content: %v", err)
	}
	return body, nil
}

// Dereference removes all $ref pointers and ensures the spec and does not use the `allOf`
// schema feature. This exists to simplify the spec for compatibility with some tools.
func (s *Generator) Dereference() error {
	// Pre-resolve all type schemas: dereference $ref and merge allOf.
	// The resolved map is used when expanding method schemas. The output
	// document does not include the schemas; "components" is left empty.
	types := make(schemaRepository, len(s.types))
	for name, schema := range s.types {
		exp, err := s.expandSchema(schema, s.types)
		if err != nil {
			return fmt.Errorf("schema %s: %w", name, err)
		}
		types[name] = exp
	}

	// Expand $ref in method param/result/error schemas and collect methods.
	methods := make(schemaRepository, len(s.methods))
	for _, name := range slices.Sorted(maps.Keys(s.methods)) {
		method, err := s.expandMethod(s.methods[name], types)
		if err != nil {
			return fmt.Errorf("method %s: %w", name, err)
		}
		methods[name] = method
	}

	clear(s.types)
	s.methods = methods
	return nil
}

// expandSchema dereferences schema using the given repository of types and also merges
// uses of "allOf".
func (s *Generator) expandSchema(schema object, types map[string]object) (object, error) {
	expanded, err := dereference(schema, types)
	if err != nil {
		return nil, err
	}
	return mergeAllOf(expanded), nil
}

// expandMethod returns a copy of method with all $ref entries in param, result,
// and error schemas expanded using the given repository of types.
func (s *Generator) expandMethod(method object, types map[string]object) (object, error) {
	out := maps.Clone(method)

	// Expand each param schema.
	if params, ok := method["params"].([]any); ok {
		expanded := make([]any, len(params))
		for i, p := range params {
			param, ok := p.(object)
			if !ok {
				expanded[i] = p
				continue
			}
			param = maps.Clone(param)
			if schema, ok := param["schema"].(object); ok {
				exp, err := s.expandSchema(schema, types)
				if err != nil {
					return nil, fmt.Errorf("param %d schema: %w", i, err)
				}
				param["schema"] = exp
			}
			expanded[i] = param
		}
		out["params"] = expanded
	}

	// Expand result schema.
	if result, ok := method["result"].(object); ok {
		result = maps.Clone(result)
		if schema, ok := result["schema"].(object); ok {
			exp, err := s.expandSchema(schema, types)
			if err != nil {
				return nil, fmt.Errorf("result schema: %w", err)
			}
			result["schema"] = exp
		}
		out["result"] = result
	}

	// Expand error data schemas.
	if errors, ok := method["errors"].([]any); ok {
		expanded := make([]any, len(errors))
		for i, e := range errors {
			errObj, ok := e.(object)
			if !ok {
				expanded[i] = e
				continue
			}
			errObj = maps.Clone(errObj)
			if schema, ok := errObj["data"].(object); ok {
				exp, err := s.expandSchema(schema, types)
				if err != nil {
					return nil, fmt.Errorf("error %d data schema: %w", i, err)
				}
				errObj["data"] = exp
			}
			expanded[i] = errObj
		}
		out["errors"] = expanded
	}

	return out, nil
}

// build creates the spec document.
func (s *Generator) build() object {
	doc := maps.Clone(s.baseDoc)
	doc["components"] = object{"schemas": repo2object(s.types)}
	var methods []any
	for _, name := range slices.Sorted(maps.Keys(s.methods)) {
		methods = append(methods, s.methods[name])
	}
	doc["methods"] = methods
	return doc
}

func repo2object(r schemaRepository) object {
	obj := make(object, len(r))
	for k, v := range r {
		obj[k] = v
	}
	return obj
}

// Validate builds the spec and checks it against the OpenRPC meta-schema.
func (s *Generator) Validate() error {
	doc := s.build()
	if err := validate(doc); err != nil {
		return fmt.Errorf("spec is invalid: %v", err)
	}
	return nil
}

func validate(doc object) error {
	err := metaschema.Validate(doc)
	if err != nil {
		tmpfile, err := os.CreateTemp("", "doc-*.json")
		if err != nil {
			return fmt.Errorf("failed to create tmpfile: %v", err)
		}
		defer tmpfile.Close()
		if err := json.NewEncoder(tmpfile).Encode(doc); err != nil {
			return fmt.Errorf("failed to encode document: %v", err)
		}
		log.Printf("spec is invalid, written to %s\n", tmpfile.Name())
	}
	return err
}

// JSON creates the spec document.
func (s *Generator) JSON() ([]byte, error) {
	doc := s.build()
	if err := validate(doc); err != nil {
		return nil, fmt.Errorf("spec is invalid: %v", err)
	}
	return json.MarshalIndent(doc, "", "  ")
}
