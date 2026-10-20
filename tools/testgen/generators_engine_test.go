package testgen

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/execution-apis/tools/iofile"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

type fixtureEngine struct {
	head   common.Hash
	status string
	id     interface{}
	target string
}

func TestTargetGasLimitValidationScript(t *testing.T) {
	const valid = `{"jsonrpc":"2.0","id":1,"result":{"payloadStatus":{"status":"VALID","latestValidHash":"0x1111","validationError":null},"payloadId":"0x0102030405060708"}}`
	cases := []struct {
		name, response string
		fail           bool
	}{
		{"same", valid, false},
		{"different-id", strings.ReplaceAll(valid, "0102030405060708", "aabbccddeeff0011"), false},
		{"reordered-keys", `{"result":{"payloadId":"0xaabbccddeeff0011","payloadStatus":{"validationError":null,"latestValidHash":"0x1111","status":"VALID"}},"id":1,"jsonrpc":"2.0"}`, false},
		{"null-id", strings.ReplaceAll(valid, `"0x0102030405060708"`, `null`), true},
		{"missing-id", strings.ReplaceAll(valid, `,"payloadId":"0x0102030405060708"`, ``), true},
		{"short-id", strings.ReplaceAll(valid, "0102030405060708", "0102"), true},
		{"long-id", strings.ReplaceAll(valid, "0102030405060708", "010203040506070809"), true},
		{"uppercase-id", strings.ReplaceAll(valid, "0102030405060708", "AABBCCDDEEFF0011"), true},
		{"non-hex-id", strings.ReplaceAll(valid, "0102030405060708", "zzzzzzzzzzzzzzzz"), true},
		{"numeric-id", strings.ReplaceAll(valid, `"0x0102030405060708"`, `123`), true},
		{"syncing", strings.ReplaceAll(valid, "VALID", "SYNCING"), true},
		{"invalid", strings.ReplaceAll(valid, "VALID", "INVALID"), true},
		{"wrong-ancestor", strings.ReplaceAll(valid, "0x1111", "0x2222"), true},
		{"missing-ancestor", strings.ReplaceAll(valid, `"latestValidHash":"0x1111",`, ``), true},
		{"validation-error", strings.ReplaceAll(valid, `"validationError":null`, `"validationError":"bad payload"`), true},
		{"missing-validation-error", strings.ReplaceAll(valid, `,"validationError":null`, ``), true},
		{"wrong-envelope-id", strings.ReplaceAll(valid, `"id":1`, `"id":2`), true},
		{"wrong-version", strings.ReplaceAll(valid, `"jsonrpc":"2.0"`, `"jsonrpc":"1.0"`), true},
		{"extra-field", strings.ReplaceAll(valid, `"result":{`, `"result":{"extra":true,`), true},
		{"error-and-result", strings.ReplaceAll(valid, `"result":{`, `"error":{"code":-32602},"result":{`), true},
		{"rpc-error", `{"jsonrpc":"2.0","id":1,"error":{"code":-38003,"message":"Invalid payload attributes"}}`, true},
	}
	for _, fixture := range EngineForkchoiceUpdatedV4.Tests {
		t.Run(fixture.Name, func(t *testing.T) {
			if !fixture.SpecOnly || fixture.ValidationScript == "" {
				t.Fatal("client-specific IDs require speconly with a validation script")
			}
			file, err := iofile.Load(fixture.Name, strings.NewReader(">> {\"jsonrpc\":\"2.0\",\"id\":1,\"method\":\"engine_forkchoiceUpdatedV4\",\"params\":[]}\n<< "+valid+"\n--\n"+fixture.ValidationScript))
			if err != nil {
				t.Fatal(err)
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					err := file.RunScript(iofile.ScriptConfig{Log: t}, []json.RawMessage{json.RawMessage(tc.response)})
					if (err != nil) != tc.fail {
						t.Fatalf("error = %v, want failure %v", err, tc.fail)
					}
				})
			}
		})
	}
}

func (e *fixtureEngine) ForkchoiceUpdatedV4(state map[string]interface{}, attrs map[string]interface{}, custody *interface{}) (interface{}, error) {
	e.target, _ = attrs["targetGasLimit"].(string)
	return map[string]interface{}{
		"payloadStatus": map[string]interface{}{"status": e.status, "latestValidHash": e.head, "validationError": nil},
		"payloadId":     e.id,
	}, nil
}

func TestTargetGasLimitGenerator(t *testing.T) {
	fork, slot := uint64(0), uint64(60)
	head := types.NewBlockWithHeader(&types.Header{Number: big.NewInt(60), Time: 600, SlotNumber: &slot})
	chain := &Chain{
		genesis: core.Genesis{Config: &params.ChainConfig{LondonBlock: big.NewInt(0), AmsterdamTime: &fork}},
		blocks:  []*types.Block{head},
	}
	for _, tc := range []struct {
		name, status string
		id           interface{}
		fail         bool
	}{
		{"valid", "VALID", "0x0102030405060708", false},
		{"syncing", "SYNCING", "0x0102030405060708", true},
		{"invalid", "INVALID", "0x0102030405060708", true},
		{"no-build", "VALID", nil, true},
		{"short-id", "VALID", "0x0102", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			api := &fixtureEngine{head: head.Hash(), status: tc.status, id: tc.id}
			server := rpc.NewServer()
			if err := server.RegisterName("engine", api); err != nil {
				t.Fatal(err)
			}
			defer server.Stop()
			client := rpc.DialInProc(server)
			defer client.Close()
			fixture := targetGasLimitTest("test", "0xffffffffffffffff")
			err := fixture.Run(context.Background(), NewT(&fixture, client, chain))
			if (err != nil) != tc.fail {
				t.Fatalf("error = %v, want failure %v", err, tc.fail)
			}
			if api.target != "0xffffffffffffffff" {
				t.Fatalf("target changed before transmission: %s", api.target)
			}
		})
	}
	chain.genesis.Config.AmsterdamTime = nil
	fixture := targetGasLimitTest("test", "0x1")
	err := fixture.Run(context.Background(), NewT(&fixture, nil, chain))
	if err == nil || !strings.Contains(err.Error(), "Amsterdam test chain") {
		t.Fatalf("expected explicit fork prerequisite, got %v", err)
	}
}
