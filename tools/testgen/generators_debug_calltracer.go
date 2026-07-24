package testgen

import (
	"context"
	"fmt"
	"regexp"
	"slices"

	"github.com/ethereum/go-ethereum/common"
)

// Genesis predeploys created by hivechain for callTracer testing; the callee
// addresses are hardcoded in the calltree contract.
const (
	calltreeAddr           = "0x9dcd17433742f4c0ca53122ab541d0ba67fc27d0"
	calltreeCallmeAddr     = "0x9dcd17433742f4c0ca53122ab541d0ba67fc27d1"
	calltreeCallrevertAddr = "0x9dcd17433742f4c0ca53122ab541d0ba67fc27d3"
)

var callFrameTypes = map[string]bool{
	"CALL": true, "CALLCODE": true, "DELEGATECALL": true, "STATICCALL": true,
	"CREATE": true, "CREATE2": true, "SELFDESTRUCT": true,
}

// expectedCallTreeTypes is the subcall sequence produced by one invocation of
// the hivechain calltree contract.
var expectedCallTreeTypes = []string{
	"CALL", "CALL", "STATICCALL", "STATICCALL", "DELEGATECALL", "CALLCODE", "CALL", "CREATE",
}

// checkFieldPattern reports a violation when obj[key] is present, non-null,
// and not a string matching pattern. Missing and null fields are reported by
// the callers' presence checks.
func checkFieldPattern(me *multiError, prefix string, obj map[string]interface{}, key string, pattern *regexp.Regexp) {
	if v, ok := obj[key]; ok && v != nil {
		s, isStr := v.(string)
		if !isStr || !pattern.MatchString(s) {
			me.add("%s: field %q must match %s, got %v", prefix, key, pattern, v)
		}
	}
}

// callTracerOpts describes the tracerConfig a trace was requested with, so the
// validator can enforce config-dependent presence rules.
type callTracerOpts struct {
	onlyTopCall bool
	withLog     bool
}

// validateCallFrame validates a CallFrame tree against the callTracer spec.
// All violations are collected and returned together.
//
// Key rules enforced:
//   - type, from, gas, gasUsed, input are required; type is a known mnemonic
//   - absent fields MUST be omitted, never null
//   - to MUST be present except on failed CREATE/CREATE2, where it MUST be
//     omitted
//   - value MUST be absent for STATICCALL and present for all other frame
//     types
//   - error MUST be absent on success; revertReason only alongside error
//   - calls MUST be absent when empty or when onlyTopCall is set
//   - logs MUST be absent without withLog, and on failed frames and their
//     descendants
func validateCallFrame(frame map[string]interface{}, opts callTracerOpts) error {
	var me multiError
	validateCallFrameAt("", frame, opts, false, &me)
	return me.err()
}

func validateCallFrameAt(path string, frame map[string]interface{}, opts callTracerOpts, parentFailed bool, me *multiError) {
	prefix := "frame"
	if path != "" {
		prefix = path
	}

	for _, key := range []string{"type", "from", "gas", "gasUsed", "input"} {
		if _, ok := frame[key]; !ok {
			me.add("%s: missing required field %q", prefix, key)
		}
	}
	for key, val := range frame {
		if val == nil {
			me.add("%s: field %q MUST be omitted, not null", prefix, key)
		}
	}

	frameType, _ := frame["type"].(string)
	if frameType != "" && !callFrameTypes[frameType] {
		me.add("%s: unknown frame type %q", prefix, frameType)
	}
	checkFieldPattern(me, prefix, frame, "from", addressPattern)
	checkFieldPattern(me, prefix, frame, "to", addressPattern)
	checkFieldPattern(me, prefix, frame, "gas", uint256Pattern)
	checkFieldPattern(me, prefix, frame, "gasUsed", uint256Pattern)
	checkFieldPattern(me, prefix, frame, "value", uint256Pattern)
	checkFieldPattern(me, prefix, frame, "input", bytesPattern)
	checkFieldPattern(me, prefix, frame, "output", bytesPattern)

	errVal, hasError := frame["error"]
	if hasError {
		if s, ok := errVal.(string); !ok || s == "" {
			me.add("%s: \"error\" must be a non-empty string when present, got %v", prefix, errVal)
		}
	}
	if _, hasReason := frame["revertReason"]; hasReason && !hasError {
		me.add("%s: \"revertReason\" MUST only be present alongside \"error\"", prefix)
	}

	_, hasTo := frame["to"]
	createFailed := (frameType == "CREATE" || frameType == "CREATE2") && hasError
	if !hasTo && !createFailed {
		me.add("%s: \"to\" MUST be present except on failed CREATE/CREATE2", prefix)
	}
	if hasTo && createFailed {
		me.add("%s: \"to\" MUST be omitted on failed CREATE/CREATE2", prefix)
	}
	_, hasValue := frame["value"]
	switch frameType {
	case "STATICCALL":
		if hasValue {
			me.add("%s: \"value\" MUST be omitted for STATICCALL", prefix)
		}
	case "CALL", "CALLCODE", "DELEGATECALL", "CREATE", "CREATE2", "SELFDESTRUCT":
		if !hasValue {
			me.add("%s: \"value\" MUST be present for %s", prefix, frameType)
		}
	}

	failed := hasError || parentFailed
	if logsVal, hasLogs := frame["logs"]; hasLogs && logsVal != nil {
		switch {
		case !opts.withLog:
			me.add("%s: \"logs\" MUST be absent without withLog", prefix)
		case failed:
			me.add("%s: \"logs\" MUST be absent on failed frames and their descendants", prefix)
		}
		logs, ok := logsVal.([]interface{})
		if !ok {
			me.add("%s: \"logs\" must be an array, got %T", prefix, logsVal)
		} else {
			if len(logs) == 0 {
				me.add("%s: \"logs\" MUST be omitted when empty", prefix)
			}
			for i, l := range logs {
				log, ok := l.(map[string]interface{})
				if !ok {
					me.add("%s: logs[%d] must be an object, got %T", prefix, i, l)
					continue
				}
				validateCallLog(fmt.Sprintf("%s.logs[%d]", prefix, i), log, me)
			}
		}
	}

	callsVal, hasCalls := frame["calls"]
	if hasCalls && callsVal != nil {
		if opts.onlyTopCall {
			me.add("%s: \"calls\" MUST be absent when onlyTopCall is set", prefix)
		}
		calls, ok := callsVal.([]interface{})
		if !ok {
			me.add("%s: \"calls\" must be an array, got %T", prefix, callsVal)
			return
		}
		if len(calls) == 0 {
			me.add("%s: \"calls\" MUST be omitted when empty", prefix)
		}
		for i, c := range calls {
			sub, ok := c.(map[string]interface{})
			if !ok {
				me.add("%s: calls[%d] must be an object, got %T", prefix, i, c)
				continue
			}
			validateCallFrameAt(fmt.Sprintf("%s.calls[%d]", prefix, i), sub, opts, failed, me)
		}
	}
}

