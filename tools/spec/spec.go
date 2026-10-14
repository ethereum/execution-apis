package spec

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	openrpc "github.com/open-rpc/spec-types/generated/packages/go/v1_4"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

type Spec struct {
	fileName string
	doc      *openrpc.OpenrpcDocument
	cdoc     any // JSON-marshaled version of spec document
	jdoc     json.RawMessage
	methods  map[string]*MethodSchema
}

type ContentDescriptor struct {
	name     string
	required bool
	schema   openrpc.JSONSchemaObject
}

// MethodSchema stores all the schemas neccessary to validate a request or
// response corresponding to the method.
type MethodSchema struct {
	name   string
	index  int // array index in Spec.doc.methods
	params []*ContentDescriptor
	result *ContentDescriptor
	errors map[int64]*ErrorSchema
}

type ErrorSchema struct {
	index         int   // array index in Spec.doc.methods[*].errors
	code          int64 // error code
	hasDataSchema bool
}

// Load reads a spec file.
func Load(path string) (*Spec, error) {
	spec, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc openrpc.OpenrpcDocument
	if err := json.Unmarshal(spec, &doc); err != nil {
		return nil, err
	}
	return New(path, &doc)
}

// LoadString reads spec JSON from a string.
func LoadString(fileName, specSource string) (*Spec, error) {
	var doc openrpc.OpenrpcDocument
	if err := json.NewDecoder(strings.NewReader(specSource)).Decode(&doc); err != nil {
		return nil, err
	}
	return New(fileName, &doc)
}

// New indexes the methods and errors in the given OpenRPC spec.
func New(fileName string, doc *openrpc.OpenrpcDocument) (*Spec, error) {
	if doc.Methods == nil {
		return nil, errors.New(`missing "methods" in spec`)
	}

	// Iterate over each method in the OpenRPC spec and pull out the parameter
	// schema and result schema.
	methods := make(map[string]*MethodSchema)
	for methodIndex, method := range *doc.Methods {
		if method.ReferenceObject != nil {
			return nil, fmt.Errorf("reference object not supported, %s", *method.ReferenceObject.Ref)
		}
		var (
			method = method.MethodObject
			ms     = MethodSchema{name: string(*method.Name), index: methodIndex}
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
			cd := &ContentDescriptor{
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
			ms.errors = make(map[int64]*ErrorSchema, len(*method.Errors))
			for i, eref := range *method.Errors {
				obj, err := checkEOR(eref)
				if err != nil {
					return nil, fmt.Errorf("errors[%d]: %v", i, err)
				}
				code := int64(*obj.Code)
				if ms.errors[code] != nil {
					return nil, fmt.Errorf("errors[%d]: duplicate error code %d", i, code)
				}
				ms.errors[code] = &ErrorSchema{index: i, code: code, hasDataSchema: obj.Data != nil}
			}
		}
		ms.result = &ContentDescriptor{
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
