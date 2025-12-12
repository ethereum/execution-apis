package testgen

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"
	"slices"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto/kzg4844"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/ethclient/gethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/holiman/uint256"
	"golang.org/x/exp/maps"
)

var (
	emitContract = common.HexToAddress("0x7dcd17433742f4c0ca53122ab541d0ba67fc27df")
	nonAccount   = common.HexToAddress("0xc1cadaffffffffffffffffffffffffffffffffff")
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

	// If SpecOnly is true, the client response doesn't have to match exactly and is
	// checked for spec validity only.
	SpecOnly bool

	Run func(context.Context, *T) error
}

// AllMethods is a slice of all JSON-RPC methods with tests.
var AllMethods = []MethodTests{
	EthBlockNumber,
	EthGetBlockByNumber,
	EthGetBlockByHash,
	EthGetProof,
	EthChainID,
	EthGetBalance,
	EthGetCode,
	EthGetStorage,
	EthCall,
	EthSimulateV1,
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
	EthSyncing,
	EthFeeHistory,
	EthGetLogs,
	DebugGetRawHeader,
	DebugGetRawBlock,
	DebugGetRawReceipts,
	DebugGetRawTransaction,
	EthBlobBaseFee,
	NetVersion,

	// -- gas price tests are disabled because of non-determinism
	// EthGasPrice,
	// EthMaxPriorityFeePerGas,

	// -- uncle APIs are not required anymore after the merge
	// EthGetUncleByBlockNumberAndIndex,
}