func validateCallLog(prefix string, log map[string]interface{}, me *multiError) {
	for _, key := range []string{"address", "topics", "data", "position"} {
		if _, ok := log[key]; !ok {
			me.add("%s: missing required field %q", prefix, key)
		}
	}
	checkFieldPattern(me, prefix, log, "address", addressPattern)
	checkFieldPattern(me, prefix, log, "data", bytesPattern)
	checkFieldPattern(me, prefix, log, "position", uint256Pattern)
	checkFieldPattern(me, prefix, log, "index", uint256Pattern)
	if topicsVal, ok := log["topics"]; ok && topicsVal != nil {
		topics, isArr := topicsVal.([]interface{})
		if !isArr {
			me.add("%s: \"topics\" must be an array, got %T", prefix, topicsVal)
			return
		}
		for i, tv := range topics {
			s, isStr := tv.(string)
			if !isStr || !bytes32Pattern.MatchString(s) {
				me.add("%s: topics[%d] must be a 32-byte hash, got %v", prefix, i, tv)
			}
		}
	}
}

// validateCallTracerBlockEntries validates a callTracer block trace response.
func validateCallTracerBlockEntries(result []interface{}, opts callTracerOpts) error {
	var me multiError
	for i, e := range result {
		prefix := fmt.Sprintf("entry[%d]", i)
		entry, ok := e.(map[string]interface{})
		if !ok {
			me.add("%s: must be an object, got %T", prefix, e)
			continue
		}
		if _, ok := entry["txHash"].(string); !ok {
			me.add("%s: missing required field \"txHash\"", prefix)
		}
		resVal, hasResult := entry["result"]
		_, hasError := entry["error"]
		if !hasResult && !hasError {
			me.add("%s: one of \"result\" or \"error\" must be present", prefix)
			continue
		}
		if hasResult {
			frame, ok := resVal.(map[string]interface{})
			if !ok {
				me.add("%s: \"result\" must be an object, got %T", prefix, resVal)
				continue
			}
			validateCallFrameAt(prefix+".result", frame, opts, false, &me)
		}
	}
	return me.err()
}

// frameTypes returns the type of every direct subcall of a frame.
func frameTypes(frame map[string]interface{}) []string {
	calls, _ := frame["calls"].([]interface{})
	var out []string
	for _, c := range calls {
		if sub, ok := c.(map[string]interface{}); ok {
			t, _ := sub["type"].(string)
			out = append(out, t)
		}
	}
	return out
}

var callTracerCfg = map[string]interface{}{"tracer": "callTracer"}

func callTracerCfgWith(tracerConfig map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{"tracer": "callTracer", "tracerConfig": tracerConfig}
}

