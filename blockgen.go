package main

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// genSimpleChain generates a short chain with a few basic transactions.
func genSimpleChain(engine consensus.Engine) (*core.Genesis, []*types.Block, *types.Block) {
	var (
		keyHex  = "9c647b8b7c4e7c3490668fb6c11473619db80c93704c70893d3813af4090c39c"
		key, _  = crypto.HexToECDSA(keyHex)
		address = crypto.PubkeyToAddress(key.PublicKey) // 658bdf435d810c91414ec09147daa6db62406379
		aa      = common.Address{0xaa}
		bb      = common.Address{0xbb}
		funds   = big.NewInt(0).Mul(big.NewInt(1337), big.NewInt(params.Ether))
		gspec   = &core.Genesis{
			Config:     params.AllEthashProtocolChanges,
			Alloc:      core.GenesisAlloc{address: {Balance: funds}},
			BaseFee:    big.NewInt(params.InitialBaseFee),
			Difficulty: common.Big1,
			GasLimit:   5_000_000,
		}
		gendb  = rawdb.NewMemoryDatabase()
		signer = types.LatestSigner(gspec.Config)
	)
	gspec.Config.TerminalTotalDifficultyPassed = true
	gspec.Config.TerminalTotalDifficulty = common.Big0
	gspec.Config.ShanghaiTime = uintptr(0)

	// init 0xaa with some storage elements
	storage := make(map[common.Hash]common.Hash)
	storage[common.Hash{0x00}] = common.Hash{0x00}
	storage[common.Hash{0x01}] = common.Hash{0x01}
	storage[common.Hash{0x02}] = common.Hash{0x02}
	storage[common.Hash{0x03}] = common.HexToHash("0303")
	gspec.Alloc[aa] = core.GenesisAccount{
		Balance: common.Big1,
		Nonce:   1,
		Storage: storage,
		Code:    common.Hex2Bytes("6042"),
	}
	gspec.Alloc[bb] = core.GenesisAccount{
		Balance: common.Big2,
		Nonce:   1,
		Storage: storage,
		Code:    common.Hex2Bytes("600154600354"),
	}

	genesis := gspec.MustCommit(gendb)

	chain, _ := core.GenerateChain(gspec.Config, genesis, engine, gendb, 4, func(i int, gen *core.BlockGen) {
		tx, _ := types.SignTx(types.NewTransaction(gen.TxNonce(address), address, big.NewInt(1000), params.TxGas, new(big.Int).Add(gen.BaseFee(), common.Big1), nil), signer, key)
		gen.AddTx(tx)
		if i == 1 {
			gen.AddWithdrawal(&types.Withdrawal{
				Index:     123,
				Validator: 42,
				Address:   common.Address{0xee},
				Amount:    1337,
			})
			gen.AddWithdrawal(&types.Withdrawal{
				Index:     124,
				Validator: 13,
				Address:   common.Address{0xee},
				Amount:    1,
			})
		}
	})

	// Modify block so that recorded gas used does not equal actual.
	bad := chain[len(chain)-1]
	h := bad.Header()
	h.GasUsed += 1
	bad.WithSeal(h)
	sealedBlock := make(chan *types.Block, 1)
	if err := engine.Seal(nil, bad, sealedBlock, nil); err != nil {
		panic(err)
	}

	chain = chain[:len(chain)-1]
	return gspec, chain, bad
}

func uintptr(x uint64) *uint64 {
	return &x
}