// EthBlockNumber stores a list of all tests against the method.
var EthBlockNumber = MethodTests{
	"eth_blockNumber",
	[]Test{
		{
			Name:  "simple-test",
			About: "retrieves the client's current block number",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-chain-id",
			About: "retrieves the client's current chain id",
			Run: func(ctx context.Context, t *T) error {
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

// EthGetCode stores a list of all tests against the method.
var EthGetCode = MethodTests{
	"eth_getCode",
	[]Test{
		{
			Name:  "get-code",
			About: "requests code of an existing contract",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				err := t.rpc.CallContext(ctx, &got, "eth_getCode", emitContract, "latest")
				if err != nil {
					return err
				}
				want := t.chain.state[emitContract].Code
				if !bytes.Equal(got, want) {
					return fmt.Errorf("unexpected code (got: %s, want %s)", got, want)
				}
				return nil
			},
		},
		{
			Name: "get-code-eip7702-delegation",
			About: `requests code of an account that has an EIP-7702 delegation. the server is expected to return
the delegation designator.`,
			Run: func(ctx context.Context, t *T) error {
				account := t.chain.txinfo.EIP7702.Account
				var got hexutil.Bytes
				err := t.rpc.CallContext(ctx, &got, "eth_getCode", account, "latest")
				if err != nil {
					return err
				}
				want := t.chain.state[account].Code
				if !bytes.Equal(got, want) {
					return fmt.Errorf("unexpected code (got: %s, want %s)", got, want)
				}
				return nil
			},
		},
		{
			Name:  "get-code-unknown-account",
			About: "requests code of a non-existent account",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				err := t.rpc.CallContext(ctx, &got, "eth_getCode", nonAccount, "latest")
				if err != nil {
					return err
				}
				if len(got) > 0 {
					return fmt.Errorf("account %v has non-empty code", nonAccount)
				}
				return nil
			},
		},
	},
}

// EthGetStorage stores a list of all tests against the method.
var EthGetStorage = MethodTests{
	"eth_getStorageAt",
	[]Test{
		{
			Name:  "get-storage",
			About: "gets storage of a contract",
			Run: func(ctx context.Context, t *T) error {
				addr := emitContract
				key := common.Hash{}
				got, err := t.eth.StorageAt(ctx, addr, key, nil)
				if err != nil {
					return err
				}
				want := t.chain.Storage(addr, key)
				if !bytes.Equal(got, want) {
					return fmt.Errorf("unexpected storage value (got: %s, want %s)", got, want)
				}
				// Check for any non-zero byte in the value.
				// If it's all-zero, the slot doesn't really exist, indicating a problem with the test itself.
				nz := slices.ContainsFunc(got, func(b byte) bool { return b != 0 })
				if !nz {
					return fmt.Errorf("requested storage slot is zero")
				}
				return nil
			},
		},
		{
			Name:  "get-storage-unknown-account",
			About: "gets storage of a non-existent account",
			Run: func(ctx context.Context, t *T) error {
				key := common.Hash{1}
				got, err := t.eth.StorageAt(ctx, nonAccount, key, nil)
				if err != nil {
					return err
				}
				nz := slices.ContainsFunc(got, func(b byte) bool { return b != 0 })
				if nz {
					return fmt.Errorf("storage is non-empty")
				}
				return nil
			},
		},
		{
			Name:  "get-storage-invalid-key-too-large",
			About: "requests an invalid storage key",
			Run: func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "eth_getStorageAt", "0xaa00000000000000000000000000000000000000", "0x00000000000000000000000000000000000000000000000000000000000000000", "latest")
				if err == nil {
					return fmt.Errorf("expected error")
				}
				return nil
			},
		},
		{
			Name:  "get-storage-invalid-key",
			About: "requests an invalid storage key",
			Run: func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "eth_getStorageAt", "0xaa00000000000000000000000000000000000000", "0xasdf", "latest")
				if err == nil {
					return fmt.Errorf("expected error")
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
			Name:  "get-block-by-hash",
			About: "gets block 1",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-by-empty-hash",
			About: "gets block empty hash",
			Run: func(ctx context.Context, t *T) error {
				_, err := t.eth.BlockByHash(ctx, common.Hash{})
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
		{
			Name:  "get-block-by-notfound-hash",
			About: "gets block not found hash",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-balance",
			About: "retrieves the an account balance",
			Run: func(ctx context.Context, t *T) error {
				addr := emitContract
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
			Name:  "get-balance-unknown-account",
			About: "requests the balance of a non-existent account",
			Run: func(ctx context.Context, t *T) error {
				got, err := t.eth.BalanceAt(ctx, nonAccount, nil)
				if err != nil {
					return err
				}
				if got.Sign() > 0 {
					return fmt.Errorf("account %v has non-zero balance", nonAccount)
				}
				return nil
			},
		},
		{
			Name:  "get-balance-blockhash",
			About: "retrieves the an account's balance at a specific blockhash",
			Run: func(ctx context.Context, t *T) error {
				var (
					block = t.chain.GetBlock(int(t.chain.Head().NumberU64()) - 10)
					addr  = emitContract
					got   hexutil.Big
				)
				if err := t.rpc.CallContext(ctx, &got, "eth_getBalance", addr, block.Hash()); err != nil {
					return err
				}
				// We can't really check the result here because there is no state, but the
				// balance shouldn't be zero.
				if got.ToInt().Sign() <= 0 {
					return errors.New("invalid historical balance, should be > zero")
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
			Name:  "get-genesis",
			About: "gets block number zero",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-latest",
			About: "gets the block with tag \"latest\"",
			Run: func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, nil)
				if err != nil {
					return err
				}
				head := t.chain.Head().NumberU64()
				if n := block.Number().Uint64(); n != head {
					return fmt.Errorf("expected block %d, got block %d", head, n)
				}
				return nil
			},
		},
		{
			Name:  "get-safe",
			About: "get the block with tag \"safe\"",
			Run: func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, big.NewInt(int64(rpc.SafeBlockNumber)))
				if err != nil {
					return err
				}
				head := t.chain.Head().NumberU64()
				if n := block.Number().Uint64(); n != head {
					return fmt.Errorf("expected block %d, got block %d", head, n)
				}
				return nil
			},
		},
		{
			Name:  "get-finalized",
			About: "get the block with tag \"finalized\"",
			Run: func(ctx context.Context, t *T) error {
				block, err := t.eth.BlockByNumber(ctx, big.NewInt(int64(rpc.FinalizedBlockNumber)))
				if err != nil {
					return err
				}
				head := t.chain.Head().NumberU64()
				if n := block.Number().Uint64(); n != head {
					return fmt.Errorf("expected block %d, got block %d", head, n)
				}
				return nil
			},
		},
		{
			Name:  "get-block-london-fork",
			About: "requests a block at the London fork",
			Run: func(ctx context.Context, t *T) error {
				hdr, err := t.eth.HeaderByNumber(ctx, t.chain.config.LondonBlock)
				if err != nil {
					return err
				}
				if hdr.BaseFee == nil {
					return fmt.Errorf("missing basefee in block")
				}
				return nil
			},
		},
		{
			Name:  "get-block-merge-fork",
			About: "requests a block at the merge (Paris) fork",
			Run: func(ctx context.Context, t *T) error {
				hdr, err := t.eth.HeaderByNumber(ctx, t.chain.config.MergeNetsplitBlock)
				if err != nil {
					return err
				}
				if hdr.Difficulty.Sign() > 0 {
					return fmt.Errorf("block difficulty > 0")
				}
				return nil
			},
		},
		{
			Name:  "get-block-shanghai-fork",
			About: "requests a block at the Shanghai fork",
			Run: func(ctx context.Context, t *T) error {
				blocknum := t.chain.BlockAtTime(*t.chain.config.ShanghaiTime).Number()
				hdr, err := t.eth.HeaderByNumber(ctx, blocknum)
				if err != nil {
					return err
				}
				if hdr.WithdrawalsHash == nil {
					return fmt.Errorf("block has no withdrawalsHash")
				}
				return nil
			},
		},
		{
			Name:  "get-block-cancun-fork",
			About: "requests a block at the Cancun fork",
			Run: func(ctx context.Context, t *T) error {
				blocknum := t.chain.BlockAtTime(*t.chain.config.CancunTime).Number()
				b, err := t.eth.HeaderByNumber(ctx, blocknum)
				if err != nil {
					return err
				}
				if b.BlobGasUsed == nil {
					return fmt.Errorf("block has no blobGasUsed")
				}
				return nil
			},
		},
		{
			Name:  "get-block-prague-fork",
			About: "requests a block at the Prague fork",
			Run: func(ctx context.Context, t *T) error {
				blocknum := t.chain.txinfo.EIP7002.Block
				hdr, err := t.eth.HeaderByNumber(ctx, big.NewInt(int64(blocknum)))
				if err != nil {
					return err
				}
				if hdr.RequestsHash == nil || *hdr.RequestsHash == types.EmptyRequestsHash {
					return fmt.Errorf("block hash empty or missing requestsHash")
				}
				return nil
			},
		},
		{
			Name:  "get-block-notfound",
			About: "requests a block number that does not exist",
			Run: func(ctx context.Context, t *T) error {
				_, err := t.eth.BlockByNumber(ctx, big.NewInt(1000))
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("get a non-existent block should return null")
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
			Name:  "call-contract",
			About: "performs a basic contract call with default settings",
			Run: func(ctx context.Context, t *T) error {
				msg := ethereum.CallMsg{
					To: &t.chain.txinfo.CallMeContract.Addr,
					// This is the expected input that makes the call pass.
					// See https://github.com/ethereum/hive/blob/master/cmd/hivechain/contracts/callme.eas
					Data: []byte{0xff, 0x01},
				}
				result, err := t.eth.CallContract(ctx, msg, nil)
				if err != nil {
					return err
				}
				want := []byte{0xff, 0xee}
				if !bytes.Equal(result, want) {
					return fmt.Errorf("unexpected return value (got: %#x, want: %#x)", result, want)
				}
				return nil
			},
		},
		{
			Name: "call-callenv",
			About: `Performs a call to the callenv contract, which echoes the EVM transaction environment.
See https://github.com/ethereum/hive/tree/master/cmd/hivechain/contracts/callenv.eas for the output structure.`,
			Run: func(ctx context.Context, t *T) error {
				msg := ethereum.CallMsg{
					To: &t.chain.txinfo.CallEnvContract.Addr,
				}
				result, err := t.eth.CallContract(ctx, msg, nil)
				if err != nil {
					return err
				}
				if len(result) == 0 {
					return fmt.Errorf("empty call result")
				}
				return nil
			},
		},
		{
			Name: "call-callenv-options-eip1559",
			About: `Performs a call to the callenv contract, which echoes the EVM transaction environment.
This call uses EIP1559 transaction options.
See https://github.com/ethereum/hive/tree/master/cmd/hivechain/contracts/callenv.eas for the output structure.`,
			Run: func(ctx context.Context, t *T) error {
				sender, _ := t.chain.GetSender(1)
				basefee := t.chain.Head().BaseFee()
				basefee.Add(basefee, big.NewInt(1))
				msg := ethereum.CallMsg{
					From:      sender,
					To:        &t.chain.txinfo.CallEnvContract.Addr,
					Gas:       60000,
					GasFeeCap: basefee,
					GasTipCap: big.NewInt(11),
					Value:     big.NewInt(23),
					Data:      []byte{0x33, 0x34, 0x35},
				}
				result, err := t.eth.CallContract(ctx, msg, nil)
				if err != nil {
					return err
				}
				if len(result) == 0 {
					return fmt.Errorf("empty call result")
				}
				return nil
			},
		},
		{
			Name:  "call-eip7702-delegation",
			About: `Performs a call to an account that has an EIP-7702 code delegation.`,
			Run: func(ctx context.Context, t *T) error {
				msg := ethereum.CallMsg{
					To:  &t.chain.txinfo.EIP7702.Account,
					Gas: 100000,
				}
				result, err := t.eth.CallContract(ctx, msg, nil)
				if err != nil {
					return err
				}
				if len(result) == 0 {
					return fmt.Errorf("empty call result")
				}
				expectedOutput := slices.Concat(
					make([]byte, 12),
					t.chain.txinfo.EIP7702.Account[:],
					[]byte("invoked"),
					make([]byte, 25),
				)
				if !bytes.Equal(result, expectedOutput) {
					return fmt.Errorf("wrong return value: %x", result)
				}
				return nil
			},
		},
		{
			Name:  "call-revert-abi-panic",
			About: "calls a contract that reverts with an ABI-encoded Panic(uint) value",
			Run: func(ctx context.Context, t *T) error {
				msg := ethereum.CallMsg{
					To:   &t.chain.txinfo.CallRevertContract.Addr,
					Gas:  100000,
					Data: []byte{0}, // triggers panic(uint) revert
				}
				got, err := t.eth.CallContract(ctx, msg, nil)
				if len(got) != 0 {
					return fmt.Errorf("unexpected return value (got: %s, want: nil)", hexutil.Bytes(got))
				}
				if err == nil {
					return fmt.Errorf("expected error for reverting call")
				}
				return nil
			},
		},
		{
			Name:  "call-revert-abi-error",
			About: "calls a contract that reverts with an ABI-encoded Error(string) value",
			Run: func(ctx context.Context, t *T) error {
				msg := ethereum.CallMsg{
					To:   &t.chain.txinfo.CallRevertContract.Addr,
					Gas:  100000,
					Data: []byte{1}, // triggers error(string) revert
				}
				got, err := t.eth.CallContract(ctx, msg, nil)
				if len(got) != 0 {
					return fmt.Errorf("unexpected return value (got: %s, want: nil)", hexutil.Bytes(got))
				}
				if err == nil {
					return fmt.Errorf("expected error for reverting call")
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
			Name:  "estimate-simple-transfer",
			About: "estimates a simple transfer",
			Run: func(ctx context.Context, t *T) error {
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
			Name:     "estimate-successful-call",
			About:    "estimates a successful contract call",
			SpecOnly: true, // EVM gas estimation is not required to be identical across clients
			Run: func(ctx context.Context, t *T) error {
				caller := common.Address{1, 2, 3}
				callme := t.chain.txinfo.CallMeContract.Addr
				msg := ethereum.CallMsg{
					From: caller,
					To:   &callme,
					// This is the expected input that makes the call pass.
					// See https://github.com/ethereum/hive/blob/master/cmd/hivechain/contracts/callme.eas
					Data: []byte{0xff, 0x01},
				}
				got, err := t.eth.EstimateGas(ctx, msg)
				if err != nil {
					return err
				}
				want := uint64(21270)
				if got != want {
					return fmt.Errorf("unexpected return value (got: %d, want: %d)", got, want)
				}
				return nil
			},
		},
		{
			Name:     "estimate-failed-call",
			About:    "estimates a contract call that reverts",
			SpecOnly: true, // EVM gas estimation is not required to be identical across clients
			Run: func(ctx context.Context, t *T) error {
				caller := common.Address{1, 2, 3}
				callme := t.chain.txinfo.CallMeContract.Addr
				msg := ethereum.CallMsg{
					From: caller,
					To:   &callme,
					Data: []byte{0xff, 0x03, 0x04, 0x05},
				}
				if _, err := t.eth.EstimateGas(ctx, msg); err == nil {
					return fmt.Errorf("expected error for failed contract call")
				}
				return nil
			},
		},
		{
			Name:     "estimate-call-abi-error",
			About:    "estimates a contract call that reverts using Solidity Error(string) data",
			SpecOnly: true, // EVM gas estimation is not required to be identical across clients
			Run: func(ctx context.Context, t *T) error {
				caller := common.Address{1, 2, 3}
				contract := t.chain.txinfo.CallRevertContract.Addr
				msg := ethereum.CallMsg{
					From: caller,
					To:   &contract,
					Data: []byte{1}, // triggers error(string) revert
				}
				if _, err := t.eth.EstimateGas(ctx, msg); err == nil {
					return fmt.Errorf("expected error for failed contract call")
				}
				return nil
			},
		},
		{
			Name:     "estimate-with-eip7702",
			About:    "checks that including an EIP-7720 authorization in the message increases gas",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				to := common.Address{0x01}
				baseMsg := map[string]any{
					"from":  sender,
					"to":    to,
					"value": hexutil.Uint64(1),
					"nonce": hexutil.Uint64(nonce),
				}

				withAuth := map[string]any{
					"type":  "0x4",
					"from":  sender,
					"to":    to,
					"value": hexutil.Uint64(1),
					"nonce": hexutil.Uint64(nonce),
					"authorizationList": []map[string]any{
						{
							"chainId": "0x1",
							"address": "0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
							"nonce":   "0x0",
							"yParity": "0x0",
							"r":       "0x1111111111111111111111111111111111111111111111111111111111111111",
							"s":       "0x2222222222222222222222222222222222222222222222222222222222222222",
						},
					},
				}
				var baseGas, authGas hexutil.Uint64
				if err := t.rpc.CallContext(ctx, &baseGas, "eth_estimateGas", baseMsg); err != nil {
					return fmt.Errorf("base estimation failed: %v", err)
				}
				if err := t.rpc.CallContext(ctx, &authGas, "eth_estimateGas", withAuth); err != nil {
					return fmt.Errorf("with auth estimation failed: %v", err)
				}
				if authGas <= baseGas {
					return fmt.Errorf("expected higher gas with auth (got: %d, base: %d)", authGas, baseGas)
				}
				return nil
			},
		},
		{
			Name:     "estimate-with-eip4844",
			About:    "checks gas estimation for blob transactions (EIP-4844)",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				to := common.Address{0x01}
				msg := map[string]any{
					"type":             "0x3",
					"from":             sender,
					"to":               to,
					"value":            hexutil.Uint64(1),
					"nonce":            hexutil.Uint64(nonce),
					"maxFeePerBlobGas": "0x5",
					"blobVersionedHashes": []string{
						"0x0100000000000000000000000000000000000000000000000000000000000000",
					},
				}
				var gas hexutil.Uint64
				if err := t.rpc.CallContext(ctx, &gas, "eth_estimateGas", msg); err != nil {
					return fmt.Errorf("estimation failed: %v", err)
				}
				if gas < 21000 {
					return fmt.Errorf("expected blob tx to require more than base gas, got %d", gas)
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
			Name:  "create-al-value-transfer",
			About: "estimates a simple transfer",
			Run: func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				msg := map[string]any{
					"from":  sender,
					"to":    common.Address{0x01},
					"value": hexutil.Uint64(10),
					"nonce": hexutil.Uint64(nonce),
				}
				result := make(map[string]any)
				err := t.rpc.CallContext(ctx, &result, "eth_createAccessList", msg, "latest")
				if err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:     "create-al-contract",
			About:    "creates an access list for a contract invocation that accesses storage",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				gasprice := t.chain.Head().BaseFee()
				sender, nonce := t.chain.GetSender(0)
				msg := map[string]any{
					"from":     sender,
					"to":       emitContract,
					"nonce":    hexutil.Uint64(nonce),
					"gas":      hexutil.Uint64(60000),
					"gasPrice": (*hexutil.Big)(gasprice),
					"input":    "0x010203040506",
				}
				var result struct {
					AccessList types.AccessList
				}
				err := t.rpc.CallContext(ctx, &result, "eth_createAccessList", msg, "latest")
				if err != nil {
					return err
				}
				if len(result.AccessList) == 0 {
					return fmt.Errorf("empty access list")
				}
				if result.AccessList[0].Address != emitContract {
					return fmt.Errorf("wrong address in access list entry")
				}
				if len(result.AccessList[0].StorageKeys) == 0 {
					return fmt.Errorf("no storage keys in access list entry")
				}
				return nil
			},
		},
		{
			Name: "create-al-contract-eip1559",
			About: `Creates an access list for a contract invocation that accesses storage.
This invocation uses EIP-1559 fields to specify the gas price.`,
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				gasprice := t.chain.Head().BaseFee()
				sender, nonce := t.chain.GetSender(0)
				msg := map[string]any{
					"from":                 sender,
					"to":                   emitContract,
					"nonce":                hexutil.Uint64(nonce),
					"gas":                  hexutil.Uint64(60000),
					"maxFeePerGas":         (*hexutil.Big)(gasprice),
					"maxPriorityFeePerGas": (*hexutil.Big)(big.NewInt(3)),
					"input":                "0x010203040506",
				}
				var result struct {
					AccessList types.AccessList
				}
				err := t.rpc.CallContext(ctx, &result, "eth_createAccessList", msg, "latest")
				if err != nil {
					return err
				}
				if len(result.AccessList) == 0 {
					return fmt.Errorf("empty access list")
				}
				if result.AccessList[0].Address != emitContract {
					return fmt.Errorf("wrong address in access list entry")
				}
				if len(result.AccessList[0].StorageKeys) == 0 {
					return fmt.Errorf("no storage keys in access list entry")
				}
				return nil
			},
		},
		{
			Name: "create-al-abi-revert",
			About: `Creates an access list for a contract invocation that reverts.
The server should return the accessed slots regardless of failure, and should report the failure
in the "error" field.`,
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
				msg := map[string]any{
					"to":    t.chain.txinfo.CallRevertContract.Addr,
					"gas":   hexutil.Uint64(100000),
					"input": "0x01", // triggers error(string) revert
				}
				var result struct {
					AccessList types.AccessList
					Error      string
				}
				err := t.rpc.CallContext(ctx, &result, "eth_createAccessList", msg, "latest")
				if err != nil {
					return fmt.Errorf("reverting call returned JSON-RPC error")
				}
				if len(result.AccessList) == 0 {
					return fmt.Errorf("empty access list")
				}
				if len(result.Error) == 0 {
					return fmt.Errorf("EVM revert error not signaled in response")
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
			Name:  "get-genesis",
			About: "gets tx count in block 0",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-n",
			About: "gets tx count in a non-empty block",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-genesis",
			About: "gets tx count in block 0",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-n",
			About: "gets tx count in a non-empty block",
			Run: func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("any", nil)
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
			Name:  "get-block-n",
			About: "gets tx 0 in a non-empty block",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-n",
			About: "gets tx 0 in a non-empty block",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-nonce",
			About: "gets nonce for a known account",
			Run: func(ctx context.Context, t *T) error {
				addr := findAccountWithNonce(t.chain)
				got, err := t.eth.NonceAt(ctx, addr, nil)
				if err != nil {
					return err
				}
				want := t.chain.state[addr].Nonce
				if got != want {
					return fmt.Errorf("unexpected nonce (got: %d, want: %d)", got, want)
				}
				if want == 0 {
					return fmt.Errorf("nonce for account %v is zero", addr)
				}
				return nil
			},
		},
		{
			Name: "get-nonce-eip7702-account",
			About: `Retrieves the nonce for an account that has an EIP-7702 code delegation applied.
For such accounts, the nonce stored in state does not match the 'transaction count'.`,
			Run: func(ctx context.Context, t *T) error {
				addr := t.chain.txinfo.EIP7702.Account
				got, err := t.eth.NonceAt(ctx, addr, nil)
				if err != nil {
					return err
				}
				want := t.chain.state[addr].Nonce
				if got != want {
					return fmt.Errorf("unexpected nonce (got: %d, want: %d)", got, want)
				}
				if want == 0 {
					return fmt.Errorf("nonce for account %v is zero", addr)
				}
				return nil
			},
		},
		{
			Name:  "get-nonce-unknown-account",
			About: "gets nonce for a non-existent account",
			Run: func(ctx context.Context, t *T) error {
				got, err := t.eth.NonceAt(ctx, nonAccount, nil)
				if err != nil {
					return err
				}
				want := t.chain.state[nonAccount].Nonce
				if got != want {
					return fmt.Errorf("unexpected nonce (got: %d, want: %d)", got, want)
				}
				if want != 0 {
					return fmt.Errorf("nonce for account %v is non-zero", nonAccount)
				}
				return nil
			},
		},
	},
}

func findAccountWithNonce(c *Chain) common.Address {
	accounts := maps.Keys(c.state)
	slices.SortFunc(accounts, common.Address.Cmp)
	for _, acc := range accounts {
		if c.state[acc].Nonce > 0 {
			return acc
		}
	}
	panic("no account with non-zero nonce found in state")
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
			Name:  "get-legacy-tx",
			About: "gets a legacy transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-legacy-create",
			About: "gets a legacy contract create transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-legacy-input",
			About: "gets a legacy transaction with input data",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-dynamic-fee",
			About: "gets a dynamic fee transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-access-list",
			About: "gets an access list transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-blob-tx",
			About: "gets a blob transaction",
			Run: func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("blob tx", func(i int, tx *types.Transaction) bool {
					return tx.Type() == types.BlobTxType
				})
				got, _, err := t.eth.TransactionByHash(ctx, tx.Hash())
				if err != nil {
					return err
				}
				if got.Hash() != tx.Hash() {
					return fmt.Errorf("tx mismatch (got: %s, want: %s)", got.Hash(), tx.Hash())
				}
				return nil
			},
		},
		{
			Name:  "get-setcode-tx",
			About: "retrieves an EIP-7702 transaction",
			Run: func(ctx context.Context, t *T) error {
				txhash := t.chain.txinfo.EIP7702.AuthorizeTx
				got, _, err := t.eth.TransactionByHash(ctx, txhash)
				if err != nil {
					return err
				}
				if got.Type() != types.SetCodeTxType {
					return fmt.Errorf("transaction is not of type %v", types.SetCodeTxType)
				}
				return nil
			},
		},
		{
			Name:  "get-empty-tx",
			About: "requests the zero transaction hash",
			Run: func(ctx context.Context, t *T) error {
				_, _, err := t.eth.TransactionByHash(ctx, common.Hash{})
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
		{
			Name:  "get-notfound-tx",
			About: "gets a non-existent transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-legacy-receipt",
			About: "gets the receipt for a legacy value transfer tx",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-legacy-contract",
			About: "gets a legacy contract create transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-legacy-input",
			About: "gets a legacy transaction with input data",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-dynamic-fee",
			About: "gets a dynamic fee transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-access-list",
			About: "gets an access list transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-blob-tx",
			About: "gets a blob transaction",
			Run: func(ctx context.Context, t *T) error {
				tx := t.chain.FindTransaction("blob tx", func(i int, tx *types.Transaction) bool {
					return tx.Type() == types.BlobTxType
				})
				receipt, err := t.eth.TransactionReceipt(ctx, tx.Hash())
				if err != nil {
					return err
				}
				if receipt.TxHash != tx.Hash() {
					return fmt.Errorf("wrong receipt returned")
				}
				if receipt.Type != types.BlobTxType {
					return fmt.Errorf("wrong tx type in receipt")
				}
				return nil
			},
		},
		{
			Name:  "get-setcode-tx",
			About: "gets the receipt for a EIP-7702 setcode transaction",
			Run: func(ctx context.Context, t *T) error {
				txhash := t.chain.txinfo.EIP7702.AuthorizeTx
				receipt, err := t.eth.TransactionReceipt(ctx, txhash)
				if err != nil {
					return err
				}
				if receipt.TxHash != txhash {
					return fmt.Errorf("wrong receipt returned")
				}
				if receipt.Type != types.SetCodeTxType {
					return fmt.Errorf("wrong tx type in receipt")
				}
				return nil
			},
		},
		{
			Name:  "get-empty-tx",
			About: "requests the receipt for the zero tx hash",
			Run: func(ctx context.Context, t *T) error {
				_, err := t.eth.TransactionReceipt(ctx, common.Hash{})
				if !errors.Is(err, ethereum.NotFound) {
					return errors.New("expected not found error")
				}
				return nil
			},
		},
		{
			Name:  "get-notfound-tx",
			About: "requests the receipt for a non-existent tx hash",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-receipts-0",
			About: "gets receipts for block 0",
			Run: func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", hexutil.Uint64(0)); err != nil {
					return err
				}
				// Unfortunately, receipts cannot be checked for correctness.
				return nil
			},
		},
		{
			Name:  "get-block-receipts-n",
			About: "gets receipts non-zero block",
			Run: func(ctx context.Context, t *T) error {
				block := t.chain.BlockWithTransactions("", nil)
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", hexutil.Uint64(block.NumberU64())); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "get-block-receipts-future",
			About: "gets receipts of future block",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-receipts-earliest",
			About: "gets receipts for block earliest",
			Run: func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", "earliest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "get-block-receipts-latest",
			About: "gets receipts for block latest",
			Run: func(ctx context.Context, t *T) error {
				var receipts []*types.Receipt
				if err := t.rpc.CallContext(ctx, &receipts, "eth_getBlockReceipts", "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "get-block-receipts-empty",
			About: "gets receipts for empty block hash",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-receipts-not-found",
			About: "gets receipts for notfound hash",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-block-receipts-by-hash",
			About: "gets receipts for normal block hash",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "send-legacy-transaction",
			About: "sends a raw legacy transaction",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "send-dynamic-fee-transaction",
			About: "sends a create transaction with dynamic fee",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "send-access-list-transaction",
			About: "sends a transaction with access list",
			Run: func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				basefee := t.chain.Head().BaseFee()
				basefee.Add(basefee, big.NewInt(500))
				txdata := &types.AccessListTx{
					Nonce:    nonce,
					To:       &emitContract,
					Gas:      90000,
					GasPrice: basefee,
					Data:     common.FromHex("0x010203"),
					AccessList: types.AccessList{
						{Address: emitContract, StorageKeys: []common.Hash{{0}, {1}}},
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
			Name:  "send-dynamic-fee-access-list-transaction",
			About: "sends a transaction with dynamic fee and access list",
			Run: func(ctx context.Context, t *T) error {
				sender, nonce := t.chain.GetSender(0)
				basefee := t.chain.Head().BaseFee()
				basefee.Add(basefee, big.NewInt(500))
				txdata := &types.DynamicFeeTx{
					Nonce:     nonce,
					To:        &emitContract,
					Gas:       80000,
					GasTipCap: big.NewInt(500),
					GasFeeCap: basefee,
					Data:      common.FromHex("0x01020304"),
					AccessList: types.AccessList{
						{Address: emitContract, StorageKeys: []common.Hash{{0}, {1}}},
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
			Name:  "send-blob-tx",
			About: "sends a blob transaction",
			Run: func(ctx context.Context, t *T) error {
				var (
					sender, nonce      = t.chain.GetSender(3)
					basefee            = uint256.MustFromBig(t.chain.Head().BaseFee())
					fee                = uint256.NewInt(500)
					emptyBlob          = kzg4844.Blob{}
					emptyBlobCommit, _ = kzg4844.BlobToCommitment(&emptyBlob)
					emptyBlobProof, _  = kzg4844.ComputeBlobProof(&emptyBlob, emptyBlobCommit)
				)
				fee.Add(basefee, fee)
				sidecar := &types.BlobTxSidecar{
					Blobs:       []kzg4844.Blob{emptyBlob},
					Commitments: []kzg4844.Commitment{emptyBlobCommit},
					Proofs:      []kzg4844.Proof{emptyBlobProof},
				}

				txdata := &types.BlobTx{
					Nonce:     nonce,
					To:        emitContract,
					Gas:       80000,
					GasTipCap: uint256.NewInt(500),
					GasFeeCap: fee,
					Data:      common.FromHex("0xa9059cbb000000000000000000000000cff33720980c026cc155dcb366861477e988fd870000000000000000000000000000000000000000000000000000000002fd6892"), // transfer(address to, uint256 value)
					AccessList: types.AccessList{
						{Address: emitContract, StorageKeys: []common.Hash{{0}, {1}}},
					},
					BlobHashes: sidecar.BlobHashes(),
					BlobFeeCap: uint256.NewInt(params.BlobTxBlobGasPerBlob),
					Sidecar:    sidecar,
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
			Name:  "get-current-gas-price",
			About: "gets the current gas price in wei",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-current-tip",
			About: "gets the current maxPriorityFeePerGas in wei",
			Run: func(ctx context.Context, t *T) error {
				if _, err := t.eth.SuggestGasTipCap(ctx); err != nil {
					return err
				}
				return nil
			},
		},
	},
}

var EthBlobBaseFee = MethodTests{
	"eth_blobBaseFee",
	[]Test{
		{
			Name:  "get-current-blobfee",
			About: "gets the current blob fee in wei",
			Run: func(ctx context.Context, t *T) error {
				var result hexutil.Big
				err := t.rpc.CallContext(ctx, &result, "eth_blobBaseFee")
				return err
			},
		},
	},
}

// EthFeeHistory stores a list of all tests against the method.
var EthFeeHistory = MethodTests{
	"eth_feeHistory",
	[]Test{
		{
			Name:     "fee-history",
			About:    "gets fee history information",
			SpecOnly: true,
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "check-syncing",
			About: "checks client syncing status",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-uncle",
			About: "gets uncle header",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-account-proof-latest",
			About: "requests the account proof for a known account",
			Run: func(ctx context.Context, t *T) error {
				result, err := t.geth.GetProof(ctx, emitContract, []string{}, nil)
				if err != nil {
					return err
				}
				balance := t.chain.Balance(emitContract)
				if result.Balance.Cmp(balance) != 0 {
					return fmt.Errorf("unexpected balance (got: %s, want: %s)", result.Balance, balance)
				}
				if result.Balance.Sign() == 0 {
					return fmt.Errorf("balance is zero, does the account exist?")
				}
				return nil
			},
		},
		{
			Name:  "get-account-proof-blockhash",
			About: "gets proof for a certain account at the specified blockhash",
			Run: func(ctx context.Context, t *T) error {
				type accountResult struct {
					Balance *hexutil.Big `json:"balance"`
				}
				var result accountResult
				if err := t.rpc.CallContext(ctx, &result, "eth_getProof", emitContract, []string{}, t.chain.Head().Hash()); err != nil {
					return err
				}
				balance := t.chain.Balance(emitContract)
				if result.Balance.ToInt().Cmp(balance) != 0 {
					return fmt.Errorf("unexpected balance (got: %s, want: %s)", result.Balance, balance)
				}
				if result.Balance.ToInt().Sign() == 0 {
					return fmt.Errorf("balance is zero, does the account exist?")
				}
				return nil
			},
		},
		{
			Name:  "get-account-proof-with-storage",
			About: "gets proof for a certain account",
			Run: func(ctx context.Context, t *T) error {
				result, err := t.geth.GetProof(ctx, emitContract, []string{"0x00"}, nil)
				if err != nil {
					return err
				}
				balance := t.chain.Balance(emitContract)
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

var EthGetLogs = MethodTests{
	"eth_getLogs",
	[]Test{
		{
			Name:  "no-topics",
			About: "queries for all logs across a range of blocks",
			Run: func(ctx context.Context, t *T) error {
				result, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(3),
				})
				if err != nil {
					return err
				}
				if len(result) == 0 {
					return fmt.Errorf("no logs returned")
				}
				return nil
			},
		},
		{
			Name:  "contract-addr",
			About: "queries for logs from a specific contract across a range of blocks",
			Run: func(ctx context.Context, t *T) error {
				result, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{
					FromBlock: big.NewInt(1),
					ToBlock:   big.NewInt(4),
					Addresses: []common.Address{emitContract},
				})
				if err != nil {
					return err
				}
				if len(result) == 0 {
					return fmt.Errorf("no logs returned")
				}
				bad := slices.ContainsFunc(result, func(l types.Log) bool {
					return l.Address != emitContract
				})
				if bad {
					return fmt.Errorf("result contains log for unrequested contract")
				}
				return nil
			},
		},
		{
			Name:  "topic-exact-match",
			About: "queries for logs with two topics, with both topics set explictly",
			Run: func(ctx context.Context, t *T) error {
				// Find a topic.
				i := slices.IndexFunc(t.chain.txinfo.LegacyEmit, func(tx TxInfo) bool {
					return tx.Block > 2
				})
				if i == -1 {
					return fmt.Errorf("no suitable tx found")
				}
				info := t.chain.txinfo.LegacyEmit[i]
				startBlock := uint64(info.Block - 1)
				endBlock := uint64(info.Block + 2)
				result, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{
					FromBlock: new(big.Int).SetUint64(startBlock),
					ToBlock:   new(big.Int).SetUint64(endBlock),
					Topics:    [][]common.Hash{{*info.LogTopic0}, {*info.LogTopic1}},
				})
				if err != nil {
					return err
				}
				if len(result) != 1 {
					return fmt.Errorf("result contains %d logs, want 1", len(result))
				}
				return nil
			},
		},
		{
			Name:  "topic-wildcard",
			About: "queries for logs with two topics, performing a wildcard match in topic position zero",
			Run: func(ctx context.Context, t *T) error {
				// Find a topic.
				i := slices.IndexFunc(t.chain.txinfo.LegacyEmit, func(tx TxInfo) bool {
					return tx.Block > 2
				})
				if i == -1 {
					return fmt.Errorf("no suitable tx found")
				}
				info := t.chain.txinfo.LegacyEmit[i]
				startBlock := uint64(info.Block - 1)
				endBlock := uint64(info.Block + 2)
				result, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{
					FromBlock: new(big.Int).SetUint64(startBlock),
					ToBlock:   new(big.Int).SetUint64(endBlock),
					Topics:    [][]common.Hash{{}, {*info.LogTopic1}},
				})
				if err != nil {
					return err
				}
				if len(result) != 1 {
					return fmt.Errorf("result contains %d logs, want 1", len(result))
				}
				return nil
			},
		},
		{
			Name:  "filter-with-blockHash",
			About: "queries for all logs of a block, identified by blockHash",
			Run: func(ctx context.Context, t *T) error {
				// Find a block with logs.
				i := slices.IndexFunc(t.chain.txinfo.LegacyEmit, func(tx TxInfo) bool {
					return tx.Block > 2
				})
				if i == -1 {
					return fmt.Errorf("no suitable tx found")
				}
				block := t.chain.GetBlock(int(t.chain.txinfo.LegacyEmit[i].Block))
				hash := block.Hash()
				result, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{BlockHash: &hash})
				if err != nil {
					return err
				}
				if len(result) == 0 {
					return fmt.Errorf("result contains no logs")
				}
				return nil
			},
		},
		{
			Name:  "filter-with-blockHash",
			About: "queries for all logs of a block, identified by blockHash",
			Run: func(ctx context.Context, t *T) error {
				// Find a block with logs.
				i := slices.IndexFunc(t.chain.txinfo.LegacyEmit, func(tx TxInfo) bool {
					return tx.Block > 2
				})
				if i == -1 {
					return fmt.Errorf("no suitable tx found")
				}
				hash := t.chain.GetBlock(int(t.chain.txinfo.LegacyEmit[i].Block)).Hash()
				result, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{BlockHash: &hash})
				if err != nil {
					return err
				}
				if len(result) == 0 {
					return fmt.Errorf("result contains no logs")
				}
				return nil
			},
		},
		{
			Name:  "filter-with-blockHash-and-topics",
			About: "queries for logs in a block, identified by blockHash",
			Run: func(ctx context.Context, t *T) error {
				// Find a block with logs.
				i := slices.IndexFunc(t.chain.txinfo.LegacyEmit, func(tx TxInfo) bool {
					return tx.Block > 2
				})
				if i == -1 {
					return fmt.Errorf("no suitable tx found")
				}
				info := t.chain.txinfo.LegacyEmit[i]
				hash := t.chain.GetBlock(int(info.Block)).Hash()
				result, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{
					BlockHash: &hash,
					Topics:    [][]common.Hash{{*info.LogTopic0}, {*info.LogTopic1}},
				})
				if err != nil {
					return err
				}
				if len(result) != 1 {
					return fmt.Errorf("expected 1 result, got %d", len(result))
				}
				return nil
			},
		},
		{
			Name:  "filter-error-future-block-range",
			About: "checks that an error is returned if `toBlock` is greater than the latest block",
			Run: func(ctx context.Context, t *T) error {
				_, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(len(t.chain.blocks) - 5)),
					ToBlock:   big.NewInt(int64(len(t.chain.blocks) + 1)),
				})
				if err == nil {
					return fmt.Errorf("expected error")
				}
				return nil
			},
		},
		{
			Name:  "filter-error-reversed-block-range",
			About: "checks that an error is returned if `fromBlock` is larger than `toBlock`",
			Run: func(ctx context.Context, t *T) error {
				_, err := t.eth.FilterLogs(ctx, ethereum.FilterQuery{
					FromBlock: big.NewInt(int64(len(t.chain.blocks) - 5)),
					ToBlock:   big.NewInt(int64(len(t.chain.blocks) - 8)),
				})
				if err == nil {
					return fmt.Errorf("expected error")
				}
				return nil
			},
		},
		{
			Name:  "filter-error-blockHash-and-range",
			About: "checks that an error is returned if `fromBlock`/`toBlock` are specified together with `blockHash`",
			Run: func(ctx context.Context, t *T) error {
				err := t.rpc.CallContext(ctx, nil, "eth_getLogs", map[string]string{
					"blockHash": t.chain.blocks[10].Hash().String(),
					"fromBlock": "0x3",
					"toBlock":   "0x4",
				})
				if err == nil {
					return fmt.Errorf("expected error")
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
			Name:  "get-genesis",
			About: "gets block 0",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawHeader", "0x0"); err != nil {
					return err
				}
				return checkHeaderRLP(t, 0, got)
			},
		},
		{
			Name:  "get-block-n",
			About: "gets non-zero block",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawHeader", "0x3"); err != nil {
					return err
				}
				return checkHeaderRLP(t, 3, got)
			},
		},
		{
			Name:  "get-invalid-number",
			About: "gets block with invalid number formatting",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-genesis",
			About: "gets block 0",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawBlock", "0x0"); err != nil {
					return err
				}
				return checkBlockRLP(t, 0, got)
			},
		},
		{
			Name:  "get-block-n",
			About: "gets non-zero block",
			Run: func(ctx context.Context, t *T) error {
				var got hexutil.Bytes
				if err := t.rpc.CallContext(ctx, &got, "debug_getRawBlock", "0x3"); err != nil {
					return err
				}
				return checkBlockRLP(t, 3, got)
			},
		},
		{
			Name:  "get-invalid-number",
			About: "gets block with invalid number formatting",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-genesis",
			About: "gets receipts for block 0",
			Run: func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getRawReceipts", "0x0")
			},
		},
		{
			Name:  "get-block-n",
			About: "gets receipts non-zero block",
			Run: func(ctx context.Context, t *T) error {
				return t.rpc.CallContext(ctx, nil, "debug_getRawReceipts", "0x3")
			},
		},
		{
			Name:  "get-invalid-number",
			About: "gets receipts with invalid number formatting",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-tx",
			About: "gets tx rlp by hash",
			Run: func(ctx context.Context, t *T) error {
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
			Name:  "get-invalid-hash",
			About: "gets tx with hash missing 0x prefix",
			Run: func(ctx context.Context, t *T) error {
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

var NetVersion = MethodTests{
	"net_version",
	[]Test{
		{
			Name:  "get-network-id",
			About: "Calls net_version to retrieve the network ID, which is expected to be equal to the chainID of the test chain.",
			Run: func(ctx context.Context, t *T) error {
				id, err := t.eth.NetworkID(ctx)
				if err != nil {
					return err
				}
				if id.Cmp(t.chain.genesis.Config.ChainID) != 0 {
					return fmt.Errorf("wrong networkID %v returned", id)
				}
				return nil
			},
		},
	},
}

var EthSimulateV1 = MethodTests{
	"eth_simulateV1",
	[]Test{
		{
			Name:  "ethSimulate-blobs",
			About: "simulates a simple blob transaction",

			Run: func(ctx context.Context, t *T) error {
				var (
					emptyBlob          = kzg4844.Blob{}
					emptyBlobCommit, _ = kzg4844.BlobToCommitment(&emptyBlob)
					emptyBlobProof, _  = kzg4844.ComputeBlobProof(&emptyBlob, emptyBlobCommit)
				)
				sidecar := &types.BlobTxSidecar{
					Blobs:       []kzg4844.Blob{emptyBlob},
					Commitments: []kzg4844.Commitment{emptyBlobCommit},
					Proofs:      []kzg4844.Proof{emptyBlobProof},
				}
				blobVersionedhashes := sidecar.BlobHashes()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								BlobBaseFee:   (*hexutil.Big)(big.NewInt(0)),
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(15)),
							},
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(1000000000)},
								common.Address{0xc2}: OverrideAccount{Code: getFirstThreeBlobs()},
							},
							Calls: []TransactionArgs{{
								From:                &common.Address{0xc0},
								To:                  &common.Address{0xc2},
								MaxFeePerGas:        (*hexutil.Big)(big.NewInt(16)),
								MaxFeePerBlobGas:    *newRPCBalance(10),
								BlobVersionedHashes: &blobVersionedhashes,
							}},
						},
						{
							BlockOverrides: &BlockOverrides{
								BlobBaseFee: (*hexutil.Big)(big.NewInt(1)),
							},
							Calls: []TransactionArgs{{
								From:                &common.Address{0xc0},
								To:                  &common.Address{0xc2},
								MaxFeePerGas:        (*hexutil.Big)(big.NewInt(16)),
								MaxFeePerBlobGas:    *newRPCBalance(10),
								BlobVersionedHashes: &blobVersionedhashes,
							}},
						},
					},
					Validation:             true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple",
			About: "simulates a ethSimulate transfer",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(1000)},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							}, {
								From:  &common.Address{0xc1},
								To:    &common.Address{0xc2},
								Value: *newRPCBalance(1000),
							}},
						},
					},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-validation-fulltx",
			About: "simulates a ethSimulate transfer",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(15)),
							},
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(1000000000000)},
							},
							Calls: []TransactionArgs{{
								From:         &common.Address{0xc0},
								To:           &common.Address{0xc1},
								MaxFeePerGas: (*hexutil.Big)(big.NewInt(16)),
								Value:        *newRPCBalance(10000000000),
							}, {
								From:         &common.Address{0xc1},
								To:           &common.Address{0xc2},
								MaxFeePerGas: (*hexutil.Big)(big.NewInt(16)),
								Value:        *newRPCBalance(1000),
							}},
						},
					},
					Validation:             true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-more-params-validate",
			About: "simulates a simple do-nothing transaction with more fields set",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(3360000)},
							},
							BlockOverrides: &BlockOverrides{
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(0)),
							},
							Calls: []TransactionArgs{{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Gas:                  getUint64Ptr(0x52080),
								Value:                *newRPCBalance(0),
								MaxFeePerGas:         (*hexutil.Big)(big.NewInt(0)),
								MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(0)),
								MaxFeePerBlobGas:     (*hexutil.Big)(big.NewInt(0)),
								Nonce:                getUint64Ptr(0),
								Input:                hex2Bytes(""),
							}},
						},
					},
					Validation:             true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-empty-validation",
			About: "simulates empty with validation",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls:        []CallBatch{{}},
					Validation:             true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-empty",
			About: "simulates empty",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls:        []CallBatch{{}},
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-more-params-validate",
			About: "simulates a simple do-nothing transaction with more fields set",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(3360000)},
							},
							BlockOverrides: &BlockOverrides{
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(0)),
							},
							Calls: []TransactionArgs{{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Gas:                  getUint64Ptr(0x52080),
								Value:                *newRPCBalance(0),
								MaxFeePerGas:         (*hexutil.Big)(big.NewInt(0)),
								MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(0)),
								MaxFeePerBlobGas:     (*hexutil.Big)(big.NewInt(0)),
								Nonce:                getUint64Ptr(0),
								Input:                hex2Bytes(""),
							}},
						},
					},
					Validation:             true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-with-validation-no-funds",
			About: "simulates a ethSimulate transfer with validation and not enough funds",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(1000)},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							}, {
								From:  &common.Address{0xc1},
								To:    &common.Address{0xc2},
								Value: *newRPCBalance(1000),
							}},
						},
					},
					Validation: false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-no-funds",
			About: "simulates a simple ethSimulate transfer when account has no funds",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							}, {
								From:  &common.Address{0xc1},
								To:    &common.Address{0xc2},
								Value: *newRPCBalance(1000),
							}},
						},
					},
					Validation: false,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-overwrite-existing-contract",
			About: "overwrites existing contract with new contract",
			Run: func(ctx context.Context, t *T) error {
				contractAddr := common.HexToAddress("0000000000000000000000000000000000031ec7")
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &contractAddr,
								Input: hex2Bytes("a9059cbb0000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000a"), // transfer(address,uint256)
							}},
						},
						{
							StateOverrides: &StateOverride{
								contractAddr: OverrideAccount{Code: getBlockProperties()},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &contractAddr,
								Input: hex2Bytes("a9059cbb0000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000000a"), // transfer(address,uint256)
							}},
						},
					},
					Validation: false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},

		{
			Name:  "ethSimulate-overflow-nonce",
			About: "test to overflow nonce",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Nonce: getUint64Ptr(0xFFFFFFFFFFFFFFFF)},
							},
							Calls: []TransactionArgs{
								{
									From: &common.Address{0xc0},
									To:   &common.Address{0xc1},
								},
								{
									From: &common.Address{0xc0},
									To:   &common.Address{0xc1},
								},
							},
						},
					},
					Validation: false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-overflow-nonce-validation",
			About: "test to overflow nonce-validation",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Nonce: getUint64Ptr(0xFFFFFFFFFFFFFFFF)},
							},
							Calls: []TransactionArgs{
								{
									From: &common.Address{0xc0},
									To:   &common.Address{0xc1},
								},
								{
									From: &common.Address{0xc0},
									To:   &common.Address{0xc1},
								},
							},
						},
					},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-no-funds-with-balance-querying",
			About: "simulates a simple ethSimulate transfer when account has no funds with querying balances before and after",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								Code: getBalanceGetter(),
							},
						},
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c000000000000000000000000000000000000000"), // gets balance of c0
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c100000000000000000000000000000000000000"), // gets balance of c1
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c000000000000000000000000000000000000000"), // gets balance of c0
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c100000000000000000000000000000000000000"), // gets balance of c1
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c000000000000000000000000000000000000000"), // gets balance of c0
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c100000000000000000000000000000000000000"), // gets balance of c1
							},
						},
					}},
					Validation:             false,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-check-that-balance-is-there-after-new-block",
			About: "checks that balances are kept to next block",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(10000),
							},
							common.Address{0xc2}: OverrideAccount{
								Code: getBalanceGetter(),
							},
						},
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c000000000000000000000000000000000000000"), // gets balance of c0
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c100000000000000000000000000000000000000"), // gets balance of c1
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							},
						},
					}, {
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c000000000000000000000000000000000000000"), // gets balance of c0
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("f8b2cb4f000000000000000000000000c100000000000000000000000000000000000000"), // gets balance of c1
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-no-funds-with-validation",
			About: "simulates a simple ethSimulate transfer when account has no funds with validation",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
								Nonce: getUint64Ptr(0),
							}, {
								From:  &common.Address{0xc1},
								To:    &common.Address{0xc2},
								Value: *newRPCBalance(1000),
								Nonce: getUint64Ptr(1),
							}},
						},
					},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-no-funds-with-validation-without-nonces",
			About: "simulates a simple ethSimulate transfer when account has no funds with validation. This should fail as the nonce is not set for the second transaction.",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
								Nonce: getUint64Ptr(0),
							}, {
								From:  &common.Address{0xc1},
								To:    &common.Address{0xc2},
								Value: *newRPCBalance(1000),
							}},
						},
					},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-send-from-contract",
			About: "Sending eth from contract",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(1000), Code: getEthForwarder()},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-send-from-contract-no-balance",
			About: "Sending eth from contract without balance",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Code: getEthForwarder()},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-send-from-contract-with-validation",
			About: "Sending eth from contract with validation enabled",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(1000), Code: getEthForwarder()},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers:         true,
					Validation:             true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-transfer-over-BlockStateCalls",
			About: "simulates a transfering value over multiple BlockStateCalls",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(5000)},
						},
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(2000),
							}, {
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc3},
								Value: *newRPCBalance(2000),
							},
						},
					}, {
						StateOverrides: &StateOverride{
							{0xc3}: OverrideAccount{Balance: newRPCBalance(5000)},
						},
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc1},
								To:    &common.Address{0xc2},
								Value: *newRPCBalance(1000),
							}, {
								From:  &common.Address{0xc3},
								To:    &common.Address{0xc2},
								Value: *newRPCBalance(1000),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-block-num",
			About: "simulates calls overriding the block num",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 1)),
						},
						Calls: []TransactionArgs{
							{
								From: &common.Address{0xc0},
								Input: &hexutil.Bytes{
									0x43,             // NUMBER
									0x60, 0x00, 0x52, // MSTORE offset 0
									0x60, 0x20, 0x60, 0x00, 0xf3, // RETURN
								},
							},
						},
					}, {
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 2)),
						},
						Calls: []TransactionArgs{{
							From: &common.Address{0xc1},
							Input: &hexutil.Bytes{
								0x43,             // NUMBER
								0x60, 0x00, 0x52, // MSTORE offset 0
								0x60, 0x20, 0x60, 0x00, 0xf3,
							},
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-block-num-order-38020",
			About: "simulates calls with invalid block num order (-38020)",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 100)),
						},
						Calls: []TransactionArgs{{
							From: &common.Address{0xc1},
							Input: &hexutil.Bytes{
								0x43,             // NUMBER
								0x60, 0x00, 0x52, // MSTORE offset 0
								0x60, 0x20, 0x60, 0x00, 0xf3, // RETURN
							},
						}},
					}, {
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 90)),
						},
						Calls: []TransactionArgs{{
							From: &common.Address{0xc0},
							Input: &hexutil.Bytes{
								0x43,             // NUMBER
								0x60, 0x00, 0x52, // MSTORE offset 0
								0x60, 0x20, 0x60, 0x00, 0xf3, // RETURN
							},
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-block-timestamp-order-38021",
			About: "Error: simulates calls with invalid timestamp order (-38021)",
			Run: func(ctx context.Context, t *T) error {
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 12),
							},
						}, {
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 11),
							},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-block-timestamp-non-increment",
			About: "Error: simulates calls with timestamp staying the same",
			Run: func(ctx context.Context, t *T) error {
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 12),
							},
						}, {
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 12),
							},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-block-timestamps-incrementing",
			About: "checks that you can set timestamp and increment it in next block",
			Run: func(ctx context.Context, t *T) error {
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 11),
							},
						}, {
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 12),
							},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-block-timestamp-auto-increment",
			About: "Error: simulates calls with timestamp incrementing over another",
			Run: func(ctx context.Context, t *T) error {
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 11),
							},
						},
						{
							BlockOverrides: &BlockOverrides{},
						},
						{
							BlockOverrides: &BlockOverrides{
								Time: getUint64Ptr(latestBlockTime + 12),
							},
						},
						{
							BlockOverrides: &BlockOverrides{},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-set-read-storage",
			About: "simulates calls setting and reading from storage contract",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								Code: hex2Bytes("608060405234801561001057600080fd5b50600436106100365760003560e01c80632e64cec11461003b5780636057361d14610059575b600080fd5b610043610075565b60405161005091906100d9565b60405180910390f35b610073600480360381019061006e919061009d565b61007e565b005b60008054905090565b8060008190555050565b60008135905061009781610103565b92915050565b6000602082840312156100b3576100b26100fe565b5b60006100c184828501610088565b91505092915050565b6100d3816100f4565b82525050565b60006020820190506100ee60008301846100ca565b92915050565b6000819050919050565b600080fd5b61010c816100f4565b811461011757600080fd5b5056fea2646970667358221220404e37f487a89a932dca5e77faaf6ca2de3b991f93d230604b1b8daaef64766264736f6c63430008070033"),
							},
						},
						Calls: []TransactionArgs{{
							// Set value to 5
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("6057361d0000000000000000000000000000000000000000000000000000000000000005"),
						}, {
							// Read value
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("2e64cec1"),
						},
						},
					}},
				}
				res := make([]interface{}, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-logs",
			About: "simulates calls with logs",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								// Yul Code:
								// object "Test" {
								//    code {
								//        let hash:u256 := 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
								//        log1(0, 0, hash)
								//        return (0, 0)
								//    }
								// }
								Code: hex2Bytes("7fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff80600080a1600080f3"),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("6057361d0000000000000000000000000000000000000000000000000000000000000005"),
						}},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-blockhash-simple",
			About: "gets blockhash of block 1 (included in original chain)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								Code: blockHashCallerByteCode(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000001"),
						}},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls) != 1 {
					return fmt.Errorf("unexpected number of call results (have: %d, want: %d)", len(res[0].Calls), 1)
				}
				if err := checkBlockHash(common.BytesToHash(res[0].Calls[0].ReturnData), t.chain.GetBlock(1).Hash()); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-blockhash-complex",
			About: "gets blockhash of simulated block",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(2000000),
							},
							common.Address{0xc2}: OverrideAccount{
								Code: blockHashDeltaCallerByteCode(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000001"),
						}},
					}, {
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 30)),
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("ee82ac5e000000000000000000000000000000000000000000000000000000000000000f"),
						}},
					}, {
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 40)),
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("ee82ac5e000000000000000000000000000000000000000000000000000000000000001d"),
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-blockhash-start-before-head",
			About: "gets blockhash of simulated block",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(2000000),
							},
							common.Address{0xc2}: OverrideAccount{
								Code: blockHashDeltaCallerByteCode(),
							},
						},
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 5)),
						},
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000001"),
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000002"),
							},
						},
					}, {
						BlockOverrides: &BlockOverrides{
							Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 6)),
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000013"),
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-empty-with-block-num-set-firstblock",
			About: "set block number otherwise empty",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, (*hexutil.Big)(big.NewInt(1)))
				return nil
			},
		},
		{
			Name:  "ethSimulate-empty-with-block-num-set-minusone",
			About: "set block number otherwise empty with latest - 1",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, (*hexutil.Big)(big.NewInt(latestBlockNumber-1)))
				return nil
			},
		},
		{
			Name:  "ethSimulate-empty-with-block-num-set-current",
			About: "set block number otherwise empty with latest",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, (*hexutil.Big)(big.NewInt(latestBlockNumber)))
				return nil
			},
		},
		{
			Name:  "ethSimulate-empty-with-block-num-set-plus1",
			About: "set block number otherwise empty with latest + 1",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, (*hexutil.Big)(big.NewInt(latestBlockNumber+1)))
				return nil
			},
		},
		{
			Name:  "ethSimulate-self-destructing-state-override",
			About: "when selfdestructing a state override, the state override should go away",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								Code: selfDestructor(),
							},
							common.Address{0xc3}: OverrideAccount{
								Code: getCode(),
							},
						},
					}, {
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc3},
							Input: hex2Bytes("dce4a447000000000000000000000000c200000000000000000000000000000000000000"), //at(0xc2)
						}},
					}, {
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("83197ef0"), //destroy()
						}},
					}, {
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc3},
							Input: hex2Bytes("dce4a447000000000000000000000000c200000000000000000000000000000000000000"), //at(0xc2)
						}},
					}, {
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								Code: selfDestructor(),
							},
						},
					}, {
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc3},
							Input: hex2Bytes("dce4a447000000000000000000000000c200000000000000000000000000000000000000"), //at(0xc2)
						}},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-run-out-of-gas-in-block-38015",
			About: "we should get out of gas error if a block consumes too much gas (-38015)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(2000000),
							},
							common.Address{0xc2}: OverrideAccount{
								Code: gasSpender(),
							},
						},
						BlockOverrides: &BlockOverrides{
							GasLimit: getUint64Ptr(1500000),
						},
					}, {
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("815b8ab400000000000000000000000000000000000000000000000000000000000f4240"), //spendGas(1000000)
							},
							{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("815b8ab400000000000000000000000000000000000000000000000000000000000f4240"), //spendGas(1000000)
							},
						}},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-run-gas-spending",
			About: "spend a lot gas in separate blocks",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(2000000),
								},
								common.Address{0xc2}: OverrideAccount{
									Code: gasSpender(),
								},
							},
							BlockOverrides: &BlockOverrides{
								GasLimit: getUint64Ptr(1500000),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("815b8ab40000000000000000000000000000000000000000000000000000000000000000"), //spendGas(0)
								},
							},
						},
						{
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("815b8ab40000000000000000000000000000000000000000000000000000000000000000"), //spendGas(0)
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("815b8ab400000000000000000000000000000000000000000000000000000000000f4240"), //spendGas(1000000)
								},
							},
						},
						{
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("815b8ab40000000000000000000000000000000000000000000000000000000000000000"), //spendGas(0)
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("815b8ab400000000000000000000000000000000000000000000000000000000000f4240"), //spendGas(1000000)
								},
							},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-eth-send-should-produce-logs",
			About: "when sending eth we should get ETH logs when traceTransfers is set",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls[0].Logs) != 1 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 1)
				}
				if res[0].Calls[0].Logs[0].Address.String() != "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE" {
					return fmt.Errorf("unexpected log address (have: %s, want: %s)", res[0].Calls[0].Logs[0].Address.String(), "0xEeeeeEeeeEeEeeEeEeEeeEEEeeeeEeeeeeeeEEeE")
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-address-twice",
			About: "override address twice",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
							common.Address{0xc0}: OverrideAccount{Code: getRevertingContract()},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-empty-calls-and-overrides-ethSimulate",
			About: "ethSimulate with state overrides and calls but they are empty",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{},
							Calls:          []TransactionArgs{{}},
						},
						{
							StateOverrides: &StateOverride{},
							Calls:          []TransactionArgs{{}},
						},
					},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-address-twice-in-separate-BlockStateCalls",
			About: "override address twice in separate BlockStateCalls",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							}},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Value: *newRPCBalance(1000),
							}},
						},
					},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-eth-send-should-not-produce-logs-on-revert",
			About: "we should not be producing eth logs if the transaction reverts and ETH is not sent",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
							common.Address{0xc1}: OverrideAccount{Code: getRevertingContract()},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers:         true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls[0].Logs) != 0 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 0)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-eth-send-should-produce-more-logs-on-forward",
			About: "we should be getting more logs if eth is forwarded",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
							common.Address{0xc1}: OverrideAccount{Code: getEthForwarder()},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
							Input: hex2Bytes("4b64e4920000000000000000000000000000000000000000000000000000000000000100"),
						}},
					}},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls[0].Logs) != 2 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 2)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-eth-send-should-produce-no-logs-on-forward-revert",
			About: "we should be getting no logs if eth is forwarded but then the tx reverts",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
							common.Address{0xc1}: OverrideAccount{Code: getEthForwarder()},
							common.Address{0xc2}: OverrideAccount{Code: getRevertingContract()},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
							Input: hex2Bytes("4b64e492c200000000000000000000000000000000000000000000000000000000000000"), //foward(0xc2)
						}},
					}},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls[0].Logs) != 0 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 0)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-eth-send-should-not-produce-logs-by-default",
			About: "when sending eth we should not get ETH logs by default",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls[0].Logs) != 0 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 0)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-transaction-too-low-nonce-38010",
			About: "Error: Nonce too low (-38010)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Nonce: getUint64Ptr(10)},
						},
						Calls: []TransactionArgs{{
							Nonce: getUint64Ptr(0),
							From:  &common.Address{0xc1},
							To:    &common.Address{0xc1},
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-transaction-too-high-nonce",
			About: "Error: Nonce too high",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{
							Nonce: getUint64Ptr(100),
							From:  &common.Address{0xc1},
							To:    &common.Address{0xc1},
						}},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-basefee-too-low-with-validation-38012",
			About: "Error: BaseFeePerGas too low with validation (-38012)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
						},
						BlockOverrides: &BlockOverrides{
							BaseFeePerGas: (*hexutil.Big)(big.NewInt(10)),
						},
						Calls: []TransactionArgs{{
							From:                 &common.Address{0xc0},
							To:                   &common.Address{0xc1},
							MaxFeePerGas:         (*hexutil.Big)(big.NewInt(0)),
							MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(0)),
						}},
					}},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-basefee-too-low-without-validation-38012",
			About: "Error: BaseFeePerGas too low with no validation (-38012)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(2000)},
						},
						BlockOverrides: &BlockOverrides{
							BaseFeePerGas: (*hexutil.Big)(big.NewInt(10)),
						},
						Calls: []TransactionArgs{{
							From:                 &common.Address{0xc1},
							To:                   &common.Address{0xc1},
							MaxFeePerGas:         (*hexutil.Big)(big.NewInt(0)),
							MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(0)),
						}},
					}},
					Validation: false,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-basefee-too-low-without-validation-38012-without-basefee-override",
			About: "tries to send transaction with zero basefee",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{
							From:                 &common.Address{0xc1},
							To:                   &common.Address{0xc1},
							MaxFeePerGas:         (*hexutil.Big)(big.NewInt(0)),
							MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(0)),
						}},
					}},
					Validation: false,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-instrict-gas-38013",
			About: "Error: Not enough gas provided to pay for intrinsic gas (-38013)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{
							From: &common.Address{0xc1},
							To:   &common.Address{0xc1},
							Gas:  getUint64Ptr(0),
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-gas-fees-and-value-error-38014",
			About: "Error: Insufficient funds to pay for gas fees and value (-38014)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-gas-fees-and-value-error-38014-with-validation",
			About: "Error: Insufficient funds to pay for gas fees and value (-38014) with validation",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1000),
						}},
					}},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-move-to-address-itself-reference-38022",
			About: "Error: MovePrecompileToAddress referenced itself in replacement (-38022)",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(200000)},
							common.Address{0xc1}: OverrideAccount{MovePrecompileToAddress: &common.Address{0xc1}},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Value: *newRPCBalance(1),
						}},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-move-two-non-precompiles-accounts-to-same",
			About: "Move two non-precompiles to same adddress",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0x1}: OverrideAccount{
								MovePrecompileToAddress: &common.Address{0xc2},
							},
							common.Address{0x2}: OverrideAccount{
								MovePrecompileToAddress: &common.Address{0xc2},
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-move-two-accounts-to-same-38023",
			About: "Move two accounts to the same destination (-38023)",
			Run: func(ctx context.Context, t *T) error {
				ecRecoverAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000001"))
				keccakAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000002"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							ecRecoverAddress: OverrideAccount{
								MovePrecompileToAddress: &common.Address{0xc2},
							},
							keccakAddress: OverrideAccount{
								MovePrecompileToAddress: &common.Address{0xc2},
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-try-to-move-non-precompile",
			About: "try to move non-precompile",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Nonce: getUint64Ptr(5)},
							},
						}, {
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{MovePrecompileToAddress: &common.Address{0xc1}},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc0},
									Nonce: getUint64Ptr(0),
								},
								{
									From:  &common.Address{0xc1},
									To:    &common.Address{0xc1},
									Nonce: getUint64Ptr(5),
								},
							},
						},
					},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-make-call-with-future-block",
			About: "start ethSimulate with future block",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							Calls: []TransactionArgs{{
								From: &common.Address{0xc0},
								To:   &common.Address{0xc0},
							}},
						},
					},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "0x111")
				return nil
			},
		},
		{
			Name:  "ethSimulate-check-that-nonce-increases",
			About: "check that nonce increases",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(20000000000)},
							},
							BlockOverrides: &BlockOverrides{
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(9)),
							},
							Calls: []TransactionArgs{
								{
									From:         &common.Address{0xc0},
									To:           &common.Address{0xc0},
									Nonce:        getUint64Ptr(0),
									MaxFeePerGas: (*hexutil.Big)(big.NewInt(15)),
								},
								{
									From:         &common.Address{0xc0},
									To:           &common.Address{0xc0},
									Nonce:        getUint64Ptr(1),
									MaxFeePerGas: (*hexutil.Big)(big.NewInt(15)),
								},
								{
									From:         &common.Address{0xc0},
									To:           &common.Address{0xc0},
									Nonce:        getUint64Ptr(2),
									MaxFeePerGas: (*hexutil.Big)(big.NewInt(15)),
								},
							},
						},
					},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-check-invalid-nonce",
			About: "check that nonce cannot decrease",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(1)),
							},
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(20000)},
							},
						}, {
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc0},
									Nonce: getUint64Ptr(0),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc0},
									Nonce: getUint64Ptr(1),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc0},
									Nonce: getUint64Ptr(0),
								},
							},
						},
					},
					Validation: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-all-in-BlockStateCalls",
			About: "override all values in block and see that they are set in return value",
			Run: func(ctx context.Context, t *T) error {
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				latestBlockNumber := t.chain.Head().Number().Int64()
				feeRecipient := common.Address{0xc2}
				randDao := common.Hash{0xc3}
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						BlockOverrides: &BlockOverrides{
							Number:        (*hexutil.Big)(big.NewInt(latestBlockNumber + 10)),
							Time:          getUint64Ptr(latestBlockTime + 10),
							GasLimit:      getUint64Ptr(1004),
							FeeRecipient:  &feeRecipient,
							PrevRandao:    &randDao,
							BaseFeePerGas: (*hexutil.Big)(big.NewInt(1007)),
						},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-move-ecrecover-and-call",
			About: "move ecrecover and try calling it",
			Run: func(ctx context.Context, t *T) error {
				ecRecoverAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000001"))
				ecRecoverMovedToAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123456"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{ // just call ecrecover normally
							{ // call with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
						},
					}, {
						StateOverrides: &StateOverride{ // move ecRecover and call it in new address
							ecRecoverAddress: OverrideAccount{
								MovePrecompileToAddress: &ecRecoverMovedToAddress,
							},
						},
						Calls: []TransactionArgs{
							{ // call with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
							{ // call with valid params, the old address, should fail as it was moved
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls) != len(params.BlockStateCalls[0].Calls) {
					return fmt.Errorf("unexpected number of call results (have: %d, want: %d)", len(res[0].Calls), len(params.BlockStateCalls[0].Calls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-move-ecrecover-twice-and-call",
			About: "move ecrecover and try calling it, then move it again and call it",
			Run: func(ctx context.Context, t *T) error {
				ecRecoverAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000001"))
				ecRecoverMovedToAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123456"))
				ecRecoverMovedToAddress2 := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123457"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{ // just call ecrecover normally
							{ // call with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
						},
					}, {
						StateOverrides: &StateOverride{ // move ecRecover and call it in new address
							ecRecoverAddress: OverrideAccount{
								MovePrecompileToAddress: &ecRecoverMovedToAddress,
							},
						},
						Calls: []TransactionArgs{
							{ // call with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
							{ // call with valid params, the old address, should fail as it was moved
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
						},
					}, {
						StateOverrides: &StateOverride{ // move ecRecover and call it in new address
							ecRecoverAddress: OverrideAccount{
								MovePrecompileToAddress: &ecRecoverMovedToAddress2,
							},
						},
						Calls: []TransactionArgs{
							{ // call with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
							{ // call with valid params, the old address, should fail as it was moved
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress2,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-ecrecover",
			About: "override ecrecover",
			Run: func(ctx context.Context, t *T) error {
				ecRecoverAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000001"))
				ecRecoverMovedToAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123456"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							ecRecoverAddress: OverrideAccount{
								Code:                    getEcRecoverOverride(),
								MovePrecompileToAddress: &ecRecoverMovedToAddress,
							},
							common.Address{0xc1}: OverrideAccount{Balance: newRPCBalance(200000)},
						},
						Calls: []TransactionArgs{
							{ // call with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
							{ // add override
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("c00692604554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000000000000000000000000000d8da6bf26964af9d7eed9e03e53415d37aa96045"),
							},
							{ // now it should resolve to 0xd8dA6BF26964aF9D7eEd9e03E53415D37aA96045
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
							{ // call with new invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554490000000000000000000000000000000000000000000000000000000000"),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls) != len(params.BlockStateCalls[0].Calls) {
					return fmt.Errorf("unexpected number of call results (have: %d, want: %d)", len(res[0].Calls), len(params.BlockStateCalls[0].Calls))
				}
				zeroAddr := common.Address{0x0}
				if common.BytesToAddress(res[0].Calls[0].ReturnData) != zeroAddr {
					return fmt.Errorf("unexpected ReturnData (have: %d, want: %d)", common.BytesToAddress(res[0].Calls[0].ReturnData), zeroAddr)
				}
				successReturn := common.BytesToAddress(*hex2Bytes("b11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a"))
				if common.BytesToAddress(res[0].Calls[1].ReturnData) != successReturn {
					return fmt.Errorf("unexpected calls 1 ReturnData (have: %d, want: %d)", common.BytesToAddress(res[0].Calls[1].ReturnData), successReturn)
				}
				vitalikReturn := common.BytesToAddress(*hex2Bytes("d8dA6BF26964aF9D7eEd9e03E53415D37aA96045"))
				if common.BytesToAddress(res[0].Calls[3].ReturnData) != vitalikReturn {
					return fmt.Errorf("unexpected calls 3 ReturnData (have: %d, want: %d)", common.BytesToAddress(res[0].Calls[3].ReturnData), vitalikReturn)
				}
				if common.BytesToAddress(res[0].Calls[4].ReturnData) != successReturn {
					return fmt.Errorf("unexpected calls 4 ReturnData (have: %d, want: %d)", common.BytesToAddress(res[0].Calls[4].ReturnData), successReturn)
				}
				if common.BytesToAddress(res[0].Calls[5].ReturnData) != zeroAddr {
					return fmt.Errorf("unexpected calls 5 ReturnData (have: %d, want: %d)", common.BytesToAddress(res[0].Calls[5].ReturnData), zeroAddr)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-sha256",
			About: "override sha256 precompile",
			Run: func(ctx context.Context, t *T) error {
				sha256Address := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000002"))
				sha256MovedToAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123456"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							sha256Address: OverrideAccount{
								Code:                    hex2Bytes(""),
								MovePrecompileToAddress: &sha256MovedToAddress,
							},
						},
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &sha256MovedToAddress,
								Input: hex2Bytes("1234"),
							},
							{
								From:  &common.Address{0xc0},
								To:    &sha256Address,
								Input: hex2Bytes("1234"),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-two-blocks-with-complete-eth-sends",
			About: "two blocks with eth sends",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						BlockOverrides: &BlockOverrides{
							BaseFeePerGas: (*hexutil.Big)(big.NewInt(10)),
						},
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(900001000)},
							common.Address{0xc1}: OverrideAccount{Balance: newRPCBalance(900002000)},
						},
						Calls: []TransactionArgs{
							{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Input:                hex2Bytes(""),
								Gas:                  getUint64Ptr(21000),
								MaxFeePerGas:         *newRPCBalance(20),
								MaxPriorityFeePerGas: *newRPCBalance(1),
								MaxFeePerBlobGas:     *newRPCBalance(0),
								Value:                *newRPCBalance(101),
								Nonce:                getUint64Ptr(0),
							},
							{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Input:                hex2Bytes(""),
								Gas:                  getUint64Ptr(21000),
								MaxFeePerGas:         *newRPCBalance(20),
								MaxPriorityFeePerGas: *newRPCBalance(1),
								MaxFeePerBlobGas:     *newRPCBalance(0),
								Value:                *newRPCBalance(102),
								Nonce:                getUint64Ptr(1),
							},
							{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Input:                hex2Bytes(""),
								Gas:                  getUint64Ptr(21000),
								MaxFeePerGas:         *newRPCBalance(20),
								MaxPriorityFeePerGas: *newRPCBalance(1),
								MaxFeePerBlobGas:     *newRPCBalance(0),
								Value:                *newRPCBalance(103),
								Nonce:                getUint64Ptr(2),
							},
							{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Input:                hex2Bytes(""),
								Gas:                  getUint64Ptr(21000),
								MaxFeePerGas:         *newRPCBalance(20),
								MaxPriorityFeePerGas: *newRPCBalance(1),
								MaxFeePerBlobGas:     *newRPCBalance(0),
								Value:                *newRPCBalance(104),
								Nonce:                getUint64Ptr(3),
							},
							{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Input:                hex2Bytes(""),
								Gas:                  getUint64Ptr(21000),
								MaxFeePerGas:         *newRPCBalance(20),
								MaxPriorityFeePerGas: *newRPCBalance(1),
								MaxFeePerBlobGas:     *newRPCBalance(0),
								Value:                *newRPCBalance(105),
								Nonce:                getUint64Ptr(4),
							},
							{
								From:                 &common.Address{0xc0},
								To:                   &common.Address{0xc1},
								Input:                hex2Bytes(""),
								Gas:                  getUint64Ptr(21000),
								MaxFeePerGas:         *newRPCBalance(20),
								MaxPriorityFeePerGas: *newRPCBalance(1),
								MaxFeePerBlobGas:     *newRPCBalance(0),
								Value:                *newRPCBalance(106),
								Nonce:                getUint64Ptr(5),
							},
							{
								From:                 &common.Address{0xc1},
								To:                   &common.Address{0xc2},
								Input:                hex2Bytes(""),
								Gas:                  getUint64Ptr(21000),
								MaxFeePerGas:         *newRPCBalance(20),
								MaxPriorityFeePerGas: *newRPCBalance(1),
								MaxFeePerBlobGas:     *newRPCBalance(0),
								Value:                *newRPCBalance(106),
								Nonce:                getUint64Ptr(0),
							},
						},
					},
						{
							Calls: []TransactionArgs{
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(20),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(101),
									Nonce:                getUint64Ptr(6),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(20),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(102),
									Nonce:                getUint64Ptr(7),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(20),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(103),
									Nonce:                getUint64Ptr(8),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(20),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(104),
									Nonce:                getUint64Ptr(9),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(20),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(105),
									Nonce:                getUint64Ptr(10),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(20),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(106),
									Nonce:                getUint64Ptr(11),
								},
								{
									From:                 &common.Address{0xc1},
									To:                   &common.Address{0xc2},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(20),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(106),
									Nonce:                getUint64Ptr(1),
								},
							},
						}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-identity",
			About: "override identity precompile",
			Run: func(ctx context.Context, t *T) error {
				identityAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000004"))
				identityMovedToAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123456"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							identityAddress: OverrideAccount{
								Code:                    hex2Bytes(""),
								MovePrecompileToAddress: &identityMovedToAddress,
							},
						},
						Calls: []TransactionArgs{
							{
								From:  &common.Address{0xc0},
								To:    &identityMovedToAddress,
								Input: hex2Bytes("1234"),
							},
							{
								From:  &common.Address{0xc0},
								To:    &identityAddress,
								Input: hex2Bytes("1234"),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-precompile-is-sending-transaction",
			About: "send transaction from a precompile",
			Run: func(ctx context.Context, t *T) error {
				identityAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000004"))
				sha256Address := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000002"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{
							{
								From:  &identityAddress,
								To:    &sha256Address,
								Input: hex2Bytes("1234"),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-state-diff",
			About: "override one state variable with statediff",
			Run: func(ctx context.Context, t *T) error {
				stateChanges := make(map[common.Hash]common.Hash)
				stateChanges[common.BytesToHash(*hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"))] = common.Hash{0x12} //slot 0 -> 0x12
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(2000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: getStorageTester(),
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"), // set storage slot 0 -> 1
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"), // set storage slot 1 -> 2
								},
							},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									StateDiff: &stateChanges, // state diff override
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000000"), // gets storage slot 0, should be 0x12 as overrided
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000001"), // gets storage slot 1, should be 2
								},
							},
						},
					},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-simple-state-diff",
			About: "override one state variable with state",
			Run: func(ctx context.Context, t *T) error {
				stateChanges := make(map[common.Hash]common.Hash)
				stateChanges[common.BytesToHash(*hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"))] = common.Hash{0x12} //slot 0 -> 0x12
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(2000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: getStorageTester(),
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"), // set storage slot 0 -> 1
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"), // set storage slot 1 -> 2
								},
							},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									State: &stateChanges, // state diff override
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000000"), // gets storage slot 0, should be 0x12 as overrided
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000001"), // gets storage slot 1, should be 0
								},
							},
						},
					},
					TraceTransfers:         true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-override-storage-slots",
			About: "override storage slots",
			Run: func(ctx context.Context, t *T) error {
				stateChanges := make(map[common.Hash]common.Hash)
				stateChanges[common.BytesToHash(*hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"))] = common.Hash{0x12} //slot 0 -> 0x12
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(2000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: getStorageTester(),
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"), // set storage slot 0 -> 1
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"), // set storage slot 1 -> 2
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000000"), // gets storage slot 0, should be 1
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000001"), // gets storage slot 1, should be 2
								},
							},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									StateDiff: &stateChanges, // state diff override
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000000"), // gets storage slot 0, should be 0x12 as overrided
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000001"), // gets storage slot 1, should be 2
								},
							},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									State: &stateChanges, // whole state override
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000000"), // gets storage slot 0, should be 0x12 as overrided
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("0ff4c9160000000000000000000000000000000000000000000000000000000000000001"), // gets storage slot 1, should be 0 as the whole storage was replaced
								},
							},
						},
					},
					TraceTransfers:         true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if res[0].Calls[2].ReturnData.String() != "0x0000000000000000000000000000000000000000000000000000000000000001" {
					return fmt.Errorf("unexpected call result (res[0].Calls[2]) (have: %s, want: %s)", res[0].Calls[2].ReturnData.String(), "0x0000000000000000000000000000000000000000000000000000000000000001")
				}
				if res[0].Calls[3].ReturnData.String() != "0x0000000000000000000000000000000000000000000000000000000000000002" {
					return fmt.Errorf("unexpected call result (res[0].Calls[3]) (have: %s, want: %s)", res[0].Calls[3].ReturnData.String(), "0x0000000000000000000000000000000000000000000000000000000000000002")
				}

				if res[1].Calls[0].ReturnData.String() != "0x1200000000000000000000000000000000000000000000000000000000000000" {
					return fmt.Errorf("unexpected call result (res[1].Calls[0]) (have: %s, want: %s)", res[1].Calls[0].ReturnData.String(), "0x1200000000000000000000000000000000000000000000000000000000000000")
				}
				if res[1].Calls[1].ReturnData.String() != "0x0000000000000000000000000000000000000000000000000000000000000002" {
					return fmt.Errorf("unexpected call result (res[1].Calls[1]) (have: %s, want: %s)", res[1].Calls[1].ReturnData.String(), "0x0000000000000000000000000000000000000000000000000000000000000002")
				}

				if res[2].Calls[0].ReturnData.String() != "0x1200000000000000000000000000000000000000000000000000000000000000" {
					return fmt.Errorf("unexpected call result (res[2].Calls[0]) (have: %s, want: %s)", res[2].Calls[0].ReturnData.String(), "0x1200000000000000000000000000000000000000000000000000000000000000")
				}
				if res[2].Calls[1].ReturnData.String() != "0x0000000000000000000000000000000000000000000000000000000000000000" {
					return fmt.Errorf("unexpected call result (res[2].Calls[1]) (have: %s, want: %s)", res[2].Calls[1].ReturnData.String(), "0x0000000000000000000000000000000000000000000000000000000000000000")
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-block-override-reflected-in-contract-simple",
			About: "Checks that block overrides are true in contract for block number and time",
			Run: func(ctx context.Context, t *T) error {
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 5)),
								Time:   getUint64Ptr(latestBlockTime + 10),
							},
						},
						{
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 10)),
								Time:   getUint64Ptr(latestBlockTime + 20),
							},
						},
						{
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 20)),
								Time:   getUint64Ptr(latestBlockTime + 30),
							},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-get-block-properties",
			About: "gets various block properties from chain",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
					},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-block-override-reflected-in-contract",
			About: "Checks that block overrides are true in contract",
			Run: func(ctx context.Context, t *T) error {
				prevRandDao1 := common.BytesToHash(*hex2Bytes("123"))
				prevRandDao2 := common.BytesToHash(*hex2Bytes("1234"))
				prevRandDao3 := common.BytesToHash(*hex2Bytes("12345"))
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							BlockOverrides: &BlockOverrides{
								Number:        (*hexutil.Big)(big.NewInt(latestBlockNumber + 5)),
								Time:          getUint64Ptr(latestBlockTime + 10),
								GasLimit:      getUint64Ptr(190000),
								FeeRecipient:  &common.Address{0xc0},
								PrevRandao:    &prevRandDao1,
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(10)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
						{
							BlockOverrides: &BlockOverrides{
								Number:        (*hexutil.Big)(big.NewInt(latestBlockNumber + 10)),
								Time:          getUint64Ptr(latestBlockTime + 10),
								GasLimit:      getUint64Ptr(300000),
								FeeRecipient:  &common.Address{0xc1},
								PrevRandao:    &prevRandDao2,
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(20)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
						{
							BlockOverrides: &BlockOverrides{
								Number:        (*hexutil.Big)(big.NewInt(latestBlockNumber + 15)),
								Time:          getUint64Ptr(latestBlockTime + 20),
								GasLimit:      getUint64Ptr(190002),
								FeeRecipient:  &common.Address{0xc2},
								PrevRandao:    &prevRandDao3,
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(30)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-add-more-non-defined-BlockStateCalls-than-fit",
			About: "Add more BlockStateCalls between two BlockStateCalls than it actually fits there",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 100)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
						{
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
						{
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 101)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
					},
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-add-more-non-defined-BlockStateCalls-than-fit-but-now-with-fit",
			About: "Not all block numbers are defined",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc1}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 105)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
						{
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
						{
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 120)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
						{
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes(""),
								},
							},
						},
					},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-fee-recipient-receiving-funds",
			About: "Check that fee recipient gets funds",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(500000000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: getBalanceGetter(),
								},
							},
							BlockOverrides: &BlockOverrides{
								Number:        (*hexutil.Big)(big.NewInt(latestBlockNumber + 42)),
								FeeRecipient:  &common.Address{0xc2},
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(10)),
							},
							Calls: []TransactionArgs{
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									MaxFeePerGas:         (*hexutil.Big)(big.NewInt(10)),
									MaxPriorityFeePerGas: (*hexutil.Big)(big.NewInt(10)),
									Input:                hex2Bytes(""),
									Nonce:                getUint64Ptr(0),
								},
								{
									From:         &common.Address{0xc0},
									To:           &common.Address{0xc1},
									MaxFeePerGas: (*hexutil.Big)(big.NewInt(10)),
									Input:        hex2Bytes("f8b2cb4f000000000000000000000000c000000000000000000000000000000000000000"), // gets balance of c0
									Nonce:        getUint64Ptr(1),
								},
								{
									From:         &common.Address{0xc0},
									To:           &common.Address{0xc1},
									MaxFeePerGas: (*hexutil.Big)(big.NewInt(10)),
									Input:        hex2Bytes("f8b2cb4f000000000000000000000000c200000000000000000000000000000000000000"), // gets balance of c2
									Nonce:        getUint64Ptr(2),
								},
							},
						},
					},
					Validation:             true,
					TraceTransfers:         true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-contract-calls-itself",
			About: "contract calls itself",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							Calls: []TransactionArgs{
								{
									From: &common.Address{0xc0},
									To:   &common.Address{0xc0},
								},
							},
						},
					},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-send-eth-and-delegate-call",
			About: "sending eth and delegate calling should only produce one log",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(2000000),
							},
							common.Address{0xc1}: OverrideAccount{
								Code: delegateCaller(),
							},
							common.Address{0xc2}: OverrideAccount{
								Code: getBlockProperties(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes("5c19a95c000000000000000000000000c200000000000000000000000000000000000000"),
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if res[0].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if len(res[0].Calls[0].Logs) != 1 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 1)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-send-eth-and-delegate-call-to-payble-contract",
			About: "sending eth and delegate calling a payable contract should only produce one log",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(2000000),
							},
							common.Address{0xc1}: OverrideAccount{
								Code: delegateCaller2(),
							},
							common.Address{0xc2}: OverrideAccount{
								Code: payableFallBack(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes(""),
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if res[0].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if len(res[0].Calls[0].Logs) != 1 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 1)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-send-eth-and-delegate-call-to-eoa",
			About: "sending eth and delegate calling a eoa should only produce one log",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(2000000),
							},
							common.Address{0xc1}: OverrideAccount{
								Code: delegateCaller2(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes(""),
							Value: *newRPCBalance(1000),
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if res[0].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if len(res[0].Calls[0].Logs) != 1 {
					return fmt.Errorf("unexpected number of logs (have: %d, want: %d)", len(res[0].Calls[0].Logs), 1)
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-extcodehash-override",
			About: "test extcodehash getting of overriden contract",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc0}: OverrideAccount{
								Balance: newRPCBalance(2000000),
							},
							common.Address{0xc1}: OverrideAccount{
								Code: extCodeHashContract(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes("b9724d63000000000000000000000000c200000000000000000000000000000000000000"), // getExtCodeHash(0xc2)
						}},
					}, {
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								Code: getBlockProperties(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes("b9724d63000000000000000000000000c200000000000000000000000000000000000000"), // getExtCodeHash(0xc2)
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if res[0].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if res[1].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if res[0].Calls[0].ReturnData.String() == res[1].Calls[0].ReturnData.String() {
					return fmt.Errorf("returndata did not change (have: %s, want: %s)", res[0].Calls[0].ReturnData.String(), res[1].Calls[0].ReturnData.String())
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-extcodehash-existing-contract",
			About: "test extcodehash getting of existing contract and then overriding it",
			Run: func(ctx context.Context, t *T) error {
				contractAddr := common.HexToAddress("0000000000000000000000000000000000031ec7")
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc1}: OverrideAccount{
								Code: extCodeHashContract(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes("b9724d630000000000000000000000000000000000000000000000000000000000031ec7"), // getExtCodeHash(0000000000000000000000000000000000031ec7)
						}},
					}, {
						StateOverrides: &StateOverride{
							contractAddr: OverrideAccount{
								Code: getBlockProperties(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes("b9724d630000000000000000000000000000000000000000000000000000000000031ec7"), // getExtCodeHash(0000000000000000000000000000000000031ec7)
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if res[0].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if res[1].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if res[0].Calls[0].ReturnData.String() == res[1].Calls[0].ReturnData.String() {
					return fmt.Errorf("returndata did not change (have: %s, want: %s)", res[0].Calls[0].ReturnData.String(), res[1].Calls[0].ReturnData.String())
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-extcodehash-precompile",
			About: "test extcodehash getting of precompile and then again after override",
			Run: func(ctx context.Context, t *T) error {
				ecRecoverAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000001"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc1}: OverrideAccount{
								Code: extCodeHashContract(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes("b9724d630000000000000000000000000000000000000000000000000000000000000001"), // getExtCodeHash(0x1)
						}},
					}, {
						StateOverrides: &StateOverride{
							ecRecoverAddress: OverrideAccount{
								Code: getBlockProperties(),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc1},
							Input: hex2Bytes("b9724d630000000000000000000000000000000000000000000000000000000000000001"), // getExtCodeHash(0x1)
						}},
					}},
					TraceTransfers: true,
					Validation:     false,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if res[0].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if res[1].Calls[0].Status != 1 {
					return fmt.Errorf("unexpected call status (have: %d, want: %d)", res[0].Calls[0].Status, 1)
				}
				if res[0].Calls[0].ReturnData.String() == res[1].Calls[0].ReturnData.String() {
					return fmt.Errorf("returndata did not change (have: %s, want: %s)", res[0].Calls[0].ReturnData.String(), res[1].Calls[0].ReturnData.String())
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-self-destructive-contract-produces-logs",
			About: "self destructive contract produces logs",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{
							common.Address{0xc2}: OverrideAccount{
								Code:    selfDestructor(),
								Balance: newRPCBalance(2000000),
							},
						},
						Calls: []TransactionArgs{{
							From:  &common.Address{0xc0},
							To:    &common.Address{0xc2},
							Input: hex2Bytes("83197ef0"), //destroy()
						}},
					}},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-no-fields-call",
			About: "make a call with no fields",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{}},
					}},
					TraceTransfers:         true,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-only-from-transaction",
			About: "make a call with only from field",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{
							From: &common.Address{0xc0},
						}},
					}},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-only-from-to-transaction",
			About: "make a call with only from and to fields",
			Run: func(ctx context.Context, t *T) error {
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						Calls: []TransactionArgs{{
							From: &common.Address{0xc0},
							To:   &common.Address{0xc1},
						}},
					}},
					TraceTransfers: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-big-block-state-calls-array",
			About: "Have a block state calls with 300 blocks",
			Run: func(ctx context.Context, t *T) error {
				calls := make([]CallBatch, 300)
				params := ethSimulateOpts{BlockStateCalls: calls}
				res := make([]blockResult, 0)
				t.rpc.Call(&res, "eth_simulateV1", params, "latest")
				return nil
			},
		},
		{
			Name:  "ethSimulate-move-ecrecover-and-call-old-and-new",
			About: "move ecrecover and try calling the moved and non-moved version",
			Run: func(ctx context.Context, t *T) error {
				ecRecoverAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000001"))
				ecRecoverMovedToAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123456"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{{
						StateOverrides: &StateOverride{ // move ecRecover and call it in new address
							ecRecoverAddress: OverrideAccount{
								MovePrecompileToAddress: &ecRecoverMovedToAddress,
							},
						},
						Calls: []TransactionArgs{
							{ // call new address with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call new address with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverMovedToAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
							{ // call old address with invalid params, should fail (resolve to 0x0)
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
							},
							{ // call old address with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
								From:  &common.Address{0xc1},
								To:    &ecRecoverAddress,
								Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
							},
						},
					}},
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				if len(res) != len(params.BlockStateCalls) {
					return fmt.Errorf("unexpected number of results (have: %d, want: %d)", len(res), len(params.BlockStateCalls))
				}
				if len(res[0].Calls) != len(params.BlockStateCalls[0].Calls) {
					return fmt.Errorf("unexpected number of call results (have: %d, want: %d)", len(res[0].Calls), len(params.BlockStateCalls[0].Calls))
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-use-as-many-features-as-possible",
			About: "try using all eth simulates features at once",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				prevRandDao := common.BytesToHash(*hex2Bytes("12345"))
				stateChanges := make(map[common.Hash]common.Hash)
				stateChanges[common.BytesToHash(*hex2Bytes("0000000000000000000000000000000000000000000000000000000000000000"))] = common.Hash{0x12} //slot 0 -> 0x12
				ecRecoverAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000000001"))
				ecRecoverMovedToAddress := common.BytesToAddress(*hex2Bytes("0000000000000000000000000000000000123456"))
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								Number:        (*hexutil.Big)(big.NewInt(latestBlockNumber + 10)),
								Time:          getUint64Ptr(latestBlockTime + 500),
								FeeRecipient:  &common.Address{0xc2},
								PrevRandao:    &prevRandDao,
								BaseFeePerGas: (*hexutil.Big)(big.NewInt(1007)),
							},
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{Balance: newRPCBalance(900001000000)},
								common.Address{0xc1}: OverrideAccount{
									Code:    extCodeHashContract(),
									Balance: newRPCBalance(900002000000),
								},
							},
							Calls: []TransactionArgs{
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(101),
									Nonce:                getUint64Ptr(0),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(102),
									Nonce:                getUint64Ptr(1),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(103),
									Nonce:                getUint64Ptr(2),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(104),
									Nonce:                getUint64Ptr(3),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(105),
									Nonce:                getUint64Ptr(4),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(106),
									Nonce:                getUint64Ptr(5),
								},
								{
									From:                 &common.Address{0xc1},
									To:                   &common.Address{0xc2},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(106),
									Nonce:                getUint64Ptr(0),
								},
							},
						},
						{
							Calls: []TransactionArgs{
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(101),
									Nonce:                getUint64Ptr(6),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(102),
									Nonce:                getUint64Ptr(7),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(103),
									Nonce:                getUint64Ptr(8),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(104),
									Nonce:                getUint64Ptr(9),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(105),
									Nonce:                getUint64Ptr(10),
								},
								{
									From:                 &common.Address{0xc0},
									To:                   &common.Address{0xc1},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(106),
									Nonce:                getUint64Ptr(11),
								},
								{
									From:                 &common.Address{0xc1},
									To:                   &common.Address{0xc2},
									Input:                hex2Bytes(""),
									Gas:                  getUint64Ptr(21000),
									MaxFeePerGas:         *newRPCBalance(2000),
									MaxPriorityFeePerGas: *newRPCBalance(1),
									MaxFeePerBlobGas:     *newRPCBalance(0),
									Value:                *newRPCBalance(106),
									Nonce:                getUint64Ptr(1),
								},
							},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(900001000000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: delegateCaller(),
								},
								common.Address{0xc2}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							Calls: []TransactionArgs{{
								From:         &common.Address{0xc0},
								To:           &common.Address{0xc1},
								Input:        hex2Bytes("5c19a95c000000000000000000000000c200000000000000000000000000000000000000"),
								Value:        *newRPCBalance(1000),
								MaxFeePerGas: *newRPCBalance(2000),
							}},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(900001000000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: getStorageTester(),
								},
							},
							Calls: []TransactionArgs{
								{
									From:         &common.Address{0xc0},
									To:           &common.Address{0xc1},
									Input:        hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000001"), // set storage slot 0 -> 1
									MaxFeePerGas: *newRPCBalance(2000),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc1},
									Input: hex2Bytes("7b8d56e300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002"), // set storage slot 1 -> 2

									MaxFeePerGas: *newRPCBalance(2000),
								},
							},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc2}: OverrideAccount{
									Code:    selfDestructor(),
									Balance: newRPCBalance(900001000000),
								},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc2},
								Input: hex2Bytes("83197ef0"), //destroy()
							}},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(900001000000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: delegateCaller(),
								},
								common.Address{0xc2}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Input: hex2Bytes("5c19a95c000000000000000000000000c200000000000000000000000000000000000000"),
								Value: *newRPCBalance(1000),
							}},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(900001000000),
								},
								common.Address{0xc2}: OverrideAccount{
									Code: blockHashCallerByteCode(),
								},
							},
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 30)),
							},
							Calls: []TransactionArgs{
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000001"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000002"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000004"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000008"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000016"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000032"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000064"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000128"),
								},
								{
									From:  &common.Address{0xc0},
									To:    &common.Address{0xc2},
									Input: hex2Bytes("ee82ac5e0000000000000000000000000000000000000000000000000000000000000256"),
								},
							},
						},
						{
							StateOverrides: &StateOverride{
								common.Address{0xc0}: OverrideAccount{
									Balance: newRPCBalance(900001000000),
								},
								common.Address{0xc1}: OverrideAccount{
									Code: delegateCaller(),
								},
								common.Address{0xc2}: OverrideAccount{
									Code: getBlockProperties(),
								},
							},
							Calls: []TransactionArgs{{
								From:  &common.Address{0xc0},
								To:    &common.Address{0xc1},
								Input: hex2Bytes("5c19a95c000000000000000000000000c200000000000000000000000000000000000000"),
								Value: *newRPCBalance(1000),
							}},
						},
						{
							Calls: []TransactionArgs{ // just call ecrecover normally
								{ // call with invalid params, should fail (resolve to 0x0)
									From:  &common.Address{0xc1},
									To:    &ecRecoverAddress,
									Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
								},
								{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
									From:  &common.Address{0xc1},
									To:    &ecRecoverAddress,
									Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
								},
							},
						},
						{
							StateOverrides: &StateOverride{ // move ecRecover and call it in new address
								ecRecoverAddress: OverrideAccount{
									MovePrecompileToAddress: &ecRecoverMovedToAddress,
								},
							},
							Calls: []TransactionArgs{
								{ // call with invalid params, should fail (resolve to 0x0)
									From:  &common.Address{0xc1},
									To:    &ecRecoverMovedToAddress,
									Input: hex2Bytes("4554480000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000007b45544800000000000000000000000000000000000000000000000000000000004554480000000000000000000000000000000000000000000000000000000000"),
								},
								{ // call with valid params, should resolve to 0xb11CaD98Ad3F8114E0b3A1F6E7228bc8424dF48a
									From:  &common.Address{0xc1},
									To:    &ecRecoverMovedToAddress,
									Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
								},
								{ // call with valid params, the old address, should fail as it was moved
									From:  &common.Address{0xc1},
									To:    &ecRecoverAddress,
									Input: hex2Bytes("1c8aff950685c2ed4bc3174f3472287b56d9517b9c948127319a09a7a36deac8000000000000000000000000000000000000000000000000000000000000001cb7cf302145348387b9e69fde82d8e634a0f8761e78da3bfa059efced97cbed0d2a66b69167cafe0ccfc726aec6ee393fea3cf0e4f3f9c394705e0f56d9bfe1c9"),
								},
							},
						},
					},
					TraceTransfers:         true,
					Validation:             false,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
		{
			Name:  "ethSimulate-blocknumber-increment",
			About: "blocknumbers should increment",
			Run: func(ctx context.Context, t *T) error {
				latestBlockNumber := t.chain.Head().Number().Int64()
				latestBlockTime := hexutil.Uint64(t.chain.Head().Time())
				params := ethSimulateOpts{
					BlockStateCalls: []CallBatch{
						{
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 10)),
								Time:   getUint64Ptr(latestBlockTime + 500),
							},
						},
						{}, {}, {}, {}, {},
						{
							BlockOverrides: &BlockOverrides{
								Number: (*hexutil.Big)(big.NewInt(latestBlockNumber + 30)),
							},
						},
						{}, {}, {},
					},
					TraceTransfers:         true,
					Validation:             false,
					ReturnFullTransactions: true,
				}
				res := make([]blockResult, 0)
				if err := t.rpc.Call(&res, "eth_simulateV1", params, "latest"); err != nil {
					return err
				}
				return nil
			},
		},
	},
}

