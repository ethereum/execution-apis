package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/lightclient/rpctestgen/testgen"
)

// runGenerator generates test fixtures against the specified client and writes
// them to the output directory.
func runGenerator(ctx context.Context) error {
	args := ctx.Value("args").(*Args)

	// Start Ethereum client.
	client, err := spawnClient(ctx, args)
	if err != nil {
		return err
	}
	defer client.Close()

	// Connect ethclient to Ethreum client.
	handler, err := NewEthclientHandler(client.HttpAddr())
	if err != nil {
		return err
	}
	defer handler.Close()

	// Generate test fixtures for all methods.
	tests := testgen.AllMethods
	for _, methodTest := range tests {
		methodDir := fmt.Sprintf("%s/%s", args.OutDir, methodTest.MethodName)
		if err := mkdir(methodDir); err != nil {
			return err
		}
		for _, test := range methodTest.Tests {
			handler.RotateLog(fmt.Sprintf("%s/%s.io", methodDir, test.Name))
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()
			test.Run(ctx, handler.ethclient)
		}
	}
	return nil
}

// spawnClient starts an Ethereum client on separate thread.
//
// It waits until the client is responding to JSON-RPC requests
// before returning.
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

	// Try to connect for 5 seconds.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
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

// mkdir makes a directory at the specified path, if it doesn't already exist.
func mkdir(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
