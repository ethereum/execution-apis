package testgen

import (
	"bytes"
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type T struct {
	eth   *ethclient.Client
	geth  *gethclient.Client
	rpc   *rpc.Client
	chain *core.BlockChain
}

func NewT(eth *ethclient.Client, geth *gethclient.Client, rpc *rpc.Client, chain *core.BlockChain) *T {
	return &T{eth, geth, rpc, chain}
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
	EthGetProof,
	DebugGetHeader,
	DebugGetBlock,
	DebugGetReceipts,
	DebugGetTransaction,
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

// EthGetProof stores a list of all tests against the method.
var EthGetProof = MethodTests{
	"eth_getProof",
	[]Test{
		{
			"get-account-proof",
			"gets proof for a certain account",
			func(ctx context.Context, t *T) error {
				addr := common.HexToAddress("aa")
				result, err := t.geth.GetProof(ctx, addr, nil, big.NewInt(3))
				if err != nil {
					return err
				}
				state, _ := t.chain.State()
				balance := state.GetBalance(addr)
				if result.Balance.Cmp(balance) != 0 {
					return fmt.Errorf("unexpected balance (got: %s, want: %s)", result.Balance, balance)
				}
				return nil
			},
		},
		{
			"get-account-proof-with-storage",
			"gets proof for a certain account",
			func(ctx context.Context, t *T) error {
				addr := common.HexToAddress("aa")
				result, err := t.geth.GetProof(ctx, addr, []string{"0x01"}, big.NewInt(3))
				if err != nil {
					return err
				}
				state, _ := t.chain.State()
				balance := state.GetBalance(addr)
				if result.Balance.Cmp(balance) != 0 {
					return fmt.Errorf("unexpected balance (got: %s, want: %s)", result.Balance, balance)
				}
				if len(result.StorageProof) == 0 {
					return fmt.Errorf("expected storage proof")
				}
				return nil
			},
		},
	},
}

var DebugGetHeader = MethodTests{
	"debug_getHeader",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getHeader", "0x0"); err != nil {
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
				if err := t.rpc.CallContext(ctx, &got, "debug_getHeader", "0x3"); err != nil {
					return err
				}
				return checkHeaderRLP(t, 3, got)
			},
		},
		{
			"get-invalid-number",
			"gets block with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getHeader", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetBlock = MethodTests{
	"debug_getBlock",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getBlock", "0x0"); err != nil {
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
				if err := t.rpc.CallContext(ctx, &got, "debug_getBlock", "0x3"); err != nil {
					return err
				}
				return checkBlockRLP(t, 3, got)
			},
		},
		{
			"get-invalid-number",
			"gets block with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getBlock", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetReceipts = MethodTests{
	"debug_getReceipts",
	[]Test{
		{
			"get-genesis",
			"gets receipts for block 0",
			func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getReceipts", "0x0")
			},
		},
		{
			"get-block-n",
			"gets receipts non-zero block",
			func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getReceipts", "0x3")
			},
		},
		{
			"get-invalid-number",
			"gets receipts with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getReceipts", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetTransaction = MethodTests{
	"debug_getTransaction",
	[]Test{
		{
			"get-tx",
			"gets tx rlp by hash",
			func(ctx context.Context, t *T) error {
				tx := t.chain.GetBlockByNumber(1).Transactions()[0]
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getTransaction", tx.Hash().Hex()); err != nil {
					return err
				}
				want, err := tx.MarshalBinary()
				if err != nil {
					return err
				}
				if !bytes.Equal(got, want) {
					return fmt.Errorf("mismatching raw tx (got: %s, want: %s)", hexutil.Bytes(got), hexutil.Bytes(want))
				}
				return nil
			},
		},
		{
			"get-invalid-hash",
			"gets tx with hash missing 0x prefix",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				err := t.rpc.CallContext(ctx, &got, "debug_getTransaction", "1000000000000000000000000000000000000000000000000000000000000001")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}
