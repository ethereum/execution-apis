package specgen

import "testing"

// verifies that refs to a component's $id become fragment refs and the $id is dropped
func TestLocalizeIDRefs(t *testing.T) {
	const id = "https://example.com/frame"
	g := New()
	g.types["Frame"] = object{
		"$id":  id,
		"type": "object",
		"properties": object{
			"to":    object{"$ref": "#/components/schemas/address"},
			"calls": object{"type": "array", "items": object{"$ref": id}},
		},
	}
	g.methods["m"] = object{"result": object{"schema": object{"$ref": id}}}

	g.LocalizeIDRefs()

	frame := g.types["Frame"]
	if _, ok := frame["$id"]; ok {
		t.Error("$id was not dropped")
	}
	props := frame["properties"].(object)
	if ref := props["calls"].(object)["items"].(object)["$ref"]; ref != "#/components/schemas/Frame" {
		t.Errorf("self ref: want #/components/schemas/Frame, got %v", ref)
	}
	if ref := props["to"].(object)["$ref"]; ref != "#/components/schemas/address" {
		t.Errorf("fragment ref changed: %v", ref)
	}
	if ref := g.methods["m"]["result"].(object)["schema"].(object)["$ref"]; ref != "#/components/schemas/Frame" {
		t.Errorf("method ref: want #/components/schemas/Frame, got %v", ref)
	}
}
