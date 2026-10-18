package testgen

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

// The trace fixtures stay on blocks before Amsterdam, the forks the trace methods are
// specified and measured against. Only trace_rawTransaction runs at the head.

// preAmsterdam returns the last listed transaction mined before Amsterdam.
func (c *Chain) preAmsterdam(txs []TxInfo) TxInfo {
	for i := len(txs) - 1; i >= 0; i-- {
		block := c.GetBlock(int(txs[i].Block))
		if c.config.AmsterdamTime == nil || block.Time() < *c.config.AmsterdamTime {
			return txs[i]
		}
	}
	panic("no transaction before Amsterdam")
}

// blockWithUncle returns the first block that includes an uncle, so its trace
// carries both reward records.
func (c *Chain) blockWithUncle() *types.Block {
	for _, block := range c.blocks {
		if len(block.Uncles()) > 0 {
			return block
		}
	}
	panic("no block with an uncle in chain")
}

// callArgs turns a mined transaction into the call object that replays it.
func callArgs(c *Chain, tx *types.Transaction) map[string]any {
	from, err := types.Sender(types.LatestSigner(c.config), tx)
	if err != nil {
		panic(err)
	}
	return map[string]any{
		"from":  from,
		"to":    tx.To(),
		"gas":   hexutil.Uint64(tx.Gas()),
		"value": (*hexutil.Big)(tx.Value()),
		"input": hexutil.Bytes(tx.Data()),
	}
}

func (c *Chain) transaction(info TxInfo) *types.Transaction {
	block := c.GetBlock(int(info.Block))
	for _, tx := range block.Transactions() {
		if tx.Hash() == info.TxHash {
			return tx
		}
	}
	panic(fmt.Sprintf("transaction %s not in block %d", info.TxHash, info.Block))
}

// traceCall calls a trace method and checks that it returned a non-empty result.
func traceCall(ctx context.Context, t *T, method string, args ...any) (json.RawMessage, error) {
	var result json.RawMessage
	if err := t.rpc.CallContext(ctx, &result, method, args...); err != nil {
		return nil, err
	}
	if len(result) == 0 || string(result) == "null" || string(result) == "[]" {
		return nil, fmt.Errorf("%s returned %s", method, result)
	}
	return result, nil
}

// traceNull calls a trace method that must answer null.
func traceNull(ctx context.Context, t *T, method string, args ...any) error {
	var result json.RawMessage
	if err := t.rpc.CallContext(ctx, &result, method, args...); err != nil {
		return err
	}
	if string(result) != "null" {
		return fmt.Errorf("%s returned %s, want null", method, result)
	}
	return nil
}

var unknownHash = common.HexToHash("0x0000000000000000000000000000000000000000000000000000000000000001")

var TraceBlock = MethodTests{
	"trace_block",
	[]Test{
		{
			Name:  "trace-block-rewards",
			About: "traces a proof-of-work block with an uncle; the records end with the block reward and then the uncle reward",
			Run: func(ctx context.Context, t *T) error {
				block := t.chain.blockWithUncle()
				_, err := traceCall(ctx, t, "trace_block", hexutil.EncodeUint64(block.NumberU64()))
				return err
			},
		},
		{
			Name:  "trace-block-call-tree",
			About: "traces a post-merge block containing a nested call tree; post-merge blocks have no reward records",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				_, err := traceCall(ctx, t, "trace_block", hexutil.EncodeUint64(uint64(info.Block)))
				return err
			},
		},
		{
			Name:  "trace-block-unknown",
			About: "traces a block number past the head, which is unknown; the result is null or an error (-32001 recommended)",
			Run: func(ctx context.Context, t *T) error {
				err := traceNull(ctx, t, "trace_block", hexutil.EncodeUint64(t.chain.Head().NumberU64()+1000))
				var rpcErr rpc.Error
				if errors.As(err, &rpcErr) {
					return nil
				}
				return err
			},
		},
	},
}

var TraceTransaction = MethodTests{
	"trace_transaction",
	[]Test{
		{
			Name:  "trace-transaction-call-tree",
			About: "traces a transaction whose call fans out into nested calls; the frames are a preorder tree",
			Run: func(ctx context.Context, t *T) error {
				_, err := traceCall(ctx, t, "trace_transaction", t.chain.preAmsterdam(t.chain.txinfo.CallTree).TxHash)
				return err
			},
		},
		{
			Name:  "trace-transaction-revert",
			About: "traces a transaction that reverts; the root frame reports the revert",
			Run: func(ctx context.Context, t *T) error {
				_, err := traceCall(ctx, t, "trace_transaction", t.chain.preAmsterdam(t.chain.txinfo.CallRevert).TxHash)
				return err
			},
		},
		{
			Name:  "trace-transaction-unknown",
			About: "traces an unknown transaction hash; the result is null",
			Run: func(ctx context.Context, t *T) error {
				return traceNull(ctx, t, "trace_transaction", unknownHash)
			},
		},
	},
}

