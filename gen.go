package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type test struct {
	Name  string
	About string
	Run   func(context.Context, *ethclient.Client)
}

var tests = []test{
	{"eth_blockNumber", "simple test", EthBlockNumber},
}

func EthBlockNumber(ctx context.Context, eth *ethclient.Client) {
	eth.BlockNumber(ctx)
}

func runGenerator(ctx context.Context) error {
	args := ctx.Value("args").(*Args)
	client, err := spawnClient(ctx, args)
	if err != nil {
		return err
	}
	defer client.Close()

	time.Sleep(time.Second)

	handler, err := NewEthclientHandler(client.HttpAddr())
	if err != nil {
		return err
	}
	defer handler.Close()

	for _, test := range tests {
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()
		handler.RotateLog(fmt.Sprintf("%s.json", test.Name))
		test.Run(ctx, handler.ethclient)
	}

	return nil
}

func spawnClient(ctx context.Context, args *Args) (Client, error) {
	var (
		client        Client
		err           error
		gspec, blocks = genSimpleChain()
	)
	switch args.ClientType {
	case "geth":
		client, err = NewGethClient(ctx, args.ClientBin, gspec, blocks, args.Verbose)
		if err != nil {
			return nil, err
		}
		client.Start(ctx, args.Verbose)
	default:
		return nil, fmt.Errorf("unsupported client: %s", args.ClientType)
	}
	c, err := rpc.DialHTTPWithClient(fmt.Sprintf("http://%s:%s", HOST, PORT), http.DefaultClient)
	if err != nil {
		return nil, err
	}
	eth := ethclient.NewClient(c)
	for {
		_, err := eth.BlockNumber(ctx)
		if err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout while fetching information (last error: %w)", err)
		case <-time.After(500 * time.Millisecond):
		}
	}
	return client, nil
}
