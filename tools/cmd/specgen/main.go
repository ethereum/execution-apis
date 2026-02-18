package main

import (
	"encoding/json"
	"flag"
	"log/slog"
	"os"
	"path/filepath"

	generator "github.com/ethereum/execution-apis/tools/internal/specgen"
)

var methodFilesFlag = []string{}
var schemaFilesFlag = []string{}
var verbose = false
var outputFile = ""

func init() {
	flag.Func("methods", "path to method files (glob syntax, repeatable)", func(s string) error {
		methodFilesFlag = append(methodFilesFlag, s)
		return nil
	})
	flag.Func("schemas", "path to schema files (glob syntax, repeatable)", func(s string) error {
		schemaFilesFlag = append(schemaFilesFlag, s)
		return nil
	})
	flag.BoolVar(&verbose, "verbose", false, "verbose output")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.StringVar(&outputFile, "output", "", "output file")
	flag.StringVar(&outputFile, "o", "", "output file")
}

func main() {
	flag.Parse()

	var logger *slog.Logger
	if verbose {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	// Must have at least one method file and one schema file.
	if len(methodFilesFlag) == 0 {
		logger.Error("must provide at least one method file")
		os.Exit(1)
	}
	if len(schemaFilesFlag) == 0 {
		logger.Error("must provide at least one schema file")
		os.Exit(1)
	}

	methodFiles := []string{}
	for _, file := range methodFilesFlag {
		if _, err := os.Stat(file); err == nil {
			logger.Info("ignoring directory", "directory", file)
			continue
		}
		files, err := filepath.Glob(file)
		if err != nil {
			logger.Error("failed to glob method files", "error", err)
			os.Exit(1)
		}
		methodFiles = append(methodFiles, files...)
	}

	schemaFiles := []string{}
	for _, file := range schemaFilesFlag {
		if _, err := os.Stat(file); err == nil {
			logger.Info("ignoring directory", "directory", file)
			continue
		}
		files, err := filepath.Glob(file)
		if err != nil {
			logger.Error("failed to glob schema files", "error", err)
			os.Exit(1)
		}
		schemaFiles = append(schemaFiles, files...)
	}

	logger.Info("method files", "count", len(methodFiles))
	logger.Debug("method file paths", "paths", methodFiles)
	logger.Info("schema files", "count", len(schemaFiles))
	logger.Debug("schema file paths", "paths", schemaFiles)

	sg := generator.NewSpecgen()

	// Read all the files
	for _, filename := range methodFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			logger.Error("failed to read method file", "error", err)
			os.Exit(1)
		}
		err = sg.AddMethods(content)
		if err != nil {
			logger.Error("failed to add methods", "error", err)
			os.Exit(1)
		}
		logger.Debug("added methods", "filename", filename)
	}
	for _, filename := range schemaFiles {
		content, err := os.ReadFile(filename)
		if err != nil {
			logger.Error("failed to read schema file", "error", err)
			os.Exit(1)
		}
		err = sg.AddSchemas(content)
		if err != nil {
			logger.Error("failed to add schemas", "error", err)
			os.Exit(1)
		}
		logger.Debug("added schemas", "filename", filename)
	}

	if outputFile == "" {
		logger.Info("no output file specified, just validating spec")
		err := sg.Validate()
		if err != nil {
			logger.Error("failed to validate specgen", "error", err)
			os.Exit(1)
		}
		logger.Info("spec is valid against the OpenRPC meta-schema")
		os.Exit(0)
	}

	logger.Debug("attempting to write to output file", "file", outputFile)
	outputBytes, err := json.MarshalIndent(sg, "", "\t")
	if err != nil {
		logger.Error("failed to marshal spec to json", "error", err)
		os.Exit(1)
	}
	err = os.WriteFile(outputFile, outputBytes, 0644)
	if err != nil {
		logger.Error("failed to write spec to output file", "error", err)
		os.Exit(1)
	}
	logger.Info("wrote spec to output file", "file", outputFile)
}