var DebugTraceCall = MethodTests{
	"debug_traceCall",
	[]Test{
		{
			Name:     "trace-call-opcode-default",
			About:    "traces a simple call with the default opcode tracer",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(0)
				call := map[string]interface{}{
					"from":  sender,
					"to":    calltreeCallmeAddr,
					"input": "0xff01",
					"gas":   "0x100000",
				}
				var result map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceCall", call, "latest"); err != nil {
					return err
				}
				return validateOpcodeTransactionTrace(result)
			},
		},
		{
			Name:     "calltracer-nested",
			About:    "traces a call to the calltree contract with the callTracer; the frame tree must contain every call type",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(0)
				call := map[string]interface{}{
					"from": sender,
					"to":   calltreeAddr,
					"gas":  "0x100000",
				}
				var result map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceCall", call, "latest", callTracerCfg); err != nil {
					return err
				}
				if err := validateCallFrame(result, callTracerOpts{}); err != nil {
					return err
				}
				got := frameTypes(result)
				if !slices.Equal(got, expectedCallTreeTypes) {
					return fmt.Errorf("unexpected subcall types: got %v, want %v", got, expectedCallTreeTypes)
				}
				return nil
			},
		},
		{
			Name:  "calltracer-revert",
			About: "traces a call that reverts with Error(string); the revert MUST be reported in the frame, not as a JSON-RPC error",
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(0)
				call := map[string]interface{}{
					"from":  sender,
					"to":    calltreeCallrevertAddr,
					"input": "0x0000000000000000000000000000000000000000000000000000000000000001",
					"gas":   "0x100000",
				}
				var result map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceCall", call, "latest", callTracerCfg); err != nil {
					return fmt.Errorf("reverted call must not produce a JSON-RPC error, got: %v", err)
				}
				if err := validateCallFrame(result, callTracerOpts{}); err != nil {
					return err
				}
				if result["error"] != "execution reverted" {
					return fmt.Errorf("expected error \"execution reverted\", got %v", result["error"])
				}
				if result["revertReason"] != "user error" {
					return fmt.Errorf("expected revertReason \"user error\", got %v", result["revertReason"])
				}
				return nil
			},
		},
		{
			Name:  "calltracer-only-top-call",
			About: "traces a call to the calltree contract with onlyTopCall; only the root frame is returned",
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(0)
				call := map[string]interface{}{
					"from": sender,
					"to":   calltreeAddr,
					"gas":  "0x100000",
				}
				cfg := callTracerCfgWith(map[string]interface{}{"onlyTopCall": true})
				var result map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceCall", call, "latest", cfg); err != nil {
					return err
				}
				opts := callTracerOpts{onlyTopCall: true}
				if err := validateCallFrame(result, opts); err != nil {
					return err
				}
				if _, hasCalls := result["calls"]; hasCalls {
					return fmt.Errorf("calls must be absent with onlyTopCall")
				}
				return nil
			},
		},
		{
			Name:  "calltracer-state-override",
			About: "traces a call with stateOverrides replacing the callee code; the trace reflects the overridden code",
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(0)
				overrideTarget := "0x00000000000000000000000000000000000000c0"
				callmeAccount, ok := t.chain.state[common.HexToAddress(calltreeCallmeAddr)]
				if !ok || len(callmeAccount.Code) == 0 {
					return fmt.Errorf("callme predeploy %s missing from chain state", calltreeCallmeAddr)
				}
				callmeCode := callmeAccount.Code
				call := map[string]interface{}{
					"from":  sender,
					"to":    overrideTarget,
					"input": "0xff01",
					"gas":   "0x100000",
				}
				cfg := map[string]interface{}{
					"tracer": "callTracer",
					"stateOverrides": map[string]interface{}{
						overrideTarget: map[string]interface{}{
							"code":  callmeCode,
							"state": map[string]interface{}{},
						},
					},
				}
				var result map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceCall", call, "latest", cfg); err != nil {
					return err
				}
				if err := validateCallFrame(result, callTracerOpts{}); err != nil {
					return err
				}
				if result["output"] != "0xffee" {
					return fmt.Errorf("expected overridden code output 0xffee, got %v", result["output"])
				}
				return nil
			},
		},
		{
			Name:  "calltracer-block-param-hash",
			About: "traces a call with the block given by hash",
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(0)
				call := map[string]interface{}{
					"from":  sender,
					"to":    calltreeCallmeAddr,
					"input": "0xff01",
					"gas":   "0x100000",
				}
				var result map[string]interface{}
				head := t.chain.Head().Hash()
				if err := t.rpc.CallContext(ctx, &result, "debug_traceCall", call, head, callTracerCfg); err != nil {
					return err
				}
				if err := validateCallFrame(result, callTracerOpts{}); err != nil {
					return err
				}
				if result["output"] != "0xffee" {
					return fmt.Errorf("expected output 0xffee, got %v", result["output"])
				}
				return nil
			},
		},
		{
			Name:  "calltracer-invalid-block",
			About: "traces a call at a non-existent block; an error is expected",
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(0)
				call := map[string]interface{}{
					"from": sender,
					"to":   calltreeCallmeAddr,
					"gas":  "0x100000",
				}
				err := t.rpc.CallContext(ctx, nil, "debug_traceCall", call, "0xfffffffff", callTracerCfg)
				if err == nil {
					return fmt.Errorf("expected error for non-existent block")
				}
				return nil
			},
		},
	},
}
