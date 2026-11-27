package testgen

import (
	"strings"
	"testing"
)

func TestValidateCallFrameStateGasReservoir(t *testing.T) {
	frame := func(extra map[string]interface{}) map[string]interface{} {
		f := map[string]interface{}{
			"type": "CALL", "from": "0x0000000000000000000000000000000000000001",
			"to": "0x0000000000000000000000000000000000000002", "gas": "0x100",
			"gasUsed": "0x10", "input": "0x", "value": "0x0",
		}
		for k, v := range extra {
			f[k] = v
		}
		return f
	}
	amsterdamRoot := func(child map[string]interface{}) map[string]interface{} {
		return frame(map[string]interface{}{
			"regularGasUsed": "0x10", "stateGasUsed": "0x0", "gasRefund": "0x0",
			"stateGasReservoir": "0x0", "calls": []interface{}{child},
		})
	}
	reservoir := map[string]interface{}{"stateGasReservoir": "0x0"}
	for _, tc := range []struct {
		name    string
		root    map[string]interface{}
		wantErr string
	}{
		{"amsterdam, every frame has it", amsterdamRoot(frame(reservoir)), ""},
		{"amsterdam, child lacks it", amsterdamRoot(frame(nil)), "MUST be present on every frame from Amsterdam"},
		{"pre-amsterdam, absent", frame(map[string]interface{}{"calls": []interface{}{frame(nil)}}), ""},
		{"pre-amsterdam, child has it", frame(map[string]interface{}{"calls": []interface{}{frame(reservoir)}}), "MUST be omitted before Amsterdam"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := validateCallFrame(tc.root, callTracerOpts{})
			if tc.wantErr == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("got %v, want an error containing %q", err, tc.wantErr)
			}
			// Block entries go through the same root check.
			if err := validateCallTracerBlockEntries([]interface{}{map[string]interface{}{"txHash": "0x00", "result": tc.root}}, callTracerOpts{}); err == nil {
				t.Fatal("block entry: want an error, got nil")
			}
		})
	}
}
