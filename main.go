package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

type Args struct {
	Verbose bool `arg:"-v,--verbose"`
}

func main() {
	var args Args
	arg.MustParse(&args)
	gspec, blocks := genSimpleChain()

	ctx := context.Background()
	geth, err := NewGethClient(ctx, "/usr/local/bin/geth", gspec, blocks, args.Verbose)
	if err != nil {
		panic(err)
	}
	defer geth.Close()

	geth.Start(ctx, args.Verbose)
	time.Sleep(2 * time.Second)
	f, err := os.Create("out.log")
	if err != nil {
		panic(err)
	}
	client := &http.Client{
		Transport: &loggingRoundTrip{
			w:     f,
			inner: http.DefaultTransport,
		},
	}
	rpcClient, err := rpc.DialHTTPWithClient("http://127.0.0.1:8545/", client)
	if err != nil {
		fmt.Println(err)
		return
	}
	eth := ethclient.NewClient(rpcClient)
	num, err := eth.BlockNumber(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(num)

	time.Sleep(3 * time.Second)
	geth.cmd.Process.Kill()
}

// loggingRoundTrip writes requests and responses to the test log.
type loggingRoundTrip struct {
	w     io.Writer
	inner http.RoundTripper
}

func (rt *loggingRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read and log the request body.
	reqBytes, err := ioutil.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(rt.w, ">>  %s\n", bytes.TrimSpace(reqBytes))
	reqCopy := *req
	reqCopy.Body = ioutil.NopCloser(bytes.NewReader(reqBytes))

	// Do the round trip.
	resp, err := rt.inner.RoundTrip(&reqCopy)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and log the response bytes.
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respCopy := *resp
	respCopy.Body = ioutil.NopCloser(bytes.NewReader(respBytes))
	fmt.Fprintf(rt.w, "<<  %s\n", bytes.TrimSpace(respBytes))
	return &respCopy, nil
}
