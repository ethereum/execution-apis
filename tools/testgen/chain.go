package testgen

import (
	"bytes"
	"compress/gzip"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rlp"
)

// Chain is a lightweight blockchain-like store, based on hivechain output files.
// This mostly exists to provide some convenient accessors for the chain.
type Chain struct {
	genesis core.Genesis
	blocks  []*types.Block
	state   map[common.Address]state.DumpAccount // state of head block
	senders map[common.Address]*senderInfo
	txinfo  *ChainTxInfo
	config  *params.ChainConfig
}

// ChainTxInfo is the structure of txinfo.json from hivechain.
type ChainTxInfo struct {
	LegacyTransfers     []TxInfo      `json:"tx-transfer-legacy"`
	AccessListTransfers []TxInfo      `json:"tx-transfer-eip2930"`
	DynamicFeeTransfers []TxInfo      `json:"tx-transfer-eip1559"`
	LegacyEmit          []TxInfo      `json:"tx-emit-legacy"`
	AccessListEmit      []TxInfo      `json:"tx-emit-eip2930"`
	DynamicFeeEmit      []TxInfo      `json:"tx-emit-eip1559"`
	CallMeContract      *ContractInfo `json:"deploy-callme"`
	CallEnvContract     *ContractInfo `json:"deploy-callenv"`
	CallRevertContract  *ContractInfo `json:"deploy-callrevert"`
	EIP7702             *EIP7702Info  `json:"tx-eip7702"`
	EIP7002             *EIP7002Info  `json:"tx-request-eip7002"`
}

// TxInfo is a transaction record created by hivechain.
type TxInfo struct {
	TxHash common.Hash    `json:"txhash"`
	Sender common.Address `json:"sender"`
	Block  hexutil.Uint64 `json:"block"`
	Index  int            `json:"indexInBlock"`
	// for emit contract txs
	LogTopic0 *common.Hash `json:"logtopic0"`
	LogTopic1 *common.Hash `json:"logtopic1"`
}

// ContactInfo is a contract deployment record created by hivechain.
type ContractInfo struct {
	Addr  common.Address `json:"contract"`
	Block hexutil.Uint64 `json:"block"`
}

type EIP7702Info struct {
	Account     common.Address `json:"account"`
	ProxyAddr   common.Address `json:"proxyAddr"`
	AuthorizeTx common.Hash    `json:"authorizeTx"`
}

type EIP7002Info struct {
	TxHash common.Hash    `json:"txhash"`
	Block  hexutil.Uint64 `json:"block"`
}

// NewChain takes the given chain.rlp file, decodes it, and returns
// the blocks from the file.
func NewChain(dir string) (*Chain, error) {
	gen, err := loadGenesis(path.Join(dir, "genesis.json"))
	if err != nil {
		return nil, err
	}
	gblock := gen.ToBlock()

	blocks, err := blocksFromFile(path.Join(dir, "chain.rlp"), gblock)
	if err != nil {
		return nil, err
	}
	state, err := readState(path.Join(dir, "headstate.json"))
	if err != nil {
		return nil, err
	}
	accounts, err := readAccounts(path.Join(dir, "accounts.json"))
	if err != nil {
		return nil, err
	}
	txinfo, err := readTxInfo(path.Join(dir, "txinfo.json"))
	if err != nil {
		return nil, err
	}
	return &Chain{
		genesis: gen,
		blocks:  blocks,
		state:   state,
		senders: accounts,
		txinfo:  txinfo,
		config:  gen.Config,
	}, nil
}

// senderInfo is an account record as output in the "accounts.json" file from
// hivechain.
type senderInfo struct {
	Key   *ecdsa.PrivateKey `json:"key"`
	Nonce uint64            `json:"nonce"`
}

// Head returns the chain head.
func (c *Chain) Head() *types.Block {
	return c.blocks[len(c.blocks)-1]
}

func (c *Chain) Config() *params.ChainConfig {
	return c.genesis.Config
}

// GetBlock returns the block at the specified number.
func (c *Chain) GetBlock(number int) *types.Block {
	return c.blocks[number]
}

// BlockAtTime returns the current block at a given timestamp.
func (c *Chain) BlockAtTime(timestamp uint64) *types.Block {
	for _, b := range c.blocks {
		if b.Time() >= timestamp {
			return b
		}
	}
	return nil
}

// BlockWithTransactions returns a block that has a matching transaction.
func (c *Chain) BlockWithTransactions(matchdesc string, match func(int, *types.Transaction) bool) *types.Block {
	for _, b := range c.blocks {
		for i, tx := range b.Transactions() {
			if match == nil || match(i, tx) {
				return b
			}
		}
	}
	panic(fmt.Sprintf("no block with matching transactions (%s) in chain", matchdesc))
}

// FindTransaction returns a matching transaction.
func (c *Chain) FindTransaction(matchdesc string, match func(int, *types.Transaction) bool) (m *types.Transaction) {
	c.BlockWithTransactions(matchdesc, func(i int, tx *types.Transaction) bool {
		if match == nil || match(i, tx) {
			m = tx
			return true
		}
		return false
	})
	return m
}

// GetSender returns the address associated with account at the index in the
// pre-funded accounts list.
func (c *Chain) GetSender(idx int) (common.Address, uint64) {
	var accounts Addresses
	for addr := range c.senders {
		accounts = append(accounts, addr)
	}
	sort.Sort(accounts)
	addr := accounts[idx]
	return addr, c.senders[addr].Nonce
}

