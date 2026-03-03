package testgen

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	gethrpc "github.com/ethereum/go-ethereum/rpc"
)

// bytes32Pattern matches a 0x-prefixed, 32-byte (64 lowercase hex character) value.
var bytes32Pattern = regexp.MustCompile(`^0x[0-9a-f]{64}$`)

// uint256Pattern matches a 0x-prefixed uint256 quantity in canonical form.
var uint256Pattern = regexp.MustCompile(`^0x(0|[1-9a-f][0-9a-f]{0,63})$`)

// multiError collects multiple validation errors and formats them as one.
type multiError struct {
	errs []string
}

func (m *multiError) add(format string, args ...interface{}) {
	m.errs = append(m.errs, fmt.Sprintf(format, args...))
}

func (m *multiError) err() error {
	if len(m.errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(m.errs, "\n  "))
}

func asNonNegativeInteger(v interface{}) (float64, bool) {
	n, ok := v.(float64)
	if !ok {
		return 0, false
	}
	if n < 0 || math.Trunc(n) != n {
		return 0, false
	}
	return n, true
}

// validateStructLog validates a single StructLog entry against the opcode tracer spec.
// All violations are collected and returned together.
//
// Key rules enforced:
//   - pc, gas, gasCost, depth, op are required
//   - error MUST be absent when no error (not null, not "")
//   - stack values MUST be canonical 0x-prefixed uint256 hex
//   - memory values MUST be 0x-prefixed, 64-char hex (bytes32)
//   - storage keys/values MUST be 0x-prefixed, 64-char hex (bytes32)
//   - storage MUST be absent (not {}) at non-SLOAD/SSTORE opcodes
func validateStructLog(i int, log map[string]interface{}) error {
	var me multiError
	prefix := fmt.Sprintf("structLogs[%d]", i)

	// Required integer fields with bounds.
	for _, field := range []string{"pc", "gas", "gasCost", "depth"} {
		v, ok := log[field]
		if !ok {
			me.add("%s: missing required field %q", prefix, field)
			continue
		}
		n, ok := asNonNegativeInteger(v)
		if !ok {
			me.add("%s: field %q must be a non-negative integer, got %v (%T)", prefix, field, v, v)
			continue
		}
		if field == "depth" && n < 1 {
			me.add("%s: field %q must be >= 1, got %v", prefix, field, n)
		}
	}

	// Required string field: op.
	opVal, ok := log["op"]
	opStr := ""
	if !ok {
		me.add("%s: missing required field \"op\"", prefix)
	} else if s, ok := opVal.(string); !ok {
		me.add("%s: field \"op\" must be a string, got %T", prefix, opVal)
	} else {
		opStr = s
	}

	// error: MUST be absent when no error occurred.
	// Clients MUST NOT include this field as null or as an empty string.
	if errVal, exists := log["error"]; exists {
		if errVal == nil {
			me.add("%s: \"error\" MUST be absent when no error (got null)", prefix)
		} else if errStr, ok := errVal.(string); !ok {
			me.add("%s: \"error\" must be a string, got %T", prefix, errVal)
		} else if errStr == "" {
			me.add("%s: \"error\" MUST be absent when no error (got empty string)", prefix)
		}
	}

	// stack: if present, each element must be canonical uint256 hex.
	if stackVal, exists := log["stack"]; exists {
		if stack, ok := stackVal.([]interface{}); !ok {
			me.add("%s: \"stack\" must be an array, got %T", prefix, stackVal)
		} else {
			for j, item := range stack {
				s, ok := item.(string)
				if !ok {
					me.add("%s: stack[%d] must be a string, got %T", prefix, j, item)
				} else if !uint256Pattern.MatchString(s) {
					me.add("%s: stack[%d] must be canonical 0x-prefixed uint256 hex, got %q", prefix, j, s)
				}
			}
		}
	}

	// memory: if present, each element must be a valid bytes32.
	if memVal, exists := log["memory"]; exists {
		if mem, ok := memVal.([]interface{}); !ok {
			me.add("%s: \"memory\" must be an array, got %T", prefix, memVal)
		} else {
			for j, item := range mem {
				s, ok := item.(string)
				if !ok {
					me.add("%s: memory[%d] must be a string, got %T", prefix, j, item)
				} else if !bytes32Pattern.MatchString(s) {
					me.add("%s: memory[%d] must be 0x-prefixed 32-byte hex (got %q)", prefix, j, s)
				}
			}
		}
	}

	// returnData: if present, must be a 0x-prefixed hex string.
	if rdVal, exists := log["returnData"]; exists {
		if rd, ok := rdVal.(string); !ok {
			me.add("%s: \"returnData\" must be a string, got %T", prefix, rdVal)
		} else if _, err := hexutil.Decode(rd); err != nil {
			me.add("%s: \"returnData\" must be valid 0x-prefixed hex bytes (got %q): %v", prefix, rd, err)
		}
	}

	// refund: optional, but must be a non-negative integer when present.
	if refundVal, exists := log["refund"]; exists {
		if _, ok := asNonNegativeInteger(refundVal); !ok {
			me.add("%s: \"refund\" must be a non-negative integer, got %v (%T)", prefix, refundVal, refundVal)
		}
	}

	// storage: if present, must only appear at SLOAD/SSTORE; must not be empty object; keys/values must be bytes32.
	if storageVal, exists := log["storage"]; exists {
		if storageVal == nil {
			me.add("%s: \"storage\" MUST be absent (not null)", prefix)
		} else if storage, ok := storageVal.(map[string]interface{}); !ok {
			me.add("%s: \"storage\" must be an object, got %T", prefix, storageVal)
		} else {
			// An empty storage object must be absent, not present as {}.
			if len(storage) == 0 {
				me.add("%s: \"storage\" MUST be absent (not {}) when no storage slots have been accessed", prefix)
			}
			// Storage is only valid at SLOAD and SSTORE opcodes.
			if opStr != "" && opStr != "SLOAD" && opStr != "SSTORE" {
				me.add("%s: \"storage\" MUST be absent at %q opcode (only SLOAD and SSTORE may populate storage)", prefix, opStr)
			}
			for k, v := range storage {
				if !bytes32Pattern.MatchString(k) {
					me.add("%s: storage key %q must be 0x-prefixed 32-byte hex", prefix, k)
				}
				vs, ok := v.(string)
				if !ok {
					me.add("%s: storage value for key %q must be a string, got %T", prefix, k, v)
				} else if !bytes32Pattern.MatchString(vs) {
					me.add("%s: storage value for key %q must be 0x-prefixed 32-byte hex (got %q)", prefix, k, vs)
				}
			}
		}
	}

	return me.err()
}

