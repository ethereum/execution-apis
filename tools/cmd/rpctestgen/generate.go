package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cespare/cp"
	"github.com/ethereum/execution-apis/tools/cmd/rpctestgen/testgen"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// runGenerator generates test fixtures against the specified client and writes
// them to the output directory.
func runGenerator(ctx context.Context) error {
	args := ctx.Value(ARGS).(*Args)

	// Initialize generated chain.
	chain, err := testgen.NewChain(args.ChainDir)
	if err != nil {
		return err
	}
	if err := copyChainFiles(args.ChainDir, args.OutDir); err != nil {
		return err
	}

	// Start Ethereum client.
	client, err := spawnClient(ctx, args)
	if err != nil {
		return err
	}
	defer client.Close()

	err = client.AfterStart(ctx)
	if err != nil {
		return err
	}

	// Generate test fixtures for all methods. Store them in the format:
	// outputDir/methodName/testName.io
	fmt.Println("filling tests...")
	tests := testgen.AllMethods
	fails := 0
	for _, methodTest := range tests {
		// Skip tests that don't match regexp.
		if !args.tests.MatchString(methodTest.Name) {
			continue
		}

		methodDir := fmt.Sprintf("%s/%s", args.OutDir, methodTest.Name)
		if err := mkdir(methodDir); err != nil {
			return err
		}
		for _, test := range methodTest.Tests {
			filename := fmt.Sprintf("%s/%s.io", methodDir, test.Name)
			fmt.Printf("generating %s", filename)

			// Connect ethclient to Ethereum client. This happens
			// every test to force the json-rpc id to always be 0.
			handler, err := newEthclientHandler(client.HttpAddr())
			if err != nil {
				return err
			}

			// Write the exchange for each test in a separte file.
			handler.RotateLog(filename)
			if test.About != "" {
				handler.WriteComment(test.About)
			}
			if test.SpecOnly {
				if test.About != "" {
					handler.WriteComment("")
				}
				handler.WriteComment("speconly: client response is only checked for schema validity.")
			}

			// Fail test fill if request exceeds timeout.
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			err = test.Run(ctx, testgen.NewT(handler.rpc, chain))
			if err != nil {
				fmt.Println(" fail.")
				fmt.Fprintf(os.Stderr, "failed to fill %s/%s: %s\n", methodTest.Name, test.Name, err)
				fails++
				continue
			}
			fmt.Println("  done.")
			handler.Close()
		}
	}

	if fails > 0 {
		return fmt.Errorf("%d tests failed to fill", fails)
	}
	return nil
}

// spawnClient starts an Ethereum client on a separate thread.
//
// It waits until the client is responding to JSON-RPC requests
// before returning.
func spawnClient(ctx context.Context, args *Args) (Client, error) {
	var (
		client Client
		err    error
	)

	// Initialize specified client and start it in a separate thread.
	switch args.ClientType {
	case "geth":
		client, err = newGethClient(ctx, args.ClientBin, args.ChainDir, args.Verbose)
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

// copyChainFiles copies the chain into the tests output directory.
// Here we copy all files required to run the tests.
func copyChainFiles(chainDir, outDir string) error {
	files := []string{
		"genesis.json",
		"chain.rlp",
		"forkenv.json",
		"headfcu.json",
	}
	err := os.MkdirAll(outDir, 0755)
	if err != nil {
		return err
	}
	for _, f := range files {
		fmt.Println("copying", f)
		err = cp.CopyFileOverwrite(filepath.Join(outDir, f), filepath.Join(chainDir, f))
		if err != nil {
			return err
		}
	}
	return nil
}

// tryConnection checks if a client's JSON-RPC API is accepting requests.
func tryConnection(ctx context.Context, addr string, waitTime time.Duration) error {
	c, err := rpc.DialOptions(ctx, addr)
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