var TraceGet = MethodTests{
	"trace_get",
	[]Test{
		{
			Name:  "trace-get-root",
			About: "gets the root frame of a call tree with an empty path",
			Run: func(ctx context.Context, t *T) error {
				_, err := traceCall(ctx, t, "trace_get", t.chain.preAmsterdam(t.chain.txinfo.CallTree).TxHash, []hexutil.Uint64{})
				return err
			},
		},
		{
			Name:  "trace-get-child",
			About: "gets the first child frame of a call tree by its trace address",
			Run: func(ctx context.Context, t *T) error {
				_, err := traceCall(ctx, t, "trace_get", t.chain.preAmsterdam(t.chain.txinfo.CallTree).TxHash, []hexutil.Uint64{0})
				return err
			},
		},
		{
			Name:  "trace-get-missing-path",
			About: "gets a trace address that the call tree does not have; the result is null",
			Run: func(ctx context.Context, t *T) error {
				return traceNull(ctx, t, "trace_get", t.chain.preAmsterdam(t.chain.txinfo.CallTree).TxHash, []hexutil.Uint64{1000})
			},
		},
	},
}

var TraceReplayTransaction = MethodTests{
	"trace_replayTransaction",
	[]Test{
		{
			Name:  "trace-replay-call-tree",
			About: "replays a call tree transaction with trace and stateDiff",
			Run: func(ctx context.Context, t *T) error {
				_, err := traceCall(ctx, t, "trace_replayTransaction", t.chain.preAmsterdam(t.chain.txinfo.CallTree).TxHash, []string{"trace", "stateDiff"})
				return err
			},
		},
		{
			Name:  "trace-replay-vm-trace",
			About: "replays a log-emitting transaction with vmTrace",
			Run: func(ctx context.Context, t *T) error {
				_, err := traceCall(ctx, t, "trace_replayTransaction", t.chain.preAmsterdam(t.chain.txinfo.DynamicFeeEmit).TxHash, []string{"vmTrace"})
				return err
			},
		},
		{
			Name:  "trace-replay-unknown",
			About: "replays an unknown transaction hash; the result is null",
			Run: func(ctx context.Context, t *T) error {
				return traceNull(ctx, t, "trace_replayTransaction", unknownHash, []string{"trace"})
			},
		},
	},
}

var TraceReplayBlockTransactions = MethodTests{
	"trace_replayBlockTransactions",
	[]Test{
		{
			Name:  "trace-replay-block",
			About: "replays every transaction of a post-merge block in order with trace and stateDiff",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				_, err := traceCall(ctx, t, "trace_replayBlockTransactions", hexutil.EncodeUint64(uint64(info.Block)), []string{"trace", "stateDiff"})
				return err
			},
		},
	},
}

var TraceFilter = MethodTests{
	"trace_filter",
	[]Test{
		{
			Name:  "trace-filter-range",
			About: "filters a range of proof-of-work blocks, including their reward records",
			Run: func(ctx context.Context, t *T) error {
				block := t.chain.blockWithUncle().NumberU64()
				_, err := traceCall(ctx, t, "trace_filter", map[string]any{
					"fromBlock": hexutil.EncodeUint64(block),
					"toBlock":   hexutil.EncodeUint64(block + 2),
				})
				return err
			},
		},
		{
			Name:  "trace-filter-to-address",
			About: "filters a range for frames calling the call tree's target",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				tx := t.chain.transaction(info)
				_, err := traceCall(ctx, t, "trace_filter", map[string]any{
					"fromBlock": hexutil.EncodeUint64(uint64(info.Block) - 5),
					"toBlock":   hexutil.EncodeUint64(uint64(info.Block)),
					"toAddress": []common.Address{*tx.To()},
				})
				return err
			},
		},
		{
			Name:  "trace-filter-block-hash",
			About: "selects one block by hash with the blockHash member",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				_, err := traceCall(ctx, t, "trace_filter", map[string]any{"blockHash": t.chain.GetBlock(int(info.Block)).Hash()})
				return err
			},
		},
		{
			Name:  "trace-filter-page",
			About: "pages through a range with after and count",
			Run: func(ctx context.Context, t *T) error {
				block := t.chain.blockWithUncle().NumberU64()
				_, err := traceCall(ctx, t, "trace_filter", map[string]any{
					"fromBlock": hexutil.EncodeUint64(block),
					"toBlock":   hexutil.EncodeUint64(block + 2),
					"after":     1,
					"count":     2,
				})
				return err
			},
		},
	},
}