// validateOpcodeTransactionTrace validates a decoded debug_traceTransaction response
// for compliance with the execution-apis opcode tracer specification.
// All violations across all structLogs are collected and returned together.
//
// Key rules enforced:
//   - gas, failed, returnValue, structLogs are required
//   - returnValue MUST be "0x" (not "") when empty
//   - each structLog is validated by validateStructLog
func validateOpcodeTransactionTrace(result map[string]interface{}) error {
	var me multiError

	// Required top-level fields.
	for _, field := range []string{"gas", "failed", "returnValue", "structLogs"} {
		if _, ok := result[field]; !ok {
			me.add("missing required field %q", field)
		}
	}
	// Return early if required fields are absent — remaining checks would panic.
	if err := me.err(); err != nil {
		return err
	}

	// gas: must be a non-negative integer.
	if _, ok := asNonNegativeInteger(result["gas"]); !ok {
		me.add("\"gas\" must be a non-negative integer, got %v (%T)", result["gas"], result["gas"])
	}

	// failed: must be a boolean.
	if _, ok := result["failed"].(bool); !ok {
		me.add("\"failed\" must be a boolean, got %T", result["failed"])
	}

	// returnValue: must be a 0x-prefixed hex string.
	// Empty return value MUST be "0x", not "".
	if rv, ok := result["returnValue"].(string); !ok {
		me.add("\"returnValue\" must be a string, got %T", result["returnValue"])
	} else if _, err := hexutil.Decode(rv); err != nil {
		me.add("\"returnValue\" must be valid 0x-prefixed hex bytes; empty return value must be \"0x\" not \"\" (got %q): %v", rv, err)
	}

	// structLogs: must be an array; validate each entry.
	logsVal, ok := result["structLogs"].([]interface{})
	if !ok {
		me.add("\"structLogs\" must be an array, got %T", result["structLogs"])
	} else {
		for i, logVal := range logsVal {
			log, ok := logVal.(map[string]interface{})
			if !ok {
				me.add("structLogs[%d] must be an object, got %T", i, logVal)
				continue
			}
			if err := validateStructLog(i, log); err != nil {
				me.add("%s", err.Error())
			}
		}
	}

	return me.err()
}

