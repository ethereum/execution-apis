package testgen

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
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
