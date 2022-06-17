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

type Client interface {
	Start(ctx context.Context, verbose bool) error
	HttpAddr() string
	Close() error
}

type GethClient struct {
	cmd     *exec.Cmd
	path    string
	workdir string
	blocks  []*types.Block
	genesis *core.Genesis
}

func NewGethClient(ctx context.Context, path string, genesis *core.Genesis, blocks []*types.Block, verbose bool) (*GethClient, error) {
	tmp, err := ioutil.TempDir("", "geth-")
	if err != nil {
		return nil, err
	}
	writeChain(tmp, genesis, blocks)
	datadir := fmt.Sprintf("--datadir=%s", tmp)
	err = runCmd(ctx, path, verbose, datadir, "--fakepow", "init", fmt.Sprintf("%s/genesis.json", tmp))
	if err != nil {
		return nil, err
	}
	err = runCmd(ctx, path, verbose, datadir, "--fakepow", "import", fmt.Sprintf("%s/chain.rlp", tmp))
	if err != nil {
		return nil, err
	}
	return &GethClient{path: path, genesis: genesis, blocks: blocks, workdir: tmp}, nil
}

func (g *GethClient) Start(ctx context.Context, verbose bool) error {
	g.cmd = exec.CommandContext(ctx, g.path, fmt.Sprintf("--datadir=%s", g.workdir), "--http", "--nodiscover", "--fakepow")
	if verbose {
		g.cmd.Stdout = os.Stdout
		g.cmd.Stderr = os.Stderr
	}
	if err := g.cmd.Start(); err != nil {
		return err
	}
	return nil
}

func (g *GethClient) HttpAddr() string {
	return fmt.Sprintf("http://%s:%s", HOST, PORT)
}

func (g *GethClient) Close() error {
	g.cmd.Process.Kill()
	g.cmd.Wait()
	return os.RemoveAll(g.workdir)
}

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