var TraceCall = MethodTests{
	"trace_call",
	[]Test{
		{
			Name:  "trace-call-call-tree",
			About: "simulates a mined call tree transaction as an unpriced call on its parent block",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				args := callArgs(t.chain, t.chain.transaction(info))
				_, err := traceCall(ctx, t, "trace_call", args, []string{"trace"}, hexutil.EncodeUint64(uint64(info.Block)-1))
				return err
			},
		},
		{
			Name:  "trace-call-state-diff",
			About: "simulates a value transfer to a fresh account with stateDiff; the account is born with the value",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				from := t.chain.transaction(info)
				args := callArgs(t.chain, from)
				args["to"] = common.HexToAddress("0x00000000000000000000000000000000000c0de1")
				args["input"] = hexutil.Bytes{}
				args["value"] = (*hexutil.Big)(big.NewInt(params.GWei))
				_, err := traceCall(ctx, t, "trace_call", args, []string{"trace", "stateDiff"}, hexutil.EncodeUint64(uint64(info.Block)-1))
				return err
			},
		},
		{
			Name:  "trace-call-create",
			About: "simulates a creation whose initcode returns one byte of code",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				args := callArgs(t.chain, t.chain.transaction(info))
				delete(args, "to")
				// PUSH1 0xfe PUSH1 0 MSTORE8 PUSH1 1 PUSH1 0 RETURN
				args["input"] = hexutil.Bytes(hexutil.MustDecode("0x60fe60005360016000f3"))
				args["value"] = (*hexutil.Big)(new(big.Int))
				_, err := traceCall(ctx, t, "trace_call", args, []string{"trace"}, hexutil.EncodeUint64(uint64(info.Block)-1))
				return err
			},
		},
	},
}

var TraceCallMany = MethodTests{
	"trace_callMany",
	[]Test{
		{
			Name:  "trace-call-many-sequence",
			About: "simulates two calls in sequence; the second runs on the state the first left",
			Run: func(ctx context.Context, t *T) error {
				info := t.chain.preAmsterdam(t.chain.txinfo.CallTree)
				call := callArgs(t.chain, t.chain.transaction(info))
				transfer := callArgs(t.chain, t.chain.transaction(info))
				transfer["to"] = common.HexToAddress("0x00000000000000000000000000000000000c0de2")
				transfer["input"] = hexutil.Bytes{}
				transfer["value"] = (*hexutil.Big)(big.NewInt(params.GWei))
				calls := []any{[]any{call, []string{"trace"}}, []any{transfer, []string{"trace", "stateDiff"}}}
				_, err := traceCall(ctx, t, "trace_callMany", calls, hexutil.EncodeUint64(uint64(info.Block)-1))
				return err
			},
		},
	},
}

var TraceRawTransaction = MethodTests{
	"trace_rawTransaction",
	[]Test{
		{
			Name:  "trace-raw-transfer",
			About: "traces a signed value transfer at the head without submitting it",
			Run: func(ctx context.Context, t *T) error {
				// The last prefunded account by address without code; its state nonce is
				// used, so transactions other tests submit to the pool do not shift it.
				// The recipient exists, so no account is created at the Amsterdam head.
				var sender common.Address
				for i := len(t.chain.senders) - 1; i >= 0; i-- {
					if sender, _ = t.chain.GetSender(i); len(t.chain.state[sender].Code) == 0 {
						break
					}
				}
				recipient, _ := t.chain.GetSender(0)
				head := t.chain.Head()
				tx := t.chain.MustSignTx(sender, &types.DynamicFeeTx{
					ChainID:   t.chain.config.ChainID,
					Nonce:     t.chain.state[sender].Nonce,
					GasTipCap: big.NewInt(params.GWei),
					GasFeeCap: new(big.Int).Add(new(big.Int).Mul(head.BaseFee(), big.NewInt(2)), big.NewInt(params.GWei)),
					Gas:       21000,
					To:        &recipient,
					Value:     big.NewInt(params.GWei),
				})
				raw, err := tx.MarshalBinary()
				if err != nil {
					return err
				}
				_, err = traceCall(ctx, t, "trace_rawTransaction", hexutil.Bytes(raw), []string{"trace", "stateDiff"})
				return err
			},
		},
	},
}