// TransactionArgs represents the arguments to construct a new transaction
// or a message call.
type TransactionArgs struct {
	From                 *common.Address `json:"from,omitempty"`
	To                   *common.Address `json:"to,omitempty"`
	Gas                  *hexutil.Uint64 `json:"gas,omitempty"`
	GasPrice             *hexutil.Big    `json:"gasPrice,omitempty"`
	MaxFeePerGas         *hexutil.Big    `json:"maxFeePerGas,omitempty"`
	MaxPriorityFeePerGas *hexutil.Big    `json:"maxPriorityFeePerGas,omitempty"`
	MaxFeePerBlobGas     *hexutil.Big    `json:"maxFeePerBlobGas,omitempty"`
	BlobVersionedHashes  *[]common.Hash  `json:"blobVersionedHashes,omitempty"`
	Value                *hexutil.Big    `json:"value,omitempty"`
	Nonce                *hexutil.Uint64 `json:"nonce,omitempty"`

	// We accept "data" and "input" for backwards-compatibility reasons.
	// "input" is the newer name and should be preferred by clients.
	// Issue detail: https://github.com/ethereum/go-ethereum/issues/15628
	Data  *hexutil.Bytes `json:"data,omitempty"`
	Input *hexutil.Bytes `json:"input,omitempty"`

	// Introduced by AccessListTxType transaction.
	AccessList *types.AccessList `json:"accessList,omitempty"`
	ChainID    *hexutil.Big      `json:"chainId,omitempty"`
}

