package testgen

import (
	"bytes"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

func checkHeaderRLP(t *T, n uint64, got []byte) error {
	head := t.chain.GetHeaderByNumber(n)
	if head == nil {
		return fmt.Errorf("unable to load block %d from test chain", n)
	}
	want, err := rlp.EncodeToBytes(head)
	if err != nil {
		return err
	}
	if hexutil.Encode(got) != hexutil.Encode(want) {
		return fmt.Errorf("unexpected response (got: %s, want: %s)", got, hexutil.Bytes(want))
	}
	return nil
}

func checkBlockRLP(t *T, n uint64, got []byte) error {
	head := t.chain.GetBlockByNumber(n)
	if head == nil {
		return fmt.Errorf("unable to load block %d from test chain", n)
	}
	want, err := rlp.EncodeToBytes(head)
	if err != nil {
		return err
	}
	if hexutil.Encode(got) != hexutil.Encode(want) {
		return fmt.Errorf("unexpected response (got: %s, want: %s)", got, hexutil.Bytes(want))
	}
	return nil
}

func checkBlockReceipts(t *T, n uint64, got []*types.Receipt) error {
	b := t.chain.GetBlockByNumber(n)
	if b == nil {
		return fmt.Errorf("block number %d not found", n)
	}
	want := t.chain.GetReceiptsByHash(b.Hash())
	if len(got) != len(want) {
		return fmt.Errorf("receipts length mismatch (got: %d, want: %d)", len(got), len(want))
	}
	for i := range got {
		got, _ := got[i].MarshalBinary()
		want, _ := want[i].MarshalBinary()
		if !bytes.Equal(got, want) {
			return fmt.Errorf("receipt mismatch (got: %x, want: %x)", got, want)
		}
	}
	return nil
}
