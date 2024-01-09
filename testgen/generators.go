package testgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

var (
	pk, _ = crypto.HexToECDSA("9c647b8b7c4e7c3490668fb6c11473619db80c93704c70893d3813af4090c39c")
	addr  = crypto.PubkeyToAddress(pk.PublicKey) // 658bdf435d810c91414ec09147daa6db62406379
)

var (
	// This is the address of an existing contract in the chain, which has code and some storage slots.
	contractAddr = common.HexToAddress("0x000f3df6d732807ef1319fb7b8bb8522d0beac02")
)

type T struct {
	eth   *ethclient.Client
	geth  *gethclient.Client
	rpc   *rpc.Client
	chain *Chain
}

func NewT(client *rpc.Client, chain *Chain) *T {
	eth := ethclient.NewClient(client)
	geth := gethclient.New(client)
	return &T{eth, geth, client, chain}
}

// MethodTests is a collection of tests for a certain JSON-RPC method.
type MethodTests struct {
	Name  string
	Tests []Test
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
	EthGetBlockByHash,
	EthGetHeaderByNumber,
	EthGetHeaderByHash,
	EthGetProof,
	EthChainID,
	EthGetBalance,
	EthGetCode,
	EthGetStorage,
	EthCall,
	EthEstimateGas,
	EthCreateAccessList,
	EthGetBlockTransactionCountByNumber,
	EthGetBlockTransactionCountByHash,
	EthGetTransactionByBlockHashAndIndex,
	EthGetTransactionByBlockNumberAndIndex,
	EthGetTransactionCount,
	EthGetTransactionByHash,
	EthGetTransactionReceipt,
	EthGetBlockReceipts,
	EthSendRawTransaction,
	EthGasPrice,
	EthMaxPriorityFeePerGas,
	EthSyncing,
	EthFeeHistory,
	// EthGetUncleByBlockNumberAndIndex,
	DebugGetRawHeader,
	DebugGetRawBlock,
	DebugGetRawReceipts,
	DebugGetRawTransaction,
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
				} else if want := t.chain.Head().NumberU64(); got != want {
					return fmt.Errorf("unexpect current block number (got: %d, want: %d)", got, want)
				}
				return nil
			},
		},
	},
}

// EthChainID stores a list of all tests against the method.
var EthChainID = MethodTests{
	"eth_chainId",
	[]Test{
		{
			"get-chain-id",
			"retrieves the client's current chain id",
			func(ctx context.Context, t *T) error {
				got, err := t.eth.ChainID(ctx)
				if err != nil {
					return err
				} else if want := t.chain.Config().ChainID.Uint64(); got.Uint64() != want {
					return fmt.Errorf("unexpect chain id (got: %d, want: %d)", got, want)
				}
				return nil
			},
		},
	},
}