// IncNonce increases the specified signing account's pending nonce.
func (c *Chain) IncNonce(addr common.Address, amt uint64) {
	if _, ok := c.senders[addr]; !ok {
		panic("nonce increment for non-signer")
	}
	c.senders[addr].Nonce += amt
}

// Balance returns the balance of an account at the head of the chain.
func (c *Chain) Balance(addr common.Address) *big.Int {
	bal := new(big.Int)
	if acc, ok := c.state[addr]; ok {
		bal, _ = bal.SetString(acc.Balance, 10)
	}
	return bal
}

// Balance returns the balance of an account at the head of the chain.
func (c *Chain) Storage(addr common.Address, slot common.Hash) []byte {
	v, ok := c.state[addr].Storage[slot]
	if ok {
		return common.LeftPadBytes(hexutil.MustDecode("0x"+v), 32)
	}
	return nil
}

// SignTx signs a transaction for the specified from account, so long as that
// account was in the hivechain accounts dump.
func (c *Chain) MustSignTx(from common.Address, txdata types.TxData) *types.Transaction {
	signer := types.LatestSigner(c.config)
	acc, ok := c.senders[from]
	if !ok {
		panic(fmt.Errorf("account not available for signing: %s", from))
	}
	return types.MustSignNewTx(acc.Key, signer, txdata)
}

// SignAuth signs a SetCode Authorization for the specified from account, so long as that
// account was in the hivechain accounts dump.
func (c *Chain) SignAuth(from common.Address, auth types.SetCodeAuthorization) (types.SetCodeAuthorization, error) {
	acc, ok := c.senders[from]
	if !ok {
		panic(fmt.Errorf("account not available for signing: %s", from))
	}

	signedAuth, err := types.SignSetCode(acc.Key, auth)
	if err != nil {
		return types.SetCodeAuthorization{}, err
	}
	return signedAuth, nil
}

func loadGenesis(genesisFile string) (core.Genesis, error) {
	chainConfig, err := os.ReadFile(genesisFile)
	if err != nil {
		return core.Genesis{}, err
	}
	var gen core.Genesis
	if err := json.Unmarshal(chainConfig, &gen); err != nil {
		return core.Genesis{}, err
	}
	return gen, nil
}

type Addresses []common.Address

func (a Addresses) Len() int {
	return len(a)
}

func (a Addresses) Less(i, j int) bool {
	return bytes.Compare(a[i][:], a[j][:]) < 0
}

func (a Addresses) Swap(i, j int) {
	tmp := a[i]
	a[i] = a[j]
	a[j] = tmp
}

func blocksFromFile(chainfile string, gblock *types.Block) ([]*types.Block, error) {
	// Load chain.rlp.
	fh, err := os.Open(chainfile)
	if err != nil {
		return nil, err
	}
	defer fh.Close()
	var reader io.Reader = fh
	if strings.HasSuffix(chainfile, ".gz") {
		if reader, err = gzip.NewReader(reader); err != nil {
			return nil, err
		}
	}
	stream := rlp.NewStream(reader, 0)
	var blocks = make([]*types.Block, 1)
	blocks[0] = gblock
	for i := 0; ; i++ {
		var b types.Block
		if err := stream.Decode(&b); err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("at block index %d: %v", i, err)
		}
		if b.NumberU64() != uint64(i+1) {
			return nil, fmt.Errorf("block at index %d has wrong number %d", i, b.NumberU64())
		}
		blocks = append(blocks, &b)
	}
	return blocks, nil
}

func readState(file string) (map[common.Address]state.DumpAccount, error) {
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read state: %v", err)
	}
	var dump state.Dump
	if err := json.Unmarshal(f, &dump); err != nil {
		return nil, fmt.Errorf("unable to unmarshal state: %v", err)
	}

	state := make(map[common.Address]state.DumpAccount)
	for key, acct := range dump.Accounts {
		var addr common.Address
		if err := addr.UnmarshalText([]byte(key)); err != nil {
			return nil, fmt.Errorf("invalid address %q", key)
		}
		state[addr] = acct
	}
	return state, nil
}

func readAccounts(file string) (map[common.Address]*senderInfo, error) {
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read accounts: %v", err)
	}
	type account struct {
		Key hexutil.Bytes `json:"key"`
	}
	keys := make(map[common.Address]account)
	if err := json.Unmarshal(f, &keys); err != nil {
		return nil, fmt.Errorf("unable to unmarshal accounts: %v", err)
	}
	accounts := make(map[common.Address]*senderInfo)
	for addr, acc := range keys {
		pk, err := crypto.HexToECDSA(common.Bytes2Hex(acc.Key))
		if err != nil {
			return nil, fmt.Errorf("unable to read private key for %s: %v", err, addr)
		}
		accounts[addr] = &senderInfo{Key: pk, Nonce: 0}
	}
	return accounts, nil
}

func readTxInfo(file string) (*ChainTxInfo, error) {
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("unable to read txinfo: %v", err)
	}
	var txinfo ChainTxInfo
	if err := json.Unmarshal(f, &txinfo); err != nil {
		return nil, err
	}
	return &txinfo, nil
}
