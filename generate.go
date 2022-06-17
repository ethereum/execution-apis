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

	// Connect ethclient to Ethereum client.
	handler, err := NewEthclientHandler(client.HttpAddr())
	if err != nil {
		return err
	}
	defer handler.Close()

	// Generate test fixtures for all methods. Store them in the format:
	// outputDir/methodName/testName.io
	tests := testgen.AllMethods
	for _, methodTest := range tests {
		methodDir := fmt.Sprintf("%s/%s", args.OutDir, methodTest.MethodName)
		if err := mkdir(methodDir); err != nil {
			return err
		}

		for _, test := range methodTest.Tests {
			// Write the exchange for each test in a separte file.
			handler.RotateLog(fmt.Sprintf("%s/%s.io", methodDir, test.Name))

			// Fail test fill if request exceeds timeout.
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			err := test.Run(ctx, handler.ethclient)
			if err != nil {
				fmt.Fprintf(os.Stderr, "failed to fill %s/%s: %s\n", methodTest.MethodName, test.Name, err)
			}
		}
	}
	return nil
}

// spawnClient starts an Ethereum client on a separate thread.
//
// It waits until the client is responding to JSON-RPC requests
// before returning.
func spawnClient(ctx context.Context, args *Args) (Client, error) {
	var (
		client        Client
		err           error
		gspec, blocks = genSimpleChain()
	)

	// Initialize specified client and start it in a separate thread.
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

	// Try to connect for 5 seconds. Error otherwise.
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err = tryConnection(ctx, fmt.Sprintf("http://%s:%s", HOST, PORT), 500*time.Millisecond)
	if err != nil {
		return nil, err
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

// tryConnection checks if a client's JSON-RPC API is accepting requests.
func tryConnection(ctx context.Context, addr string, waitTime time.Duration) error {
	c, err := rpc.DialHTTPWithClient(addr, http.DefaultClient)
	if err != nil {
		return err
	}
	e := ethclient.NewClient(c)
	for {
		if _, err := e.BlockNumber(ctx); err == nil {
			break
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry timeout: %w", err)
		case <-time.After(waitTime):
		}
	}
	return nil
}
