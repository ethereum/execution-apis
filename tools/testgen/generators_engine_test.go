package testgen

import (
	"context"
	"math/big"
	"strings"
	"testing"

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

func (e *fixtureEngine) ForkchoiceUpdatedV4(state map[string]interface{}, attrs map[string]interface{}, custody interface{}) (interface{}, error) {
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
			err := fixture.Run(context.Background(), NewT(client, chain))
			if (err != nil) != tc.fail {
				t.Fatalf("error = %v, want failure %v", err, tc.fail)
			}
			if api.target != "0xffffffffffffffff" {
				t.Fatalf("target changed before transmission: %s", api.target)
			}
		})
	}
	chain.genesis.Config.AmsterdamTime = nil
	err := targetGasLimitTest("test", "0x1").Run(context.Background(), NewT(nil, chain))
	if err == nil || !strings.Contains(err.Error(), "Amsterdam test chain") {
		t.Fatalf("expected explicit fork prerequisite, got %v", err)
	}
}
