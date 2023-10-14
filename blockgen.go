package main

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/holiman/uint256"
)

// genSimpleChain generates a short chain with a few basic transactions.
func genSimpleChain(engine consensus.Engine) (*core.Genesis, []*types.Block, *types.Block) {
	var (
		keyHex   = "9c647b8b7c4e7c3490668fb6c11473619db80c93704c70893d3813af4090c39c"
		key, _   = crypto.HexToECDSA(keyHex)
		address  = crypto.PubkeyToAddress(key.PublicKey) // 658bdf435d810c91414ec09147daa6db62406379
		aa       = common.Address{0xaa}
		bb       = common.Address{0xbb}
		funds    = big.NewInt(0).Mul(big.NewInt(1337), big.NewInt(params.Ether))
		contract = common.HexToAddress("0000000000000000000000000000000000031ec7")
		gspec    = &core.Genesis{
			Config: params.AllEthashProtocolChanges,
			Alloc: core.GenesisAlloc{
				address: {Balance: funds},
				common.HexToAddress("a94f5374fce5edbc8e2a8697c15331677e6ebf0b"): {Balance: funds},
			},
			BaseFee:    big.NewInt(params.InitialBaseFee),
			Difficulty: common.Big1,
			GasLimit:   5_000_000,
		}
		gendb = rawdb.NewMemoryDatabase()
	)
	gspec.Config.TerminalTotalDifficultyPassed = true
	gspec.Config.TerminalTotalDifficulty = common.Big0
	gspec.Config.ShanghaiTime = uintptr(0)
	gspec.Config.CancunTime = uintptr(0)

	signer := types.LatestSigner(gspec.Config)

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

	// // SPDX-License-Identifier: GPL-3.0
	// pragma solidity >=0.7.0 <0.9.0;
	//
	// contract Token {
	//     event Transfer(address indexed from, address indexed to, uint256 value);
	//     function transfer(address to, uint256 value) public returns (bool) {
	//         emit Transfer(msg.sender, to, value);
	//         return true;
	//     }
	// }
	gspec.Alloc[contract] = core.GenesisAccount{
		Balance: big.NewInt(params.Ether),
		Code:    common.FromHex("0x608060405234801561001057600080fd5b506004361061002b5760003560e01c8063a9059cbb14610030575b600080fd5b61004a6004803603810190610045919061016a565b610060565b60405161005791906101c5565b60405180910390f35b60008273ffffffffffffffffffffffffffffffffffffffff163373ffffffffffffffffffffffffffffffffffffffff167fddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef846040516100bf91906101ef565b60405180910390a36001905092915050565b600080fd5b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000610101826100d6565b9050919050565b610111816100f6565b811461011c57600080fd5b50565b60008135905061012e81610108565b92915050565b6000819050919050565b61014781610134565b811461015257600080fd5b50565b6000813590506101648161013e565b92915050565b60008060408385031215610181576101806100d1565b5b600061018f8582860161011f565b92505060206101a085828601610155565b9150509250929050565b60008115159050919050565b6101bf816101aa565b82525050565b60006020820190506101da60008301846101b6565b92915050565b6101e981610134565b82525050565b600060208201905061020460008301846101e0565b9291505056fea2646970667358221220b469033f4b77b9565ee84e0a2f04d496b18160d26034d54f9487e57788fd36d564736f6c63430008120033"),
	}

	genesis := gspec.MustCommit(gendb, trie.NewDatabase(gendb, trie.HashDefaults))

	chain, _ := core.GenerateChain(gspec.Config, genesis, engine, gendb, 10, func(i int, gen *core.BlockGen) {
		gen.SetParentBeaconRoot(common.Hash{byte(i)})
		var (
			tx  *types.Transaction
			err error
		)
		switch i {
		case 2:
			// create contract
			tx, err = types.SignTx(types.NewTx(&types.LegacyTx{Nonce: uint64(i), To: nil, Gas: 53100, GasPrice: gen.BaseFee(), Data: common.FromHex("0x60806040")}), signer, key)
		case 3:
			// with logs
			// transfer(address to, uint256 value)
			data := fmt.Sprintf("0xa9059cbb%s%s", common.HexToHash(common.BigToAddress(big.NewInt(int64(i + 1))).Hex()).String()[2:], common.BytesToHash([]byte{byte(i + 11)}).String()[2:])
			tx, err = types.SignTx(types.NewTx(&types.LegacyTx{Nonce: uint64(i), To: &contract, Gas: 60000, GasPrice: gen.BaseFee(), Data: common.FromHex(data)}), signer, key)
		case 4:
			// dynamic fee with logs
			// transfer(address to, uint256 value)
			data := fmt.Sprintf("0xa9059cbb%s%s", common.HexToHash(common.BigToAddress(big.NewInt(int64(i + 1))).Hex()).String()[2:], common.BytesToHash([]byte{byte(i + 11)}).String()[2:])
			fee := big.NewInt(500)
			fee.Add(fee, gen.BaseFee())
			tx, err = types.SignTx(types.NewTx(&types.DynamicFeeTx{Nonce: uint64(i), To: &contract, Gas: 60000, Value: big.NewInt(1), GasTipCap: big.NewInt(500), GasFeeCap: fee, Data: common.FromHex(data)}), signer, key)
		case 5:
			// access list with contract create
			accessList := types.AccessList{{
				Address:     contract,
				StorageKeys: []common.Hash{{0}},
			}}
			tx, err = types.SignTx(types.NewTx(&types.AccessListTx{Nonce: uint64(i), To: nil, Gas: 58100, GasPrice: gen.BaseFee(), Data: common.FromHex("0x60806040"), AccessList: accessList}), signer, key)
		case 6:
			fee := uint256.NewInt(500)
			fee.Add(fee, uint256.MustFromBig(gen.BaseFee()))
			data := fmt.Sprintf("0xa9059cbb%s%s", common.HexToHash(common.BigToAddress(big.NewInt(int64(i + 1))).Hex()).String()[2:], common.BytesToHash([]byte{byte(i + 11)}).String()[2:])

			// blob tx
			inner := types.BlobTx{
				Nonce:      uint64(i),
				To:         contract,
				Gas:        60000,
				Value:      uint256.NewInt(1),
				GasTipCap:  uint256.NewInt(500),
				GasFeeCap:  fee,
				Data:       common.FromHex(data),
				BlobFeeCap: uint256.NewInt(params.BlobTxBlobGasPerBlob),
				BlobHashes: []common.Hash{{0x01, 0x42}},
			}
			tx, err = types.SignTx(types.NewTx(&inner), signer, key)
		default:
			tx, err = types.SignTx(types.NewTransaction(gen.TxNonce(address), address, big.NewInt(1000), params.TxGas, new(big.Int).Add(gen.BaseFee(), common.Big1), nil), signer, key)
		}
		if err != nil {
			panic(err)
		}

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
