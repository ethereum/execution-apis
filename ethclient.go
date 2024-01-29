package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/ethereum/go-ethereum/rpc"
)

type ethclientHandler struct {
	rpc       *rpc.Client
	logFile   *os.File
	transport *loggingRoundTrip
}

func newEthclientHandler(addr string) (*ethclientHandler, error) {
	rt := &loggingRoundTrip{
		inner: http.DefaultTransport,
	}
	httpClient := rpc.WithHTTPClient(&http.Client{Transport: rt})
	ctx := context.Background()
	rpcClient, err := rpc.DialOptions(ctx, addr, httpClient)
	if err != nil {
		return nil, err
	}
	return &ethclientHandler{
		rpc:       rpcClient,
		logFile:   nil,
		transport: rt,
	}, nil
}

func (l *ethclientHandler) RotateLog(filename string) error {
	if l.logFile != nil {
		if err := l.logFile.Close(); err != nil {
			return err
		}
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	l.logFile = f
	l.transport.w = f
	return nil
}

func (l *ethclientHandler) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

// loggingRoundTrip writes requests and responses to the test log.
type loggingRoundTrip struct {
	w     io.Writer
	inner http.RoundTripper
}

func (rt *loggingRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read and log the request body.
	reqBytes, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	fmt.Fprintf(rt.w, ">> %s\n", bytes.TrimSpace(reqBytes))
	reqCopy := *req
	reqCopy.Body = io.NopCloser(bytes.NewReader(reqBytes))

	// Do the round trip.
	resp, err := rt.inner.RoundTrip(&reqCopy)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read and log the response bytes.
	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	respCopy := *resp
	respCopy.Body = io.NopCloser(bytes.NewReader(respBytes))
	fmt.Fprintf(rt.w, "<< %s\n", bytes.TrimSpace(respBytes))
	return &respCopy, nil
}
