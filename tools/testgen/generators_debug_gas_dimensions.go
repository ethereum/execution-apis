package testgen

import (
	"context"
	"fmt"
	"regexp"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

// intPattern matches the signed hex integer base type.
var intPattern = regexp.MustCompile(`^(0x0|-?0x[1-9a-f][0-9a-f]*)$`)

// gasDimensionFields are the EIP-8037 settlement fields of a transaction trace.
var gasDimensionFields = []string{"executionGasUsed", "stateGasUsed", "gasRefund"}

// rootOnlyFrameFields are CallFrame fields that MUST NOT appear on nested frames.
var rootOnlyFrameFields = []string{"executionGasUsed", "gasRefund"}

// isAmsterdamBlock reports whether the given block of the test chain is at or
// after the Amsterdam fork.
func isAmsterdamBlock(t *T, number hexutil.Uint64) bool {
	block := t.chain.GetBlock(int(number))
	return t.chain.Config().IsAmsterdam(block.Number(), block.Time())
}

// checkGasDimensionPresence checks that the EIP-8037 settlement fields are all
// present from Amsterdam on and all absent before.
func checkGasDimensionPresence(obj map[string]interface{}, amsterdam bool) error {
	for _, key := range gasDimensionFields {
		_, ok := obj[key]
		switch {
		case amsterdam && !ok:
			return fmt.Errorf("%q MUST be present for Amsterdam transactions", key)
		case !amsterdam && ok:
			return fmt.Errorf("%q MUST be omitted before Amsterdam", key)
		}
	}
	return nil
}

// checkSettlementSum checks executionGasUsed + stateGasUsed == gasUsed +
// gasRefund, which holds when the calldata floor binds neither quantity.
func checkSettlementSum(gasUsed, execution, state, refund uint64) error {
	if execution+state != gasUsed+refund {
		return fmt.Errorf("executionGasUsed (%d) + stateGasUsed (%d) != gasUsed (%d) + gasRefund (%d)", execution, state, gasUsed, refund)
	}
	return nil
}

func hexField(frame map[string]interface{}, key string) (uint64, error) {
	s, _ := frame[key].(string)
	v, err := hexutil.DecodeUint64(s)
	if err != nil {
		return 0, fmt.Errorf("field %q: %v", key, err)
	}
	return v, nil
}

func intField(result map[string]interface{}, key string) (uint64, error) {
	v, ok := asNonNegativeInteger(result[key])
	if !ok {
		return 0, fmt.Errorf("%q must be a non-negative integer, got %v (%T)", key, result[key], result[key])
	}
	return uint64(v), nil
}

// traceCallTracerGasDimensions traces an Amsterdam calltree invocation with the
// callTracer and checks the root frame's EIP-8037 settlement.
func traceCallTracerGasDimensions(ctx context.Context, t *T) error {
	txs := t.chain.txinfo.CallTreeTxs
	info := txs[len(txs)-1]
	if !isAmsterdamBlock(t, info.Block) {
		return fmt.Errorf("calltree tx in block %d is not on Amsterdam", info.Block)
	}
	var result map[string]interface{}
	if err := t.rpc.CallContext(ctx, &result, "debug_traceTransaction", info.TxHash, callTracerCfg); err != nil {
		return err
	}
	if err := validateCallFrame(result, callTracerOpts{}); err != nil {
		return err
	}
	if err := checkGasDimensionPresence(result, true); err != nil {
		return err
	}
	var vals [4]uint64
	for i, key := range []string{"gasUsed", "executionGasUsed", "stateGasUsed", "gasRefund"} {
		v, err := hexField(result, key)
		if err != nil {
			return err
		}
		vals[i] = v
	}
	// The calltree invocation executes far more gas than its calldata floor.
	return checkSettlementSum(vals[0], vals[1], vals[2], vals[3])
}

// traceCallTracerNoGasDimensions traces a pre-Amsterdam transfer with the
// callTracer and checks the EIP-8037 settlement fields are absent.
func traceCallTracerNoGasDimensions(ctx context.Context, t *T) error {
	info := t.chain.txinfo.LegacyTransfers[0]
	if isAmsterdamBlock(t, info.Block) {
		return fmt.Errorf("transfer in block %d is not before Amsterdam", info.Block)
	}
	var result map[string]interface{}
	if err := t.rpc.CallContext(ctx, &result, "debug_traceTransaction", info.TxHash, callTracerCfg); err != nil {
		return err
	}
	if err := validateCallFrame(result, callTracerOpts{}); err != nil {
		return err
	}
	return checkGasDimensionPresence(result, false)
}

// traceOpcodeGasDimensions traces an Amsterdam transfer with the opcode logger
// and checks the transaction-level EIP-8037 settlement.
func traceOpcodeGasDimensions(ctx context.Context, t *T) error {
	txs := t.chain.txinfo.DynamicFeeTransfers
	info := txs[len(txs)-1]
	if !isAmsterdamBlock(t, info.Block) {
		return fmt.Errorf("transfer in block %d is not on Amsterdam", info.Block)
	}
	var result map[string]interface{}
	if err := t.rpc.CallContext(ctx, &result, "debug_traceTransaction", info.TxHash); err != nil {
		return err
	}
	if err := validateOpcodeTransactionTrace(result); err != nil {
		return err
	}
	if err := checkGasDimensionPresence(result, true); err != nil {
		return err
	}
	var vals [4]uint64
	for i, key := range []string{"gas", "executionGasUsed", "stateGasUsed", "gasRefund"} {
		v, err := intField(result, key)
		if err != nil {
			return err
		}
		vals[i] = v
	}
	// A plain transfer has no calldata, so its floor equals the intrinsic gas.
	return checkSettlementSum(vals[0], vals[1], vals[2], vals[3])
}