// BlockOverrides is a set of header fields to override.
type BlockOverrides struct {
	Number        *hexutil.Big    `json:"number,omitempty"`
	Time          *hexutil.Uint64 `json:"time,omitempty"`
	GasLimit      *hexutil.Uint64 `json:"gasLimit,omitempty"`
	FeeRecipient  *common.Address `json:"feeRecipient,omitempty"`
	PrevRandao    *common.Hash    `json:"prevRandao,omitempty"`
	BaseFeePerGas *hexutil.Big    `json:"baseFeePerGas,omitempty"`
	BlobBaseFee   *hexutil.Big    `json:"blobBaseFee,omitempty"`
}

// OverrideAccount indicates the overriding fields of account during the execution
// of a message call.
// Note, state and stateDiff can't be specified at the same time. If state is
// set, message execution will only use the data in the given state. Otherwise
// if statDiff is set, all diff will be applied first and then execute the call
// message.
type OverrideAccount struct {
	Nonce                   *hexutil.Uint64              `json:"nonce,omitempty"`
	Code                    *hexutil.Bytes               `json:"code,omitempty"`
	Balance                 **hexutil.Big                `json:"balance,omitempty"`
	State                   *map[common.Hash]common.Hash `json:"state,omitempty"`
	StateDiff               *map[common.Hash]common.Hash `json:"stateDiff,omitempty"`
	MovePrecompileToAddress *common.Address              `json:"movePrecompileToAddress,omitempty"`
}