// EthGetHeaderByNumber stores a list of all tests against the method.
var EthGetHeaderByNumber = MethodTests{
	"eth_getHeaderByNumber",
	[]Test{
		{
			"get-header-by-number",
			"gets a header by number",
			func(ctx context.Context, t *T) error {
				var got *types.Header
				err := t.rpc.CallContext(ctx, got, "eth_getHeaderByNumber", "0x1")
				if err != nil {
					return err
				}
				want := t.chain.GetBlock(1)
				if reflect.DeepEqual(got, want.Header()) {
					return fmt.Errorf("unexpected header (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
	},
}

// EthGetHeaderByHash stores a list of all tests against the method.
var EthGetHeaderByHash = MethodTests{
	"eth_getHeaderByHash",
	[]Test{
		{
			"get-header-by-hash",
			"gets a header by hash",
			func(ctx context.Context, t *T) error {
				want := t.chain.GetBlock(1).Header()
				var got *types.Header
				err := t.rpc.CallContext(ctx, got, "eth_getHeaderByHash", want.Hash())
				if err != nil {
					return err
				}
				if reflect.DeepEqual(got, want) {
					return fmt.Errorf("unexpected header (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
	},
}

// EthGetCode stores a list of all tests against the method.
var EthGetCode = MethodTests{
	"eth_getCode",
	[]Test{
		{
			"get-code",
			"gets code for 0xaa",
			func(ctx context.Context, t *T) error {
				addr := common.Address{0xaa}
				var got hexutil.Bytes
				err := t.rpc.CallContext(ctx, &got, "eth_getCode", addr, "latest")
				if err != nil {
					return err
				}
				want := t.chain.state[addr].Code
				if !bytes.Equal(got, want) {
					return fmt.Errorf("unexpected code (got: %s, want %s)", got, want)
				}
				return nil
			},
		},
	},
}

// EthGetStorage stores a list of all tests against the method.
var EthGetStorage = MethodTests{
	"eth_getStorage",
	[]Test{
		{
			"get-storage",
			"gets storage for 0xaa",
			func(ctx context.Context, t *T) error {
				addr := common.Address{0xaa}
				key := common.Hash{0x01}
				got, err := t.eth.StorageAt(ctx, addr, key, nil)
				if err != nil {
					return err
				}
				want := t.chain.Storage(addr, key)
				if !bytes.Equal(got, want) {
					return fmt.Errorf("unexpected storage value (got: %s, want %s)", got, want)
				}
				return nil
			},
		},
	},
}

// EthGetBlockByHash stores a list of all tests against the method.
var EthGetBlockByHash = MethodTests{
	"eth_getBlockByHash",
	[]Test{
		{
			"get-block-by-hash",
			"gets block 1",
			func(ctx context.Context, t *T) error {
				want := t.chain.GetBlock(1).Header()
				got, err := t.eth.BlockByHash(ctx, want.Hash())
				if err != nil {
					return err
				}
				if got.Hash() != want.Hash() {
					return fmt.Errorf("unexpected block (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
		{
			"get-block-by-empty-hash",
			"gets block empty hash",
			func(ctx context.Context, t *T) error {
				_, err := t.eth.BlockByHash(ctx, common.Hash{})
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
		{
			"get-block-by-notfound-hash",
			"gets block not found hash",
			func(ctx context.Context, t *T) error {
				_, err := t.eth.BlockByHash(ctx, common.HexToHash("deadbeef"))
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
	},
}

// EthChainID stores a list of all tests against the method.
var EthGetBalance = MethodTests{
	"eth_getBalance",
	[]Test{
		{
			"get-balance",
			"retrieves the an account's balance",
			func(ctx context.Context, t *T) error {
				addr := common.Address{0xaa}
				got, err := t.eth.BalanceAt(ctx, addr, nil)
				if err != nil {
					return err
				}
				want := t.chain.Balance(addr)
				if got.Cmp(want) != 0 {
					return fmt.Errorf("unexpect balance (got: %d, want: %d)", got, want)
				}
				return nil
			},
		},
		{
			"get-balance-blockhash",
			"retrieves the an account's balance at a specific blockhash",
			func(ctx context.Context, t *T) error {
				var (
					block = t.chain.GetBlock(1)
					addr  = common.Address{0xaa}
					got   hexutil.Big
				)
				if err := t.rpc.CallContext(ctx, &got, "eth_getBalance", addr, block.Hash()); err != nil {
					return err
				}
				// TODO: fix
				// state, _ := t.chain.StateAt(block.Root())
				// want := state.GetBalance(addr)
				// if got.ToInt().Uint64() != want.Uint64() {
				// 	return fmt.Errorf("unexpect balance (got: %d, want: %d)", got.ToInt(), want)
				// }
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
				block, err := t.eth.BlockByNumber(ctx, big.NewInt(0))
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
			"get-latest",
			"gets block latest",
			func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, nil)
				if err != nil {
					return err
				}
				if n := block.Number().Uint64(); n != 9 {
					return fmt.Errorf("expected block 9, got block %d", n)
				}
				return nil
			},
		},
		{
			"get-safe",
			"gets block safe",
			func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, big.NewInt(int64(rpc.SafeBlockNumber)))
				if err != nil {
					return err
				}
				if n := block.Number().Uint64(); n != 9 {
					return fmt.Errorf("expected block 9, got block %d", n)
				}
				return nil
			},
		},
		{
			"get-finalized",
			"gets block finalized",
			func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
				if err != nil {
					return err
				}
				if n := block.Number().Uint64(); n != 9 {
					return fmt.Errorf("expected block 9, got block %d", n)
				}
				return nil
			},
		},
		{
			"get-block-n",
			"gets block 2",
			func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, big.NewInt(2))
				if err != nil {
					return err
				}
				if n := block.Number().Uint64(); n != 2 {
					return fmt.Errorf("expected block 2, got block %d", n)
				}
				return nil
			},
		},
		{
			"get-block-notfound",
			"gets block notfound",
			func(ctx context.Context, t *T) error {
				_, err := t.eth.BlockByNumber(ctx, big.NewInt(1000))
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("get a non-existent block should return notfound")
				}
				return nil
			},
		},
	},
}

// EthCall stores a list of all tests against the method.
var EthCall = MethodTests{
	"eth_call",
	[]Test{
		{
			"call-simple-transfer",
			"simulates a simple transfer",
			func(ctx context.Context, t *T) error {
				msg := ethereum.CallMsg{From: common.Address{0xaa}, To: &common.Address{0x01}, Gas: 100000}
				got, err := t.eth.CallContract(ctx, msg, nil)
				if err != nil {
					return err
				}
				if len(got) != 0 {
					return fmt.Errorf("unexpected return value (got: %s, want: nil)", hexutil.Bytes(got))
				}
				return nil
			},
		},
		{
			"call-simple-contract",
			"simulates a simple contract call with no return",
			func(ctx context.Context, t *T) error {
				aa := common.Address{0xaa}
				msg := ethereum.CallMsg{From: aa, To: &aa}
				got, err := t.eth.CallContract(ctx, msg, nil)
				if err != nil {
					return err
				}
				if len(got) != 0 {
					return fmt.Errorf("unexpected return value (got: %s, want: nil)", hexutil.Bytes(got))
				}
				return nil
			},
		},
	},
}

// EthEstimateGas stores a list of all tests against the method.
var EthEstimateGas = MethodTests{
	"eth_estimateGas",
	[]Test{
		{
			"estimate-simple-transfer",
			"estimates a simple transfer",
			func(ctx context.Context, t *T) error {
				msg := ethereum.CallMsg{From: common.Address{0xaa}, To: &common.Address{0x01}}
				got, err := t.eth.EstimateGas(ctx, msg)
				if err != nil {
					return err
				}
				if got != params.TxGas {
					return fmt.Errorf("unexpected return value (got: %d, want: %d)", got, params.TxGas)
				}
				return nil
			},
		},
		{
			"estimate-simple-contract",
			"estimates a simple contract call with no return",
			func(ctx context.Context, t *T) error {
				aa := common.Address{0xaa}
				msg := ethereum.CallMsg{From: aa, To: &aa}
				got, err := t.eth.EstimateGas(ctx, msg)
				if err != nil {
					return err
				}
				want := params.TxGas + 3
				if got != want {
					return fmt.Errorf("unexpected return value (got: %d, want: %d)", got, want)
				}
				return nil
			},
		},
	},
}

// EthEstimateGas stores a list of all tests against the method.
var EthCreateAccessList = MethodTests{
	"eth_createAccessList",
	[]Test{
		{
			"create-al-simple-transfer",
			"estimates a simple transfer",
			func(ctx context.Context, t *T) error {
				msg := make(map[string]interface{})
				msg["from"] = addr
				msg["to"] = common.Address{0x01}

				got := make(map[string]interface{})
				err := t.rpc.CallContext(ctx, &got, "eth_createAccessList", msg, "latest")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			"create-al-simple-contract",
			"estimates a simple contract call with no return",
			func(ctx context.Context, t *T) error {
				msg := make(map[string]interface{})
				msg["from"] = addr
				msg["to"] = common.Address{0xaa}

				got := make(map[string]interface{})
				err := t.rpc.CallContext(ctx, &got, "eth_createAccessList", msg, "latest")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			"create-al-multiple-reads",
			"estimates a simple contract call with no return",
			func(ctx context.Context, t *T) error {
				msg := make(map[string]interface{})
				msg["from"] = addr
				msg["to"] = common.Address{0xbb}

				got := make(map[string]interface{})
				err := t.rpc.CallContext(ctx, &got, "eth_createAccessList", msg, "latest")
				if err != nil {
					return err
				}
				return nil
			},
		},
	},
}

// EthGetBlockTransactionCountByNumber stores a list of all tests against the method.
var EthGetBlockTransactionCountByNumber = MethodTests{
	"eth_getBlockTransactionCountByNumber",
	[]Test{
		{
			"get-genesis",
			"gets tx count in block 0",
			func(ctx context.Context, t *T) error {
				var got hexutil.Uint
				err := t.rpc.CallContext(ctx, &got, "eth_getBlockTransactionCountByNumber", hexutil.Uint(0))
				if err != nil {
					return err
				}
				if int(got) != 0 {
					return fmt.Errorf("tx counts don't match (got: %d, want: %d)", int(got), 0)
				}
				return nil
			},
		},
		{
			"get-block-n",
			"gets tx count in a non-empty block",
			func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				var got hexutil.Uint
				err := t.rpc.CallContext(ctx, &got, "eth_getBlockTransactionCountByNumber", hexutil.Uint64(block.NumberU64()))
				if err != nil {
					return err
				}
				want := len(block.Transactions())
				if int(got) != want {
					return fmt.Errorf("tx counts don't match (got: %d, want: %d)", int(got), want)
				}
				return nil
			},
		},
	},
}

// EthGetBlockTransactionCountByHash stores a list of all tests against the method.
var EthGetBlockTransactionCountByHash = MethodTests{
	"eth_getBlockTransactionCountByHash",
	[]Test{
		{
			"get-genesis",
			"gets tx count in block 0",
			func(ctx context.Context, t *T) error {
				block := t.chain.GetBlock(0)
				var got hexutil.Uint
				err := t.rpc.CallContext(ctx, &got, "eth_getBlockTransactionCountByHash", block.Hash())
				if err != nil {
					return err
				}
				if int(got) != 0 {
					return fmt.Errorf("tx counts don't match (got: %d, want: %d)", int(got), 0)
				}
				return nil
			},
		},
		{
			"get-block-n",
			"gets tx count in a non-empty block",
			func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				var got hexutil.Uint
				err := t.rpc.CallContext(ctx, &got, "eth_getBlockTransactionCountByHash", block.Hash())
				if err != nil {
					return err
				}
				want := len(block.Transactions())
				if int(got) != want {
					return fmt.Errorf("tx counts don't match (got: %d, want: %d)", int(got), want)
				}
				return nil
			},
		},
	},
}

// EthGetTransactionByBlockHashAndIndex stores a list of all tests against the method.
var EthGetTransactionByBlockHashAndIndex = MethodTests{
	"eth_getTransactionByBlockNumberAndIndex",
	[]Test{
		{
			"get-block-n",
			"gets tx 0 in a non-empty block",
			func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				var got types.Transaction
				err := t.rpc.CallContext(ctx, &got, "eth_getTransactionByBlockNumberAndIndex", hexutil.Uint64(block.NumberU64()), hexutil.Uint(0))
				if err != nil {
					return err
				}
				want := block.Transactions()[0]
				if got.Hash() != want.Hash() {
					return fmt.Errorf("tx don't match (got: %d, want: %d)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
	},
}

// EthGetTransactionByBlockNumberAndIndex stores a list of all tests against the method.
var EthGetTransactionByBlockNumberAndIndex = MethodTests{
	"eth_getTransactionByBlockHashAndIndex",
	[]Test{
		{
			"get-block-n",
			"gets tx 0 in a non-empty block",
			func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				var got types.Transaction
				err := t.rpc.CallContext(ctx, &got, "eth_getTransactionByBlockHashAndIndex", block.Hash(), hexutil.Uint(0))
				if err != nil {
					return err
				}
				want := block.Transactions()[0]
				if got.Hash() != want.Hash() {
					return fmt.Errorf("tx don't match (got: %d, want: %d)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
	},
}

// EthGetTransactionCount stores a list of all tests against the method.
var EthGetTransactionCount = MethodTests{
	"eth_getTransactionCount",
	[]Test{
		{
			"get-contract-nonce",
			"gets nonce for a known contract account",
			func(ctx context.Context, t *T) error {
				got, err := t.eth.NonceAt(ctx, contractAddr, nil)
				if err != nil {
					return err
				}
				want := t.chain.state[addr].Nonce
				if got != want {
					return fmt.Errorf("unexpected nonce (got: %d, want: %d)", got, want)
				}
				return nil
			},
		},
	},
}

func matchLegacyValueTransfer(i int, tx *types.Transaction) bool {
	return tx.Type() == types.LegacyTxType && tx.To() != nil && len(tx.Data()) == 0
}

func matchLegacyCreate(i int, tx *types.Transaction) bool {
	return tx.Type() == types.LegacyTxType && tx.To() == nil
}

func matchLegacyTxWithInput(i int, tx *types.Transaction) bool {
	return tx.Type() == types.LegacyTxType && len(tx.Data()) > 0
}

// EthGetTransactionByHash stores a list of all tests against the method.
var EthGetTransactionByHash = MethodTests{
	"eth_getTransactionByHash",
	[]Test{
		{
			"get-legacy-tx",
			"gets a legacy transaction",
			func(ctx context.Context, t *T) error {
				want := t.chain.FindTransaction("legacy tx", matchLegacyValueTransfer)
				got, _, err := t.eth.TransactionByHash(ctx, want.Hash())
				if err != nil {
					return err
				}
				if got.Hash() != want.Hash() {
					return fmt.Errorf("tx mismatch (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
		{
			"get-legacy-create",
			"gets a legacy contract create transaction",
			func(ctx context.Context, t *T) error {
				want := t.chain.FindTransaction("legacy create", matchLegacyCreate)
				got, _, err := t.eth.TransactionByHash(ctx, want.Hash())
				if err != nil {
					return err
				}
				if got.Hash() != want.Hash() {
					return fmt.Errorf("tx mismatch (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
		{
			"get-legacy-input",
			"gets a legacy transaction with input data",
			func(ctx context.Context, t *T) error {
				want := t.chain.FindTransaction("legacy tx w/ input", matchLegacyTxWithInput)
				got, _, err := t.eth.TransactionByHash(ctx, want.Hash())
				if err != nil {
					return err
				}
				if got.Hash() != want.Hash() {
					return fmt.Errorf("tx mismatch (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
		{
			"get-dynamic-fee",
			"gets a dynamic fee transaction",
			func(ctx context.Context, t *T) error {
				want := t.chain.FindTransaction("dynamic fee tx", func(i int, tx *types.Transaction) bool {
					return tx.Type() == types.DynamicFeeTxType
				})
				got, _, err := t.eth.TransactionByHash(ctx, want.Hash())
				if err != nil {
					return err
				}
				if got.Hash() != want.Hash() {
					return fmt.Errorf("tx mismatch (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
		{
			"get-access-list",
			"gets an access list transaction",
			func(ctx context.Context, t *T) error {
				want := t.chain.FindTransaction("access list tx", func(i int, tx *types.Transaction) bool {
					return tx.Type() == types.AccessListTxType
				})
				got, _, err := t.eth.TransactionByHash(ctx, want.Hash())
				if err != nil {
					return err
				}
				if got.Hash() != want.Hash() {
					return fmt.Errorf("tx mismatch (got: %s, want: %s)", got.Hash(), want.Hash())
				}
				return nil
			},
		},
		{
			"get-empty-tx",
			"requests the zero transaction hash",
			func(ctx context.Context, t *T) error {
				_, _, err := t.eth.TransactionByHash(ctx, common.Hash{})
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
		{
			"get-notfound-tx",
			"gets a non-existent transaction",
			func(ctx context.Context, t *T) error {
				_, _, err := t.eth.TransactionByHash(ctx, common.HexToHash("deadbeef"))
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
	},
}

// EthGetTransactionReceipt stores a list of all tests against the method.
var EthGetTransactionReceipt = MethodTests{
	"eth_getTransactionReceipt",
	[]Test{
		{
			"get-legacy-receipt",
			"gets the receipt for a legacy value transfer tx",
			func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("legacy tx", matchLegacyValueTransfer)
				receipt, err := t.eth.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					return err
				}
				if receipt.TxHash != tx.Hash() {
					return fmt.Errorf("wrong receipt returned")
				}
				return nil
			},
		},
		{
			"get-legacy-contract",
			"gets a legacy contract create transaction",
			func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("legacy create", matchLegacyCreate)
				receipt, err := t.eth.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					return err
				}
				if receipt.TxHash != tx.Hash() {
					return fmt.Errorf("wrong receipt returned")
				}
				if receipt.ContractAddress == (common.Address{}) {
					return fmt.Errorf("missing created address in receipt")
				}
				return nil
			},
		},
		{
			"get-legacy-input",
			"gets a legacy transaction with input data",
			func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("legacy tx w/ input", matchLegacyTxWithInput)
				receipt, err := t.eth.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					return err
				}
				if receipt.TxHash != tx.Hash() {
					return fmt.Errorf("wrong receipt returned")
				}
				return nil
			},
		},
		{
			"get-dynamic-fee",
			"gets a dynamic fee transaction",
			func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("dynamic fee tx", func(i int, tx *types.Transaction) bool {
					return tx.Type() == types.DynamicFeeTxType
				})
				receipt, err := t.eth.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					return err
				}
				if receipt.TxHash != tx.Hash() {
					return fmt.Errorf("wrong receipt returned")
				}
				if receipt.Type != types.DynamicFeeTxType {
					return fmt.Errorf("wrong tx type in receipt")
				}
				return nil
			},
		},
		{
			"get-access-list",
			"gets an access list transaction",
			func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("access list tx", func(i int, tx *types.Transaction) bool {
					return tx.Type() == types.AccessListTxType
				})
				receipt, err := t.eth.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					return err
				}
				if receipt.TxHash != tx.Hash() {
					return fmt.Errorf("wrong receipt returned")
				}
				if receipt.Type != types.AccessListTxType {
					return fmt.Errorf("wrong tx type in receipt")
				}
				return nil
			},
		},
		{
			"get-empty-tx",
			"requests the receipt for the zero tx hash",
			func(ctx context.Context, t *T) error {
				_, err := t.eth.TransactionReceipt(ctx, common.Hash{})
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
		{
			"get-notfound-tx",
			"requests the receipt for a non-existent tx hash",
			func(ctx context.Context, t *T) error {
				_, err := t.eth.TransactionReceipt(ctx, common.HexToHash("deadbeef"))
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
	},
}

var EthGetBlockReceipts = MethodTests{
	"eth_getBlockReceipts",
	[]Test{
		{
			"get-block-receipts-0",
			"gets receipts for block 0",
			func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", hexutil.Uint64(0)); err != nil {
					return err
				}
				// Unfortunately, receipts cannot be checked for correctness.
				return nil
			},
		},
		{
			"get-block-receipts-n",
			"gets receipts non-zero block",
			func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", hexutil.Uint64(block.NumberU64())); err != nil {
					return err
				}
				return nil
			},
		},
		{
			"get-block-receipts-future",
			"gets receipts of future block",
			func(ctx context.Context, t *T) error {
				var (
					receipts []*types.Receipt
					future   = t.chain.Head().NumberU64() + 1
				)
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", hexutil.Uint64(future)); err != nil {
					return err
				}
				if len(receipts) != 0 {
					return fmt.Errorf("expected not found, got: %d receipts)", len(receipts))
				}
				return nil
			},
		},
		{
			"get-block-receipts-earliest",
			"gets receipts for block earliest",
			func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", "earliest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			"get-block-receipts-latest",
			"gets receipts for block latest",
			func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			"get-block-receipts-empty",
			"gets receipts for empty block hash",
			func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", common.Hash{}); err != nil {
					return err
				}
				if len(receipts) != 0 {
					return fmt.Errorf("expected not found, got: %d receipts)", len(receipts))
				}
				return nil
			},
		},
		{
			"get-block-receipts-not-found",
			"gets receipts for notfound hash",
			func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", common.HexToHash("deadbeef")); err != nil {
					return err
				}
				if len(receipts) != 0 {
					return fmt.Errorf("expected not found, got: %d receipts)", len(receipts))
				}
				return nil
			},
		},
		{
			"get-block-receipts-by-hash",
			"gets receipts for normal block hash",
			func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", block.Hash()); err != nil {
					return err
				}
				return nil
			},
		},
	},
}

// EthSendRawTransaction stores a list of all tests against the method.
var EthSendRawTransaction = MethodTests{
	"eth_sendRawTransaction",
	[]Test{
		{
			"send-legacy-transaction",
			"sends a raw legacy transaction",
			func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				head := t.chain.Head()
				txdata := &types.LegacyTx{
					Nonce:    nonce,
					To:       &common.Address{0xaa},
					Value:    big.NewInt(10),
					Gas:      25000,
					GasPrice: new(big.Int).Add(head.BaseFee(), big.NewInt(1)),
					Data:     common.FromHex("5544"),
				}
				tx := t.chain.MustSignTx(sender, txdata)
				if err := t.eth.SendTransaction(ctx, tx); err != nil {
					return err
				}
				t.chain.IncNonce(sender, 1)
				return nil
			},
		},
		{
			"send-dynamic-fee-transaction",
			"sends a transaction with dynamic fee",
			func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				basefee := t.chain.Head().BaseFee()
				basefee.Add(basefee, big.NewInt(500))
				txdata := &types.DynamicFeeTx{
					Nonce:     nonce,
					To:        nil,
					Gas:       60000,
					Value:     big.NewInt(42),
					GasTipCap: big.NewInt(500),
					GasFeeCap: basefee,
					Data:      common.FromHex("0x3d602d80600a3d3981f3363d3d373d3d3d363d734d11c446473105a02b5c1ab9ebe9b03f33902a295af43d82803e903d91602b57fd5bf3"), // eip1167.minimal.proxy
				}
				tx := t.chain.MustSignTx(sender, txdata)
				if err := t.eth.SendTransaction(ctx, tx); err != nil {
					return err
				}
				t.chain.IncNonce(sender, 1)
				return nil
			},
		},
		{
			"send-access-list-transaction",
			"sends a transaction with access list",
			func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				basefee := t.chain.Head().BaseFee()
				basefee.Add(basefee, big.NewInt(500))
				txdata := &types.AccessListTx{
					Nonce:    nonce,
					To:       &contractAddr,
					Gas:      90000,
					GasPrice: basefee,
					Data:     common.FromHex("0xa9059cbb000000000000000000000000cff33720980c026cc155dcb366861477e988fd870000000000000000000000000000000000000000000000000000000002fd6892"), // transfer(address to, uint256 value)
					AccessList: types.AccessList{
						{Address: contractAddr, StorageKeys: []common.Hash{{0}, {1}}},
					},
				}
				tx := t.chain.MustSignTx(sender, txdata)
				if err := t.eth.SendTransaction(ctx, tx); err != nil {
					return err
				}
				t.chain.IncNonce(sender, 1)
				return nil
			},
		},
		{
			"send-dynamic-fee-access-list-transaction",
			"sends a transaction with dynamic fee and access list",
			func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				basefee := t.chain.Head().BaseFee()
				basefee.Add(basefee, big.NewInt(500))
				txdata := &types.DynamicFeeTx{
					Nonce:     nonce,
					To:        &contractAddr,
					Gas:       80000,
					GasTipCap: big.NewInt(500),
					GasFeeCap: basefee,
					Data:      common.FromHex("0xa9059cbb000000000000000000000000cff33720980c026cc155dcb366861477e988fd870000000000000000000000000000000000000000000000000000000002fd6892"), // transfer(address to, uint256 value)
					AccessList: types.AccessList{
						{Address: contractAddr, StorageKeys: []common.Hash{{0}, {1}}},
					},
				}
				tx := t.chain.MustSignTx(sender, txdata)
				if err := t.eth.SendTransaction(ctx, tx); err != nil {
					return err
				}
				t.chain.IncNonce(sender, 1)
				return nil
			},
		},
	},
}

// EthGasPrice stores a list of all tests against the method.
var EthGasPrice = MethodTests{
	"eth_gasPrice",
	[]Test{
		{
			"get-current-gas-price",
			"gets the current gas price in wei",
			func(ctx context.Context, t *T) error {
				if _, err := t.eth.SuggestGasPrice(ctx); err != nil {
					return err
				}
				return nil
			},
		},
	},
}

// EthMaxPriorityFeePerGas stores a list of all tests against the method.
var EthMaxPriorityFeePerGas = MethodTests{
	"eth_maxPriorityFeePerGas",
	[]Test{
		{
			"get-current-tip",
			"gets the current maxPriorityFeePerGas in wei",
			func(ctx context.Context, t *T) error {
				if _, err := t.eth.SuggestGasTipCap(ctx); err != nil {
					return err
				}
				return nil
			},
		},
	},
}

// EthFeeHistory stores a list of all tests against the method.
var EthFeeHistory = MethodTests{
	"eth_feeHistory",
	[]Test{
		{
			"fee-history",
			"gets fee history information",
			func(ctx context.Context, t *T) error {
				// Find a block/tx where the London fork is enabled.
				var dftx *types.Transaction
				block := t.chain.BlockWithTransactions("dynamic fee tx", func(i int, tx *types.Transaction) bool {
					if tx.Type() == types.DynamicFeeTxType {
						dftx = tx
						return true
					}
					return false
				})
				got, err := t.eth.FeeHistory(ctx, 1, block.Number(), []float64{95, 99})
				if err != nil {
					return err
				}
				tip, err := dftx.EffectiveGasTip(block.BaseFee())
				if err != nil {
					return fmt.Errorf("unable to get effective tip: %w", err)
				}

				if len(got.Reward) != 1 {
					return fmt.Errorf("mismatch number of rewards (got: %d, want: 1", len(got.Reward))
				}
				if got.Reward[0][0].Cmp(tip) != 0 {
					return fmt.Errorf("mismatch reward value (got: %d, want: %d)", got.Reward[0][0], tip)
				}
				return nil
			},
		},
	},
}

// EthSyncing stores a list of all tests against the method.
var EthSyncing = MethodTests{
	"eth_syncing",
	[]Test{
		{
			"check-syncing",
			"checks client syncing status",
			func(ctx context.Context, t *T) error {
				_, err := t.eth.SyncProgress(ctx)
				if err != nil {
					return err
				}
				return nil
			},
		},
	},
}

// EthGetUncleByBlockNumberAndIndex stores a list of all tests against the method.
var EthGetUncleByBlockNumberAndIndex = MethodTests{
	"eth_getUncleByBlockNumberAndIndex",
	[]Test{
		{
			"get-uncle",
			"gets uncle header",
			func(ctx context.Context, t *T) error {
				var got *types.Header
				t.rpc.CallContext(ctx, got, "eth_getUncleByBlockNumberAndIndex", hexutil.Uint(2), hexutil.Uint(0))
				want := t.chain.GetBlock(2).Uncles()[0]
				if got.Hash() != want.Hash() {
					return fmt.Errorf("mismatch uncle hash (got: %s, want: %s", got.Hash(), want.Hash())
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
				addr := common.Address{0xaa}
				result, err := t.geth.GetProof(ctx, addr, []string{}, big.NewInt(3))
				if err != nil {
					return err
				}
				balance := t.chain.Balance(addr)
				if result.Balance.Cmp(balance) != 0 {
					return fmt.Errorf("unexpected balance (got: %s, want: %s)", result.Balance, balance)
				}
				return nil
			},
		},
		{
			"get-account-proof-blockhash",
			"gets proof for a certain account at the specified blockhash",
			func(ctx context.Context, t *T) error {
				addr := common.Address{0xaa}
				type accountResult struct {
					Balance *hexutil.Big `json:"balance"`
				}
				var result accountResult
				if err := t.rpc.CallContext(ctx, &result, "eth_getProof", addr, []string{}, t.chain.Head().Hash()); err != nil {
					return err
				}
				balance := t.chain.Balance(addr)
				if result.Balance.ToInt().Cmp(balance) != 0 {
					return fmt.Errorf("unexpected balance (got: %s, want: %s)", result.Balance, balance)
				}
				return nil
			},
		},
		{
			"get-account-proof-with-storage",
			"gets proof for a certain account",
			func(ctx context.Context, t *T) error {
				addr := common.Address{0xaa}
				result, err := t.geth.GetProof(ctx, addr, []string{"0x01"}, big.NewInt(3))
				if err != nil {
					return err
				}
				balance := t.chain.Balance(addr)
				if result.Balance.Cmp(balance) != 0 {
					return fmt.Errorf("unexpected balance (got: %s, want: %s)", result.Balance, balance)
				}
				if len(result.StorageProof) == 0 || len(result.StorageProof[0].Proof) == 0 {
					return fmt.Errorf("expected storage proof")
				}
				return nil
			},
		},
	},
}

var DebugGetRawHeader = MethodTests{
	"debug_getRawHeader",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawHeader", "0x0"); err != nil {
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
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawHeader", "0x3"); err != nil {
					return err
				}
				return checkHeaderRLP(t, 3, got)
			},
		},
		{
			"get-invalid-number",
			"gets block with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getRawHeader", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetRawBlock = MethodTests{
	"debug_getRawBlock",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawBlock", "0x0"); err != nil {
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
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawBlock", "0x3"); err != nil {
					return err
				}
				return checkBlockRLP(t, 3, got)
			},
		},
		{
			"get-invalid-number",
			"gets block with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getRawBlock", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetRawReceipts = MethodTests{
	"debug_getRawReceipts",
	[]Test{
		{
			"get-genesis",
			"gets receipts for block 0",
			func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getRawReceipts", "0x0")
			},
		},
		{
			"get-block-n",
			"gets receipts non-zero block",
			func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getRawReceipts", "0x3")
			},
		},
		{
			"get-invalid-number",
			"gets receipts with invalid number formatting",
			func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "debug_getRawReceipts", "2")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}

var DebugGetRawTransaction = MethodTests{
	"debug_getRawTransaction",
	[]Test{
		{
			"get-tx",
			"gets tx rlp by hash",
			func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				tx := block.Transactions()[0]
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawTransaction", tx.Hash().Hex()); err != nil {
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
				err := t.rpc.CallContext(ctx, &got, "debug_getRawTransaction", "1000000000000000000000000000000000000000000000000000000000000001")
				if !strings.HasPrefix(err.Error(), "invalid argument 0") {
					return err
				}
				return nil
			},
		},
	},
}
