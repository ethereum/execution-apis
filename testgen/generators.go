package testgen

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type T struct {
	eth   *ethclient.Client
	rpc   *rpc.Client
	chain *core.BlockChain
}

func NewT(eth *ethclient.Client, rpc *rpc.Client, chain *core.BlockChain) *T {
	return &T{eth, rpc, chain}
}

// MethodTests is a collection of tests for a certain JSON-RPC method.
type MethodTests struct {
	MethodName string
	Tests      []Test
}

// Test is a wrapper for a function that performs an interaction with the
// client.
type Test struct {
	Name  string
	About string
	Run   func(context.Context, *T) error
}

// AllMethods is a slice of all JSON-RPC methods with tests.
var AllMethods = []MethodTests{
	EthBlockNumber,
	EthGetBlockByNumber,
	DebugGetHeaderRLP,
	DebugGetBlockRLP,
	DebugGetReceiptsRaw,
}

// EthBlockNumber stores a list of all tests against the method.
var EthBlockNumber = MethodTests{
	"eth_blockNumber",
	[]Test{
		{
			"simple-test",
			"retrieves the client's current block number",
			func(ctx context.Context, t *T) error {
				got, err := t.eth.BlockNumber(ctx)
				if err != nil {
					return err
				} else if want := t.chain.CurrentHeader().Number.Uint64(); got != want {
					return fmt.Errorf("unexpect current block number (got: %d, want: %d)", got, want)
				}
				return nil
			},
		},
	},
}

// EthGetBlockByNumber stores a list of all tests against the method.
var EthGetBlockByNumber = MethodTests{
	"eth_getBlockByNumber",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, common.Big0)
				if err != nil {
					return err
				}
				if n := block.Number().Uint64(); n != 0 {
					return fmt.Errorf("expected block 0, got block %d", n)
				}
				return nil
			},
		},
		{
			"get-block-n",
			"gets block 2",
			func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, common.Big2)
				if err != nil {
					return err
				}
				if n := block.Number().Uint64(); n != 2 {
					return fmt.Errorf("expected block 2, got block %d", n)
				}
				return nil
			},
		},
	},
}

var DebugGetHeaderRLP = MethodTests{
	"debug_getHeaderRLP",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getHeaderRLP", "0x0"); err != nil {
					return err
				}
				return checkHeaderRLP(t, 0, got)
			},
		},
		{
			"get-block-n",
			"gets non-zero block",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getHeaderRLP", "0x3"); err != nil {
					return err
				}
				return checkHeaderRLP(t, 3, got)
			},
		},
		{
			"get-invalid-number",
			"gets block with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getHeaderRLP", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetBlockRLP = MethodTests{
	"debug_getBlockRLP",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getBlockRLP", "0x0"); err != nil {
					return err
				}
				return checkBlockRLP(t, 0, got)
			},
		},
		{
			"get-block-n",
			"gets non-zero block",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getBlockRLP", "0x3"); err != nil {
					return err
				}
				return checkBlockRLP(t, 3, got)
			},
		},
		{
			"get-invalid-number",
			"gets block with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getBlockRLP", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetReceiptsRaw = MethodTests{
	"debug_getReceiptsRaw",
	[]Test{
		{
			"get-genesis",
			"gets receipts for block 0",
			func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getReceiptsRaw", "0x0")
			},
		},
		{
			"get-block-n",
			"gets receipts non-zero block",
			func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getReceiptsRaw", "0x3")
			},
		},
		{
			"get-invalid-number",
			"gets receipts with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getReceiptsRaw", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}
