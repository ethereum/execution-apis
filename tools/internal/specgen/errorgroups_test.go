package specgen

import (
	"strings"
	"testing"
)

func intPtr(v int) *int { return &v }

func ref(name string) object {
	return object{"$ref": errorGroupRefPrefix + name}
}

func TestResolveMethodErrors(t *testing.T) {
	tests := []struct {
		name           string
		groups         errorGroups
		existingErrors []any
		refs           []any
		wantCodes      []int  // expected codes in order; nil when expecting error
		wantErr        string // substring expected in error
	}{
		{
			name: "single group ref",
			groups: errorGroups{
				{
					Name:  "GasErrors",
					Range: groupRange{Min: intPtr(800), Max: intPtr(999)},
					Errors: []Error{
						{Code: 800, Message: "Gas too low"},
						{Code: 801, Message: "Out of gas"},
					},
				},
			},
			refs:      []any{ref("GasErrors")},
			wantCodes: []int{800, 801},
		},
		{
			name: "multiple groups",
			groups: errorGroups{
				{Name: "A", Errors: []Error{{Code: 1, Message: "a"}}},
				{Name: "B", Errors: []Error{{Code: 2, Message: "b"}}},
			},
			refs:      []any{ref("A"), ref("B")},
			wantCodes: []int{1, 2},
		},
		{
			name: "merge with existing",
			groups: errorGroups{
				{Name: "GasErrors", Errors: []Error{{Code: 800, Message: "Gas too low"}}},
			},
			existingErrors: []any{object{"code": 42, "message": "inline error"}},
			refs:           []any{ref("GasErrors")},
			wantCodes:      []int{42, 800},
		},
		{
			name: "inline takes precedence on duplicate code",
			groups: errorGroups{
				{Name: "ExecutionErrors", Errors: []Error{{Code: 3, Message: "Execution reverted"}}},
			},
			existingErrors: []any{object{"code": 3, "message": "Execution reverted", "data": "0xdeadbeef"}},
			refs:           []any{ref("ExecutionErrors")},
			wantCodes:      []int{3},
		},
		{
			name: "out of range code in group",
			groups: errorGroups{
				{
					Name:   "GasErrors",
					Range:  groupRange{Min: intPtr(800), Max: intPtr(999)},
					Errors: []Error{{Code: 1500, Message: "bad code"}},
				},
			},
			refs:    []any{ref("GasErrors")},
			wantErr: "out of range",
		},
		{
			name:    "missing group",
			refs:    []any{ref("Missing")},
			wantErr: "not found",
		},
		{
			name:    "invalid ref type",
			refs:    []any{"not an object"},
			wantErr: "expected object",
		},
		{
			name:    "missing $ref field",
			refs:    []any{object{"name": "no ref"}},
			wantErr: "expected $ref",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.groups.resolveMethodErrors(tt.existingErrors, tt.refs)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error should mention %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.wantCodes) {
				t.Fatalf("want %d errors, got %d", len(tt.wantCodes), len(got))
			}
			for i, code := range tt.wantCodes {
				if got[i].(object)["code"] != code {
					t.Errorf("error[%d] code: want %d, got %v", i, code, got[i].(object)["code"])
				}
			}
		})
	}
}

