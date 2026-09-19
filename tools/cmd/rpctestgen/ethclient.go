package main

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"

	"github.com/ethereum/execution-apis/tools/iofile"
	"github.com/ethereum/go-ethereum/rpc"
)

type ethclientHandler struct {
	rpc       *rpc.Client
	testFile  *os.File
	testW     *iofile.Writer
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
	return &ethclientHandler{rpc: rpcClient, transport: rt}, nil
}

func (l *ethclientHandler) NewTest(filename string) error {
	if l.testFile != nil {
		if err := l.testFile.Close(); err != nil {
			return err
		}
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	l.testFile = f
	l.testW = iofile.NewWriter(f)
	l.transport.w = l.testW
	return nil
}

// WriteValidationScript appends a script section to the test file.
func (l *ethclientHandler) WriteValidationScript(text string) error {
	return l.testW.Script(text)
}

// WriteComment adds the given text as a comment to the test file.
func (l *ethclientHandler) WriteComment(text string) error {
	return l.testW.Comment(text)
}

func (l *ethclientHandler) Close() {
	if l.testFile != nil {
		l.testFile.Close()
	}
}

// loggingRoundTrip writes requests and responses to the test log.
type loggingRoundTrip struct {
	w     *iofile.Writer
	inner http.RoundTripper
}

func (rt *loggingRoundTrip) RoundTrip(req *http.Request) (*http.Response, error) {
	// Read and log the request body.
	reqBytes, err := io.ReadAll(req.Body)
	req.Body.Close()
	if err != nil {
		return nil, err
	}
	rt.w.Send(string(reqBytes))
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
	rt.w.Receive(string(respBytes))
	return &respCopy, nil
}
