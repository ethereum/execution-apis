package testgen

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

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
	Run   func(context.Context, *ethclient.Client) error
}

// AllMethods is a slice of all JSON-RPC methods with tests.
var AllMethods = []MethodTests{
	EthBlockNumber,
	EthGetBlockByNumber,
}

// EthBlockNumber stores a list of all tests against the method.
var EthBlockNumber = MethodTests{
	"eth_blockNumber",
	[]Test{
		{
			"simple-test",
			"retrieves the client's current block number",
			ethBlockNumber,
		},
	},
}

func ethBlockNumber(ctx context.Context, eth *ethclient.Client) error {
	_, err := eth.BlockNumber(ctx)
	return err
}

// EthGetBlockByNumber stores a list of all tests against the method.
var EthGetBlockByNumber = MethodTests{
	"eth_getBlockByNumber",
	[]Test{
		{
			"get-genesis",
			"gets block 0",
			ethGetBlockByNumberGenesis,
		},
	},
}

func ethGetBlockByNumberGenesis(ctx context.Context, eth *ethclient.Client) error {
	_, err := eth.BlockByNumber(ctx, common.Big0)
	return err
}
