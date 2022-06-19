package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// genSimpleChain generates a short chain with a few basic transactions.
func genSimpleChain(engine consensus.Engine) (*core.Genesis, []*types.Block) {
	var (
		keyHex  = "9c647b8b7c4e7c3490668fb6c11473619db80c93704c70893d3813af4090c39c"
		key, _  = crypto.HexToECDSA(keyHex)
		address = crypto.PubkeyToAddress(key.PublicKey) // 658bdf435d810c91414ec09147daa6db62406379
		funds   = big.NewInt(0).Mul(big.NewInt(1337), big.NewInt(params.Ether))
		gspec   = &core.Genesis{
			Config:     params.TestChainConfig,
			Alloc:      core.GenesisAlloc{address: {Balance: funds}},
			BaseFee:    big.NewInt(params.InitialBaseFee),
			Difficulty: common.Big1,
		}
		gendb   = rawdb.NewMemoryDatabase()
		genesis = gspec.MustCommit(gendb)
		signer  = types.LatestSigner(gspec.Config)
	)
	sealingEngine := sealingEngine{engine}
	chain, _ := core.GenerateChain(gspec.Config, genesis, sealingEngine, gendb, 3, func(i int, gen *core.BlockGen) {
		tx, _ := types.SignTx(types.NewTransaction(gen.TxNonce(address), address, big.NewInt(1000), params.TxGas, gen.BaseFee(), nil), signer, key)
		gen.AddTx(tx)
	})
	return gspec, chain
}

// sealingEngine overrides FinalizeAndAssemble and performs sealing in-place.
type sealingEngine struct{ consensus.Engine }

func (e sealingEngine) FinalizeAndAssemble(chain consensus.ChainHeaderReader, header *types.Header, state *state.StateDB, txs []*types.Transaction, uncles []*types.Header, receipts []*types.Receipt) (*types.Block, error) {
	block, err := e.Engine.FinalizeAndAssemble(chain, header, state, txs, uncles, receipts)
	if err != nil {
		return nil, err
	}
	sealedBlock := make(chan *types.Block, 1)
	if err = e.Engine.Seal(nil, block, sealedBlock, nil); err != nil {
		return nil, err
	}
	fmt.Printf("sealing block %d\n", header.Number.Uint64())
	return <-sealedBlock, nil
}
