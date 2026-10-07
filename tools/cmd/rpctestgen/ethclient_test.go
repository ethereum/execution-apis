package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/rpc"
)

func TestAuthenticatedFixtureRecording(t *testing.T) {
	const token = "test-auth-header-not-for-fixtures"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.Header.Get("Authorization") != "Bearer "+token {
			t.Error("missing authentication")
		}
		var request struct {
			ID json.RawMessage `json:"id"`
		}
		if err := json.NewDecoder(req.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"jsonrpc": "2.0", "id": request.ID, "result": []string{}})
	}))
	defer server.Close()
	handler, err := newEthclientHandler(server.URL, rpc.WithHTTPAuth(func(h http.Header) error {
		h.Set("Authorization", "Bearer "+token)
		return nil
	}))
	if err != nil {
		t.Fatal(err)
	}
	defer handler.rpc.Close()
	defer handler.Close()
	file := filepath.Join(t.TempDir(), "engine.io")
	if err := handler.RotateLog(file); err != nil {
		t.Fatal(err)
	}
	var result []string
	if err := handler.rpc.CallContext(context.Background(), &result, "engine_exchangeCapabilities", []string{}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), token) || !strings.Contains(string(data), ">> ") || !strings.Contains(string(data), "<< ") {
		t.Fatalf("fixture should contain only RPC bodies: %s", data)
	}
}
