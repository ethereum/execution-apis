package testgen

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

var EngineForkchoiceUpdatedV4 = MethodTests{
	Name: "engine_forkchoiceUpdatedV4",
	Tests: []Test{
		targetGasLimitTest("target-gas-limit-zero", "0x0"),
		targetGasLimitTest("target-gas-limit-one", "0x1"),
		targetGasLimitTest("target-gas-limit-int64-max", "0x7fffffffffffffff"),
		targetGasLimitTest("target-gas-limit-uint64-high-bit", "0x8000000000000000"),
		targetGasLimitTest("target-gas-limit-uint64-max-minus-one", "0xfffffffffffffffe"),
		targetGasLimitTest("target-gas-limit-uint64-max", "0xffffffffffffffff"),
	},
}

func targetGasLimitTest(name, target string) Test {
	return Test{
		Name:             name,
		SpecOnly:         true,
		ValidationScript: targetGasLimitValidationScript,
		About: "Accepts targetGasLimit " + target + " and starts payload building at the known Amsterdam head. " +
			"The validation script checks the client-specific payloadId as 8-byte DATA and compares all other fields exactly.",
		Run: func(ctx context.Context, t *T) error {
			head := t.chain.Head()
			if !t.chain.Config().IsAmsterdam(head.Number(), head.Time()) {
				return fmt.Errorf("FCUv4 fixtures require an Amsterdam test chain")
			}
			slot := head.SlotNumber()
			if slot == nil {
				return fmt.Errorf("chain head is missing Amsterdam slotNumber")
			}
			state := map[string]interface{}{
				"headBlockHash": head.Hash(), "safeBlockHash": head.Hash(), "finalizedBlockHash": head.Hash(),
			}
			attrs := map[string]interface{}{
				"timestamp":  hexutil.EncodeUint64(head.Time() + 12),
				"prevRandao": common.Hash{}, "suggestedFeeRecipient": common.Address{},
				"withdrawals": []interface{}{}, "parentBeaconBlockRoot": common.Hash{},
				"slotNumber": hexutil.EncodeUint64(*slot + 1), "targetGasLimit": target,
			}
			var response struct {
				PayloadStatus struct {
					Status          string       `json:"status"`
					LatestValidHash *common.Hash `json:"latestValidHash"`
					ValidationError *string      `json:"validationError"`
				} `json:"payloadStatus"`
				PayloadID *hexutil.Bytes `json:"payloadId"`
			}
			if err := t.rpc.CallContext(ctx, &response, "engine_forkchoiceUpdatedV4", state, attrs, nil); err != nil {
				return err
			}
			status := response.PayloadStatus
			if status.Status != "VALID" || status.LatestValidHash == nil || *status.LatestValidHash != head.Hash() || status.ValidationError != nil {
				return fmt.Errorf("unexpected payload status: %+v", status)
			}
			if response.PayloadID == nil || len(*response.PayloadID) != 8 {
				return fmt.Errorf("expected non-null 8-byte payloadId, got %v", response.PayloadID)
			}
			return nil
		},
	}
}

// Only the opaque payload ID bytes may vary across clients. Use the fixture
// script mechanism so the Hive runner needs no Engine-specific comparison.
const targetGasLimitValidationScript = `
const expected = messages[1].expected;
const response = messages[1].response;
const id = response.result && response.result.payloadId;
if (typeof id !== "string" || !/^0x[0-9a-f]{16}$/.test(id)) {
    throw new Error("expected a non-null 8-byte payloadId");
}
response.result.payloadId = expected.result.payloadId;
assertEqual(response, expected, "response");

function assertEqual(actual, expected, path) {
    if (actual === expected) return;
    if (actual === null || expected === null ||
        typeof actual !== "object" || typeof expected !== "object" ||
        Array.isArray(actual) !== Array.isArray(expected)) {
        throw new Error(path + " differs from the expected response");
    }
    const actualKeys = Object.keys(actual).sort();
    const expectedKeys = Object.keys(expected).sort();
    if (JSON.stringify(actualKeys) !== JSON.stringify(expectedKeys)) {
        throw new Error(path + " has unexpected or missing fields");
    }
    for (const key of expectedKeys) {
        assertEqual(actual[key], expected[key], path + "." + key);
    }
}
`