func TestValidateErrorCode(t *testing.T) {
	tests := []struct {
		name    string
		group   errorGroup
		code    int
		wantErr string
	}{
		{
			name:  "in range",
			group: errorGroup{Name: "GasErrors", Range: groupRange{Min: intPtr(800), Max: intPtr(999)}},
			code:  800,
		},
		{
			name:  "at max boundary",
			group: errorGroup{Name: "GasErrors", Range: groupRange{Min: intPtr(800), Max: intPtr(999)}},
			code:  999,
		},
		{
			name:    "out of range",
			group:   errorGroup{Name: "GasErrors", Range: groupRange{Min: intPtr(800), Max: intPtr(999)}},
			code:    1500,
			wantErr: "out of range",
		},
		{
			name:  "no range",
			group: errorGroup{Name: "JSONRPCStandardErrors"},
			code:  -32700,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.group.validateErrorCode(Error{Code: tt.code})
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error should mention %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckRangeOverlap(t *testing.T) {
	tests := []struct {
		name     string
		existing errorGroups
		newGroup errorGroup
		wantErr  string
	}{
		{
			name: "no overlap",
			existing: errorGroups{
				{Name: "A", Range: groupRange{Min: intPtr(100), Max: intPtr(199)}},
			},
			newGroup: errorGroup{Name: "B", Range: groupRange{Min: intPtr(200), Max: intPtr(299)}},
		},
		{
			name: "overlap",
			existing: errorGroups{
				{Name: "A", Range: groupRange{Min: intPtr(100), Max: intPtr(250)}},
			},
			newGroup: errorGroup{Name: "B", Range: groupRange{Min: intPtr(200), Max: intPtr(299)}},
			wantErr:  "overlaps",
		},
		{
			name: "group without range is skipped",
			existing: errorGroups{
				{Name: "StandardErrors"},
			},
			newGroup: errorGroup{Name: "B", Range: groupRange{Min: intPtr(100), Max: intPtr(199)}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.existing.checkRangeOverlap(tt.newGroup)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error should mention %q, got: %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestParseErrorGroupRef(t *testing.T) {
	tests := []struct {
		name     string
		ref      string
		wantName string
		wantErr  bool
	}{
		{
			name:     "valid",
			ref:      errorGroupRefPrefix + "GasErrors",
			wantName: "GasErrors",
		},
		{
			name:    "invalid prefix",
			ref:     "#/components/schemas/Foo",
			wantErr: true,
		},
		{
			name:    "empty name",
			ref:     errorGroupRefPrefix,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, err := parseErrorGroupRef(tt.ref)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if name != tt.wantName {
				t.Errorf("want %s, got %s", tt.wantName, name)
			}
		})
	}
}

func TestParseErrorGroups(t *testing.T) {
	tests := []struct {
		name         string
		yml          string
		wantGroup    string
		wantCategory string
		wantMin      *int
		wantMax      *int
		wantCount    int
		wantErr      bool
	}{
		{
			name: "with range",
			yml: `
GasErrors:
  category: "GAS_ERRORS"
  range:
    min: 800
    max: 999
  errors:
    - code: 800
      message: "Gas too low"
    - code: 801
      message: "Out of gas"
`,
			wantGroup:    "GasErrors",
			wantCategory: "GAS_ERRORS",
			wantMin:      intPtr(800),
			wantMax:      intPtr(999),
			wantCount:    2,
		},
		{
			name:    "invalid yaml",
			yml:     `not: [valid: yaml`,
			wantErr: true,
		},
		{
			name: "no range",
			yml: `
StandardErrors:
  errors:
    - code: -32700
      message: "Parse error"
`,
			wantGroup: "StandardErrors",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			groups, err := parseErrorGroups([]byte(tt.yml))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			group, ok := groups.find(tt.wantGroup)
			if !ok {
				t.Fatalf("%s not found", tt.wantGroup)
			}
			if group.Name != tt.wantGroup {
				t.Errorf("name: want %s, got %s", tt.wantGroup, group.Name)
			}
			if tt.wantMin != nil {
				if group.Range.Min == nil || *group.Range.Min != *tt.wantMin {
					t.Errorf("range min: want %d, got %v", *tt.wantMin, group.Range.Min)
				}
			} else if group.Range.Min != nil {
				t.Errorf("range min: want nil, got %d", *group.Range.Min)
			}
			if tt.wantMax != nil {
				if group.Range.Max == nil || *group.Range.Max != *tt.wantMax {
					t.Errorf("range max: want %d, got %v", *tt.wantMax, group.Range.Max)
				}
			} else if group.Range.Max != nil {
				t.Errorf("range max: want nil, got %d", *group.Range.Max)
			}
			if len(group.Errors) != tt.wantCount {
				t.Fatalf("want %d errors, got %d", tt.wantCount, len(group.Errors))
			}
		})
	}
}

func TestErrorToObject(t *testing.T) {
	tests := []struct {
		name     string
		input    Error
		wantData any
	}{
		{
			name:  "without data",
			input: Error{Code: 800, Message: "test"},
		},
		{
			name:     "with data",
			input:    Error{Code: 3, Message: "reverted", Data: "revert reason"},
			wantData: "revert reason",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj := tt.input.toObject()
			if obj["code"] != tt.input.Code {
				t.Errorf("code: want %d, got %v", tt.input.Code, obj["code"])
			}
			if obj["message"] != tt.input.Message {
				t.Errorf("message: want %s, got %v", tt.input.Message, obj["message"])
			}
			if tt.wantData != nil {
				if obj["data"] != tt.wantData {
					t.Errorf("data: want %v, got %v", tt.wantData, obj["data"])
				}
			} else if _, has := obj["data"]; has {
				t.Error("data should not be present when nil")
			}
		})
	}
}
