package main

import (
	"fmt"
	"os"

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
	if err := checkSpec(&args); err != nil {
		exit(err)
	}
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
