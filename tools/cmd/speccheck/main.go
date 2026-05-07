package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/alexflint/go-arg"
)

type Args struct {
	SpecPath   string `arg:"--spec" help:"path to client binary" default:"openrpc.json"`
	TestsRoot  string `arg:"--tests" help:"path to tests directory" default:"tests"`
	TestsRegex string `arg:"--regexp" help:"regular expression to match tests to check" deafult:".*"`
	Verbose    bool   `arg:"-v,--verbose" help:"verbosity level of rpctestgen"`
}

func main() {
	var args Args
	arg.MustParse(&args)
	if err := run(&args); err != nil {
		exit(err)
	}
}

func run(args *Args) error {
	re, err := regexp.Compile(args.TestsRegex)
	if err != nil {
		return err
	}

	// Read all method schemas (params+result) from the OpenRPC spec.
	methods, err := parseSpec(args.SpecPath)
	if err != nil {
		return err
	}

	// Read all tests and parse out roundtrip HTTP exchanges so they can be validated.
	rts, err := readRtts(args.TestsRoot, re)
	if err != nil {
		return err
	}

	return checkSpec(methods, rts, re)
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
