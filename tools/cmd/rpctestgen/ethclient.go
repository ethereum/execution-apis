package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/ethereum/go-ethereum/rpc"
)

// errorCodeRe matches the `"code":<int>` field inside a JSON-RPC error object.
var errorCodeRe = regexp.MustCompile(`("error"\s*:\s*\{[^}]*?"code"\s*:\s*)-?\d+`)

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

// WriteComment adds the given text as a comment to the current log file.
func (l *ethclientHandler) WriteComment(text string) error {
	var b strings.Builder
	for line := range strings.Lines(text) {
		b.WriteString("//")
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			b.WriteString(" ")
			b.WriteString(line)
		}
		b.WriteString("\n")
	}
	_, err := l.logFile.WriteString(b.String())
	return err
}

func (l *ethclientHandler) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
}

// RewriteLastErrorCode substitutes the `code` digits in the last "<< " error
// response of the current log file, so fixtures assert the spec-mandated code
// regardless of what the reference client returned.
func (l *ethclientHandler) RewriteLastErrorCode(code int) error {
	if l.logFile == nil {
		return fmt.Errorf("no log file open")
	}
	filename := l.logFile.Name()
	if err := l.logFile.Close(); err != nil {
		return err
	}
	l.logFile = nil
	l.transport.w = nil

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	idx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.HasPrefix(lines[i], "<< ") {
			idx = i
			break
		}
	}
	if idx < 0 {
		return fmt.Errorf("no response line found in %s", filename)
	}
	replacement := fmt.Sprintf("${1}%d", code)
	rewritten := errorCodeRe.ReplaceAllString(lines[idx], replacement)
	if rewritten == lines[idx] {
		return fmt.Errorf("ExpectErrorCode set but no error.code field found in %s", filename)
	}
	lines[idx] = rewritten
	return os.WriteFile(filename, []byte(strings.Join(lines, "\n")), 0644)
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
