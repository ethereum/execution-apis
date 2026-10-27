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
var gasDimensionFields = []string{"regularGasUsed", "stateGasUsed", "gasRefund"}

// rootOnlyFrameFields are CallFrame fields that MUST NOT appear on nested frames.
var rootOnlyFrameFields = []string{"regularGasUsed", "gasRefund"}

// isAmsterdamBlock reports whether the given block of the test chain is at or
// after the Amsterdam fork.
func isAmsterdamBlock(t *T, number hexutil.Uint64) bool {
	block := t.chain.GetBlock(int(number))
	return t.chain.Config().IsAmsterdam(block.Number(), block.Time())
}

// checkNoGasDimensions checks that the EIP-8037 settlement fields, which are
// omitted before Amsterdam, are absent.
func checkNoGasDimensions(obj map[string]interface{}) error {
	for _, key := range gasDimensionFields {
		if _, ok := obj[key]; ok {
			return fmt.Errorf("%q MUST be omitted before Amsterdam", key)
		}
	}
	return nil
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
	return checkNoGasDimensions(result)
}
