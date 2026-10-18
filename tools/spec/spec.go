package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	openrpc "github.com/open-rpc/spec-types/generated/packages/go/v1_4"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

// Spec wraps an OpenRPC spec for JSON-Schema validation.
type Spec struct {
	fileName string
	doc      *openrpc.OpenrpcDocument
	cdoc     any             // JSON-marshaled version of spec document
	jdoc     json.RawMessage // same but bytes
	methods  map[string]*methodSchema
}

type contentDescriptor struct {
	name     string
	required bool
	schema   openrpc.JSONSchemaObject
}

// methodSchema stores all the schemas neccessary to validate a request or
// response corresponding to the method.
type methodSchema struct {
	name   string
	index  int // array index in Spec.doc.methods
	params []*contentDescriptor
	result *contentDescriptor
	errors map[int64]*errorSchema
}

type errorSchema struct {
	index         int   // array index in Spec.doc.methods[*].errors
	code          int64 // error code
	hasDataSchema bool
}

// Load reads an OpenRPC spec file.
func Load(path string) (*Spec, error) {
	spec, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return parse(path, spec)
}

// LoadString reads OpenRPC spec JSON from a string.
func LoadString(fileName, specSource string) (*Spec, error) {
	return parse(fileName, []byte(specSource))
}

// parse indexes the spec and validates against the document as written. The
// generated OpenRPC types rewrite tuple `items` when marshaled back to JSON.
func parse(fileName string, data []byte) (*Spec, error) {
	var doc openrpc.OpenrpcDocument
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, err
	}
	s, err := New(fileName, &doc)
	if err != nil {
		return nil, err
	}
	// Keep numbers exact, so bounds such as the uint64 maximum are enforced.
	cdoc, err := jsonschema.UnmarshalJSON(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	s.cdoc, s.jdoc = cdoc, bytes.Clone(data)
	return s, nil
}

// New indexes the methods and errors in the given OpenRPC spec. It validates
// against the document marshaled from the generated OpenRPC types, which rewrite
// tuple `items`; Load and LoadString validate against the document as written.
func New(fileName string, doc *openrpc.OpenrpcDocument) (*Spec, error) {
	if doc.Methods == nil {
		return nil, errors.New(`missing "methods" in spec`)
	}

	// Iterate over each method in the OpenRPC spec and pull out the parameter
	// schema and result schema.
	methods := make(map[string]*methodSchema)
	for methodIndex, method := range *doc.Methods {
		if method.ReferenceObject != nil {
			return nil, fmt.Errorf("reference object not supported, %s", *method.ReferenceObject.Ref)
		}
		var (
			method = method.MethodObject
			ms     = methodSchema{name: string(*method.Name), index: methodIndex}
		)
		// Add parameter schemas.
		for i, param := range *method.Params {
			if err := checkCDOR(param); err != nil {
				return nil, fmt.Errorf("%s, parameter %d: %v", *method.Name, i, err)
			}
			required := false
			if param.ContentDescriptorObject.Required != nil && *param.ContentDescriptorObject.Required {
				required = true
			}
			cd := &contentDescriptor{
				name:     string(*param.ContentDescriptorObject.Name),
				required: required,
				schema:   *param.ContentDescriptorObject.Schema.JSONSchemaObject,
			}
			ms.params = append(ms.params, cd)
		}

		// Add result schema.
		if method.Result == nil {
			return nil, fmt.Errorf("%s: missing result", *method.Name)
		}
		cdor := openrpc.ContentDescriptorOrReference{
			ContentDescriptorObject: method.Result.ContentDescriptorObject,
			ReferenceObject:         method.Result.ReferenceObject,
		}
		if err := checkCDOR(cdor); err != nil {
			return nil, fmt.Errorf("%s: %v", *method.Name, err)
		}
		obj := method.Result.ContentDescriptorObject
		required := false
		if obj.Required != nil && *obj.Required {
			required = true
		}

		// Add error codes.
		if method.Errors != nil {
			ms.errors = make(map[int64]*errorSchema, len(*method.Errors))
			for i, eref := range *method.Errors {
				obj, err := checkEOR(eref)
				if err != nil {
					return nil, fmt.Errorf("errors[%d]: %v", i, err)
				}
				code := int64(*obj.Code)
				if ms.errors[code] != nil {
					return nil, fmt.Errorf("errors[%d]: duplicate error code %d", i, code)
				}
				ms.errors[code] = &errorSchema{index: i, code: code, hasDataSchema: obj.Data != nil}
			}
		}
		ms.result = &contentDescriptor{
			name:     string(*obj.Name),
			required: required,
			schema:   *obj.Schema.JSONSchemaObject,
		}
		methods[string(*method.Name)] = &ms
	}

	jdoc, err := json.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("spec marshaling failed: %v", err)
	}
	var cdoc any
	if err := json.Unmarshal(jdoc, &cdoc); err != nil {
		panic("spec unmarshaling failed: " + err.Error())
	}
	c := &Spec{fileName: fileName, doc: doc, cdoc: cdoc, jdoc: jdoc, methods: methods}
	return c, nil
}

// Validate checks a value against the schema pointed-to by `url`.
// The `url` is expected to be a JSON-Pointer that reaches a JSON-Schema in the spec document,
// i.e. it will typically be something like "#/components/schemas/XXX" or "#/methods[i]/params[j].schema".
func (s *Spec) Validate(url string, val any) error {
	if len(url) == 0 || url[0] != '#' {
		return fmt.Errorf("invalid schema URL, expected JSON-Pointer fragment: %q", url)
	}
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2019)
	c.AddResource(s.fileName, s.cdoc)
	schema, err := c.Compile(s.fileName + url)
	if err != nil {
		return err
	}
	return schema.Validate(val)
}

// JSON returns a marshaled representation of the spec.
func (s *Spec) JSON() json.RawMessage {
	return bytes.Clone(s.jdoc)
}

func checkCDOR(obj openrpc.ContentDescriptorOrReference) error {
	if obj.ReferenceObject != nil {
		return fmt.Errorf("references not supported")
	}
	if obj.ContentDescriptorObject == nil {
		return fmt.Errorf("missing content descriptor")
	}
	cd := obj.ContentDescriptorObject
	if cd.Name == nil {
		return fmt.Errorf("missing name")
	}
	if cd.Schema == nil || cd.Schema.JSONSchemaObject == nil {
		return fmt.Errorf("missing schema")
	}
	return nil
}

func checkEOR(obj openrpc.ErrorOrReference) (*openrpc.ErrorObject, error) {
	if obj.ReferenceObject != nil {
		return nil, fmt.Errorf("error references not supported")
	}
	if obj.ErrorObject == nil {
		return nil, fmt.Errorf("missing error descriptor")
	}
	err := obj.ErrorObject
	if err.Code == nil {
		return nil, fmt.Errorf("missing error code")
	}
	return err, nil
}
