package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/ethereum/execution-apis/tools/iofile"
	"github.com/ethereum/execution-apis/tools/spec"
)

type Args struct {
	SpecPath   string `arg:"--spec" help:"path to client binary" default:"refs-openrpc.json"`
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
	s, err := spec.Load(args.SpecPath)
	if err != nil {
		return err
	}
	// Read all tests and parse out roundtrip HTTP exchanges so they can be validated.
	tests, err := iofile.LoadDirectory(iofile.StdoutLogger, args.TestsRoot, re)
	if err != nil {
		return err
	}

	return checkTests(s, tests, re)
}

// checkSpec reads the schemas from the spec and test files, then validates
// them against each other.
func checkTests(s *spec.Spec, tests []iofile.Test, re *regexp.Regexp) error {
	for _, test := range tests {
		fmt.Println(test.Name)
		// Check all calls and responses against the spec.
		checker := spec.NewChecker(s)
		checker.SetCheckArgs(!strings.Contains(test.Name, "invalid"))
		for i, msg := range test.Messages {
			var err error
			if msg.Send {
				err = checker.Request(msg.Data)
			} else {
				err = checker.Response(msg.Data)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "message[%d]: %s\n", i, indent(msg.Data))
				return fmt.Errorf("error: %v", err)
			}
		}
		if dropped := checker.UnansweredCalls(); len(dropped) > 0 {
			return fmt.Errorf("%s has unanswered calls: %v", test.Name, dropped)
		}

		// Run validation script. Note that the 'responses' seen by the script
		// are equal to the test's declared responses.
		config := iofile.ScriptConfig{OpenRPC: s}
		responses := test.Receives()
		if err := test.RunScript(config, responses); err != nil {
			return fmt.Errorf("%s: %v", test.Name, err)
		}
	}
	fmt.Println("all passing.")
	return nil
}

func indent(msg []byte) []byte {
	var buf bytes.Buffer
	err := json.Indent(&buf, msg, "", "  ")
	if err != nil {
		return msg
	}
	return buf.Bytes()
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
