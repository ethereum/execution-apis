package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/rlp"
)

// Client is an interface for generically interacting with Ethereum clients.
type Client interface {
	// Start starts client, but does not wait for the command to exit.
	Start(ctx context.Context, verbose bool) error

	// HttpAddr returns the address where the client is servering its
	// JSON-RPC.
	HttpAddr() string

	// Close closes the client.
	Close() error
}

// GethClient is a wrapper around a go-ethereum instance on a separate thread.
type GethClient struct {
	cmd     *exec.Cmd
	path    string
	workdir string
	blocks  []*types.Block
	genesis *core.Genesis
}

// NewGethClient instantiates a new GethClient.
//
// The client's data directory is set to a temporary location and it
// initializes with the genesis and the provided blocks.
func NewGethClient(ctx context.Context, path string, genesis *core.Genesis, blocks []*types.Block, verbose bool) (*GethClient, error) {
	tmp, err := ioutil.TempDir("", "geth-")
	if err != nil {
		return nil, err
	}
	writeChain(tmp, genesis, blocks)

	var (
		isFakepow = !(ctx.Value("args").(*Args)).Ethash
		datadir   = fmt.Sprintf("--datadir=%s", tmp)
	)

	// Run geth init.
	options := []string{datadir, "init", fmt.Sprintf("%s/genesis.json", tmp)}
	options = maybePrepend(isFakepow, options, "--fakepow")
	err = runCmd(ctx, path, verbose, options...)
	if err != nil {
		return nil, err
	}

	// Run geth import.
	options = []string{datadir, "import", fmt.Sprintf("%s/chain.rlp", tmp)}
	options = maybePrepend(isFakepow, options, "--fakepow")
	err = runCmd(ctx, path, verbose, options...)
	if err != nil {
		return nil, err
	}

	return &GethClient{path: path, genesis: genesis, blocks: blocks, workdir: tmp}, nil
}

// Start starts geth, but does not wait for the command to exit.
func (g *GethClient) Start(ctx context.Context, verbose bool) error {
	fmt.Println("starting client")

	isFakepow := !(ctx.Value("args").(*Args)).Ethash
	options := []string{
		fmt.Sprintf("--datadir=%s", g.workdir),
		fmt.Sprintf("--port=%s", NETWORKPORT),
		"--nodiscover",
		"--http",
		fmt.Sprintf("--http.addr=%s", HOST),
		fmt.Sprintf("--http.port=%s", PORT),
	}
	options = maybePrepend(isFakepow, options, "--fakepow")
	g.cmd = exec.CommandContext(
		ctx,
		g.path,
		options...,
	)
	if verbose {
		g.cmd.Stdout = os.Stdout
		g.cmd.Stderr = os.Stderr
	}
	if err := g.cmd.Start(); err != nil {
		return err
	}
	return nil
}

// HttpAddr returns the address where the client is servering its JSON-RPC.
func (g *GethClient) HttpAddr() string {
	return fmt.Sprintf("http://%s:%s", HOST, PORT)
}

// Close closes the client.
func (g *GethClient) Close() error {
	g.cmd.Process.Kill()
	g.cmd.Wait()
	return os.RemoveAll(g.workdir)
}

// runCmd runs a command and outputs the command's stdout and stderr to the
// caller's stdout and stderr if verbose is set.
func runCmd(ctx context.Context, path string, verbose bool, args ...string) error {
	cmd := exec.CommandContext(ctx, path, args...)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}

// writeChain writes the genesis and blocks to disk.
func writeChain(path string, genesis *core.Genesis, blocks []*types.Block) error {
	out, err := json.Marshal(genesis)
	if err != nil {
		return err
	}
	if err := os.WriteFile(fmt.Sprintf("%s/genesis.json", path), out, 0644); err != nil {
		return err
	}
	w, err := os.OpenFile(fmt.Sprintf("%s/chain.rlp", path), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer w.Close()
	for _, block := range blocks {
		if err := rlp.Encode(w, block); err != nil {
			return err
		}
	}
	return nil
}

func maybePrepend(shouldAdd bool, options []string, maybe ...string) []string {
	if shouldAdd {
		options = append(maybe, options...)
	}
	return options
}
