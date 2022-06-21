package main

import (
	"context"
	"fmt"
	"os"

	"github.com/alexflint/go-arg"
)

const (
	HOST        string = "127.0.0.1"
	PORT        string = "13375"
	NETWORKPORT string = "13376"
)

type Args struct {
	ClientType string `arg:"--client" help:"client type" default:"geth"`
	ClientBin  string `arg:"--bin" help:"path to client binary" default:"geth"`
	OutDir     string `arg:"--out" help:"directory where test fixtures will be written" default:"tests"`
	Ethash     bool   `args:"--ethash" help:"seal blocks using proof-of-work"`
	EthashDir  string `args:"--ethashdir" help:"directory to store ethash dag (empty for in-memory only)"`
	Verbose    bool   `arg:"-v,--verbose" help:"verbosity level of rpctestgen"`
	LogLevel   string `arg:"--loglevel" help:"log level of client" default:"info"`

	logLevelInt int
}

type ArgsKey struct{}

var ARGS = ArgsKey{}

func main() {
	var args Args
	arg.MustParse(&args)

	lvl, err := loglevelToInt(args.LogLevel)
	if err != nil {
		exit(err)
	}
	args.logLevelInt = lvl

	ctx := context.Background()
	ctx = context.WithValue(ctx, ARGS, &args)

	if err := runGenerator(ctx); err != nil {
		exit(err)
	}
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func loglevelToInt(lvl string) (int, error) {
	switch lvl {
	case "err":
		return 1, nil
	case "warn":
		return 2, nil
	case "info":
		return 3, nil
	case "debug":
		return 4, nil
	case "trace":
		return 5, nil
	default:
		return 0, fmt.Errorf("unknown log level: %s", lvl)
	}
}