func validateOpcodeBlockTraceResult(block *types.Block, result []map[string]interface{}) error {
	if len(result) != block.Transactions().Len() {
		return fmt.Errorf("expected %d trace entries (one per tx), got %d", block.Transactions().Len(), len(result))
	}
	for i, entry := range result {
		// txHash must be present and be a valid 32-byte hex hash.
		txHashVal, ok := entry["txHash"]
		if !ok {
			return fmt.Errorf("entry[%d]: missing required field \"txHash\"", i)
		}
		txHash, ok := txHashVal.(string)
		if !ok {
			return fmt.Errorf("entry[%d]: \"txHash\" must be a string, got %T", i, txHashVal)
		}
		if !bytes32Pattern.MatchString(txHash) {
			return fmt.Errorf("entry[%d]: \"txHash\" must be 0x-prefixed 32-byte hex (got %q)", i, txHash)
		}
		// txHash must match the actual transaction in the block.
		wantHash := block.Transactions()[i].Hash().Hex()
		if !strings.EqualFold(txHash, wantHash) {
			return fmt.Errorf("entry[%d]: \"txHash\" mismatch (got %q, want %q)", i, txHash, wantHash)
		}
		// result must be present and conform to OpcodeTransactionTrace.
		resultVal, ok := entry["result"]
		if !ok {
			return fmt.Errorf("entry[%d]: missing required field \"result\"", i)
		}
		traceResult, ok := resultVal.(map[string]interface{})
		if !ok {
			return fmt.Errorf("entry[%d]: \"result\" must be an object, got %T", i, resultVal)
		}
		if err := validateOpcodeTransactionTrace(traceResult); err != nil {
			return fmt.Errorf("entry[%d]: %w", i, err)
		}
	}
	return nil
}

func validateReturnDataFieldBehavior(logs []interface{}, expectPresent bool) error {
	for i, logVal := range logs {
		log, ok := logVal.(map[string]interface{})
		if !ok {
			return fmt.Errorf("structLogs[%d] must be an object, got %T", i, logVal)
		}
		rd, hasReturnData := log["returnData"]
		if !expectPresent && hasReturnData {
			return fmt.Errorf("structLogs[%d]: \"returnData\" MUST be absent when enableReturnData=false", i)
		}
		if hasReturnData {
			rdStr, ok := rd.(string)
			if !ok {
				return fmt.Errorf("structLogs[%d]: \"returnData\" must be a string, got %T", i, rd)
			}
			if _, err := hexutil.Decode(rdStr); err != nil {
				return fmt.Errorf("structLogs[%d]: \"returnData\" must be valid 0x-prefixed hex bytes (got %q): %v", i, rdStr, err)
			}
		}
	}
	return nil
}

// DebugTraceTransaction tests the debug_traceTransaction method.
var DebugTraceTransaction = MethodTests{
	"debug_traceTransaction",
	[]Test{
		{
			Name:     "trace-legacy-transfer",
			About:    "traces a legacy EOA-to-EOA value transfer; structLogs must be empty since no EVM code runs",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("legacy value transfer", matchLegacyValueTransfer)
				var result map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceTransaction", tx.Hash()); err != nil {
					return err
				}
				if err := validateOpcodeTransactionTrace(result); err != nil {
					return err
				}
				// EOA-to-EOA transfers execute no EVM code; structLogs MUST be empty.
				logs := result["structLogs"].([]interface{})
				if len(logs) != 0 {
					return fmt.Errorf("EOA-to-EOA value transfer must produce 0 structLogs, got %d", len(logs))
				}
				return nil
			},
		},
		{
			Name:     "trace-contract-call",
			About:    "traces a contract call transaction; validates spec compliance of structLogs including stack encoding, error field, and storage rules",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("legacy tx with input data", matchLegacyTxWithInput)
				var result map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceTransaction", tx.Hash()); err != nil {
					return err
				}
				if err := validateOpcodeTransactionTrace(result); err != nil {
					return err
				}
				// Contract calls execute EVM code and must produce at least one structLog.
				logs := result["structLogs"].([]interface{})
				if len(logs) == 0 {
					return fmt.Errorf("expected at least one structLog for contract call, got 0")
				}
				return nil
			},
		},
		{
			Name:  "trace-unknown-tx",
			About: "requests a trace for a non-existent transaction hash; the client must return an error",
			Run: func(ctx context.Context, t *T) error {
				var result map[string]interface{}
				err := t.rpc.CallContext(ctx, &result, "debug_traceTransaction",
					"0x0000000000000000000000000000000000000000000000000000000000000001")
				if err == nil {
					return fmt.Errorf("expected error for unknown transaction hash, got result: %v", result)
				}
				return nil
			},
		},
	},
}

