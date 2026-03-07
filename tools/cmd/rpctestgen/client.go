package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/rpc"
)

// Client is an interface for generically interacting with Ethereum clients.
type Client interface {
	// Start starts client, but does not wait for the command to exit.
	Start(ctx context.Context, verbose bool) error

	// AfterStart is called after the client has been fully started.
	AfterStart(ctx context.Context) error

	// HttpAddr returns the address where the client is serving JSON-RPC.
	HttpAddr() string

	// Close closes the client.
	Close() error
}

// gethClient is a wrapper around a go-ethereum instance on a separate thread.
type gethClient struct {
	cmd     *exec.Cmd
	path    string
	workdir string
	jwt     []byte
	fcu     rpcRequest
}

type rpcRequest struct {
	Method string
	Params []any
}

// newGethClient instantiates a new GethClient.
//
// The client's data directory is set to a temporary location and it
// initializes with the genesis and the provided blocks.
func newGethClient(ctx context.Context, geth string, chaindir string, verbose bool) (*gethClient, error) {
	// Load ForkchoiceUpdated from test chain.
	var fcuRequest rpcRequest
	fcuFile := filepath.Join(chaindir, "headfcu.json")
	if err := common.LoadJSON(fcuFile, &fcuRequest); err != nil {
		return nil, err
	}

	// Initialize the data directory.
	tmp, err := os.MkdirTemp("", "rpctestgen-*")
	if err != nil {
		return nil, err
	}

	var (
		args     = ctx.Value(ARGS).(*Args)
		datadir  = fmt.Sprintf("--datadir=%s", tmp)
		gcmode   = "--gcmode=archive"
		loglevel = fmt.Sprintf("--verbosity=%d", args.logLevelInt)
		// Archive mode requires hash-based scheme.
		scheme = fmt.Sprintf("--state.scheme=%s", "hash")
	)

	// Run geth init.
	options := []string{datadir, gcmode, scheme, loglevel, "init", filepath.Join(chaindir, "genesis.json")}
	err = runCmd(ctx, geth, verbose, options...)
	if err != nil {
		return nil, err
	}

	// Run geth import.
	options = []string{datadir, gcmode, loglevel, "import", filepath.Join(chaindir, "chain.rlp")}
	err = runCmd(ctx, geth, verbose, options...)
	if err != nil {
		return nil, err
	}

	jwt := make([]byte, 32)
	rand.Read(jwt)
	if err := os.WriteFile(fmt.Sprintf("%s/jwt.hex", tmp), []byte(hexutil.Encode(jwt)), 0600); err != nil {
		return nil, err
	}

	g := &gethClient{path: geth, workdir: tmp, jwt: jwt, fcu: fcuRequest}
	return g, nil
}

// Start starts geth, but does not wait for the command to exit.
func (g *gethClient) Start(ctx context.Context, verbose bool) error {
	fmt.Println("starting client")

	var (
		args    = ctx.Value(ARGS).(*Args)
		options = []string{
			fmt.Sprintf("--datadir=%s", g.workdir),
			fmt.Sprintf("--verbosity=%d", args.logLevelInt),
			fmt.Sprintf("--port=%s", NETWORKPORT),
			"--gcmode=archive",
			"--nodiscover",
			"--http",
			"--http.api=admin,eth,debug,net,txpool",
			fmt.Sprintf("--http.addr=%s", HOST),
			fmt.Sprintf("--http.port=%s", PORT),
			fmt.Sprintf("--authrpc.port=%s", AUTHPORT),
			fmt.Sprintf("--authrpc.jwtsecret=%s", fmt.Sprintf("%s/jwt.hex", g.workdir)),
		}
	)
	g.cmd = exec.CommandContext(ctx, g.path, options...)
	if verbose {
		g.cmd.Stdout = os.Stdout
		g.cmd.Stderr = os.Stderr
	}
	if err := g.cmd.Start(); err != nil {
		return err
	}
	return nil
}

// AfterStart is called after the client has been fully started.
// We send a forkchoiceUpdatedV2 request to the engine to trigger a post-merge forkchoice.
func (g *gethClient) AfterStart(ctx context.Context) error {
	auth := node.NewJWTAuth(common.BytesToHash(g.jwt))
	endpoint := fmt.Sprintf("http://%s:%s", HOST, AUTHPORT)
	cl, err := rpc.DialOptions(ctx, endpoint, rpc.WithHTTPAuth(auth))
	if err != nil {
		return err
	}
	defer cl.Close()
	err = cl.CallContext(ctx, nil, g.fcu.Method, g.fcu.Params...)
	if err != nil {
		return fmt.Errorf("geth rejected forkchoiceUpdated: %v", err)
	}
	return nil
}

// HttpAddr returns the address where the client is servering its JSON-RPC.
func (g *gethClient) HttpAddr() string {
	return fmt.Sprintf("http://%s:%s", HOST, PORT)
}

// Close closes the client.
func (g *gethClient) Close() error {
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
