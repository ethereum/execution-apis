//go:build tools
// +build tools

package main

// This adds a dependency on the geth command so that we can build it with go build
// within the rpctestgen module.
import _ "github.com/ethereum/go-ethereum/cmd/geth"