// DebugTraceBlockByNumber tests the debug_traceBlockByNumber method.
var DebugTraceBlockByNumber = MethodTests{
	"debug_traceBlockByNumber",
	[]Test{
		{
			Name:     "trace-block-with-transactions",
			About:    "traces a block containing transactions; validates that each entry has txHash and a spec-compliant result",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				blockNum := hexutil.EncodeUint64(block.NumberU64())

				var result []map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceBlockByNumber", blockNum); err != nil {
					return err
				}
				if err := validateOpcodeBlockTraceResult(block, result); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:     "trace-block-memory-encoding",
			About:    "traces block 0x1 with memory enabled; memory chunks must be 0x-prefixed bytes32 values",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				traceCfg := map[string]interface{}{
					"disableStack":     false,
					"disableStorage":   false,
					"enableMemory":     true,
					"enableReturnData": true,
				}
				blockNum := hexutil.EncodeUint64(1)
				var result []map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceBlockByNumber", blockNum, traceCfg); err != nil {
					return fmt.Errorf("block %s: debug_traceBlockByNumber failed: %w", blockNum, err)
				}
				block := t.chain.GetBlock(1)
				if err := validateOpcodeBlockTraceResult(block, result); err != nil {
					return fmt.Errorf("block %s: %w", blockNum, err)
				}
				return nil
			},
		},
		{
			Name:     "trace-block-storage-encoding",
			About:    "traces block 0x2 with storage enabled; storage keys and values must be 0x-prefixed bytes32 values",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				traceCfg := map[string]interface{}{
					"disableStack":     false,
					"disableStorage":   false,
					"enableMemory":     false,
					"enableReturnData": true,
				}
				blockNum := hexutil.EncodeUint64(2)
				var result []map[string]interface{}
				if err := t.rpc.CallContext(ctx, &result, "debug_traceBlockByNumber", blockNum, traceCfg); err != nil {
					return fmt.Errorf("block %s: debug_traceBlockByNumber failed: %w", blockNum, err)
				}
				block := t.chain.GetBlock(2)
				if err := validateOpcodeBlockTraceResult(block, result); err != nil {
					return fmt.Errorf("block %s: %w", blockNum, err)
				}
				return nil
			},
		},
		{
			Name:     "trace-block-return-data-behavior",
			About:    "traces a block with returnData disabled and enabled to validate returnData field gating and encoding",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				blockNum := hexutil.EncodeUint64(1)

				disabledCfg := map[string]interface{}{
					"disableStack":     false,
					"disableStorage":   true,
					"enableMemory":     false,
					"enableReturnData": false,
				}
				var disabledResult []map[string]interface{}
				if err := t.rpc.CallContext(ctx, &disabledResult, "debug_traceBlockByNumber", blockNum, disabledCfg); err != nil {
					return fmt.Errorf("enableReturnData=false: debug_traceBlockByNumber failed: %w", err)
				}
				block := t.chain.GetBlock(1)
				if err := validateOpcodeBlockTraceResult(block, disabledResult); err != nil {
					return fmt.Errorf("enableReturnData=false: %w", err)
				}
				for entryIdx, entry := range disabledResult {
					traceResult, _ := entry["result"].(map[string]interface{})
					logs, _ := traceResult["structLogs"].([]interface{})
					if err := validateReturnDataFieldBehavior(logs, false); err != nil {
						return fmt.Errorf("enableReturnData=false, entry[%d]: %w", entryIdx, err)
					}
				}

				enabledCfg := map[string]interface{}{
					"disableStack":     false,
					"disableStorage":   true,
					"enableMemory":     false,
					"enableReturnData": true,
				}
				var enabledResult []map[string]interface{}
				if err := t.rpc.CallContext(ctx, &enabledResult, "debug_traceBlockByNumber", blockNum, enabledCfg); err != nil {
					return fmt.Errorf("enableReturnData=true: debug_traceBlockByNumber failed: %w", err)
				}
				if err := validateOpcodeBlockTraceResult(block, enabledResult); err != nil {
					return fmt.Errorf("enableReturnData=true: %w", err)
				}
				for entryIdx, entry := range enabledResult {
					traceResult, _ := entry["result"].(map[string]interface{})
					logs, _ := traceResult["structLogs"].([]interface{})
					if err := validateReturnDataFieldBehavior(logs, true); err != nil {
						return fmt.Errorf("enableReturnData=true, entry[%d]: %w", entryIdx, err)
					}
				}
				return nil
			},
		},
		{
			Name:  "trace-genesis",
			About: "requests a trace of the genesis block; must return an error since there is no parent state to replay from",
			Run: func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_traceBlockByNumber", "0x0")
				if err == nil {
					return fmt.Errorf("expected error tracing genesis block (no parent state available), got success")
				}
				return nil
			},
		},
		{
			Name:  "trace-block-invalid-number",
			About: "requests a trace with a non-hex block number; the client must return an error",
			Run: func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_traceBlockByNumber", "3")
				if err == nil {
					return fmt.Errorf("expected invalid-argument error for non-hex block number, got success")
				}
				if rpcErr, ok := err.(gethrpc.Error); ok && rpcErr.ErrorCode() != -32602 {
					return fmt.Errorf("expected JSON-RPC invalid params (-32602), got code %d: %v", rpcErr.ErrorCode(), err)
				}
				return nil
			},
		},
	},
}