// StateOverride is the collection of overridden accounts.
type StateOverride map[common.Address]OverrideAccount

// ethSimulateOpts is the wrapper for ethSimulate parameters.
type ethSimulateOpts struct {
	BlockStateCalls        []CallBatch `json:"blockStateCalls,omitempty"`
	TraceTransfers         bool        `json:"traceTransfers,omitempty"`
	Validation             bool        `json:"validation,omitempty"`
	ReturnFullTransactions bool        `json:"returnFullTransactions,omitempty"`
}

// CallBatch is a batch of calls to be simulated sequentially.
type CallBatch struct {
	BlockOverrides *BlockOverrides   `json:"blockOverrides,omitempty"`
	StateOverrides *StateOverride    `json:"stateOverrides,omitempty"`
	Calls          []TransactionArgs `json:"calls,omitempty"`
}

type blockResult struct {
	Number        hexutil.Uint64 `json:"number"`
	Hash          common.Hash    `json:"hash"`
	Time          hexutil.Uint64 `json:"timestamp"`
	GasLimit      hexutil.Uint64 `json:"gasLimit"`
	GasUsed       hexutil.Uint64 `json:"gasUsed"`
	FeeRecipient  common.Address `json:"feeRecipient"`
	BaseFeePerGas *hexutil.Big   `json:"baseFeePerGas"`
	PrevRandao    *common.Hash   `json:"prevRandao,omitempty"`
	Calls         []callResult   `json:"calls"`
}

type callResult struct {
	ReturnData hexutil.Bytes  `json:"ReturnData"`
	Logs       []*types.Log   `json:"logs"`
	Transfers  []transfer     `json:"transfers,omitempty"`
	GasUsed    hexutil.Uint64 `json:"gasUsed"`
	Status     hexutil.Uint64 `json:"status"`
	Error      errorResult    `json:"error,omitempty"`
}

type errorResult struct {
	Code    *big.Int `json:"code"`
	Message *string  `json:"message"`
}

type transfer struct {
	From  common.Address `json:"from"`
	To    common.Address `json:"to"`
	Value *big.Int       `json:"value"`
}

func newRPCBalance(balance int) **hexutil.Big {
	rpcBalance := (*hexutil.Big)(big.NewInt(int64(balance)))
	return &rpcBalance
}

func hex2Bytes(str string) *hexutil.Bytes {
	rpcBytes := hexutil.Bytes(common.Hex2Bytes(str))
	return &rpcBytes
}
