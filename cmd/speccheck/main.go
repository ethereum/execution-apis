package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/alexflint/go-arg"
	"github.com/santhosh-tekuri/jsonschema/v5"
)

type Args struct {
	SpecPath   string `arg:"--spec" help:"path to client binary" default:"openrpc.json"`
	TestsRoot  string `arg:"--tests" help:"path to tests directory" default:"tests"`
	TestsRegex string `arg:"--regexp" help:"regular expression to match tests to check" deafult:".*"`
	Verbose    bool   `arg:"-v,--verbose" help:"verbosity level of rpctestgen"`
}

type MethodSchema struct {
	Params [][]byte
	Result []byte
}

type RoundTrip struct {
	Method   string
	Params   []string
	Response string
}

func main() {
	var args Args
	arg.MustParse(&args)
	if err := checkSpec(&args); err != nil {
		exit(err)
	}
}

func checkSpec(args *Args) error {
	re, err := regexp.Compile(args.TestsRegex)
	if err != nil {
		return err
	}
	spec, err := os.ReadFile(args.SpecPath)
	if err != nil {
		return err
	}
	methods, err := parseMethods(spec)
	if err != nil {
		return err
	}
	rts, err := parseRoundTrips(args.TestsRoot, re)
	if err != nil {
		return err
	}

	for _, rt := range rts {
		method, ok := methods[rt.Method]
		if !ok {
			return fmt.Errorf("undefined method: %s", rt.Method)
		}
		if len(method.Params) < len(rt.Params) {
			return fmt.Errorf("too many parameters")
		}
		for i, param := range method.Params {
			schema, err := jsonschema.CompileString(fmt.Sprintf("%s.param[%d]", rt.Method, i), string(param))
			if err != nil {
				return err
			}
			fmt.Println(rt.Params[i])
			var val interface{}
			if err := json.Unmarshal([]byte(rt.Params[i]), &val); err != nil {
				return err
			}
			if err := schema.Validate(val); err != nil {
				return err
			}
		}
		schema, err := jsonschema.CompileString(fmt.Sprintf("%s.result", rt.Method), string(method.Result))
		if err != nil {
			return err
		}
		fmt.Println(rt.Response)
		var val interface{}
		if err := json.Unmarshal([]byte(rt.Response), &val); err != nil {
			return err
		}
		if err := schema.Validate(val); err != nil {
			var schema interface{}
			json.Unmarshal(method.Result, &schema)
			buf, _ := json.MarshalIndent(schema, "", "  ")
			fmt.Println(string(buf))
			fmt.Println(string(method.Result))
			fmt.Println(rt.Response)
			return err
		}
	}

	fmt.Println("all pass.")
	return nil
}

func parseRoundTrips(root string, re *regexp.Regexp) ([]*RoundTrip, error) {
	rts := make([]*RoundTrip, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			fmt.Printf("unable to walk path: %s\n", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		if fname := info.Name(); !strings.HasSuffix(fname, ".io") {
			return nil
		}
		pathname := strings.TrimSuffix(strings.TrimPrefix(path, root), ".io")
		if !re.MatchString(pathname) {
			fmt.Println("skip", pathname)
			return nil // skip
		}
		data, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		fmt.Println("loading", path)
		test, err := parseTest(data)
		if err != nil {
			return err
		}
		rts = append(rts, test...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return rts, nil
}

func exit(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}
