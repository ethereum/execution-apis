package main

import (
	"flag"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/execution-apis/tools/internal/specgen"
)

var methodFilesFlag = []string{}
var schemaFilesFlag = []string{}
var outputFile = ""
var dereferencing bool

func init() {
	flag.Func("methods", "path to method files (glob syntax, repeatable)", func(s string) error {
		methodFilesFlag = append(methodFilesFlag, s)
		return nil
	})
	flag.Func("schemas", "path to schema files (glob syntax, repeatable)", func(s string) error {
		schemaFilesFlag = append(schemaFilesFlag, s)
		return nil
	})
	flag.StringVar(&outputFile, "output", "", "output file")
	flag.StringVar(&outputFile, "o", "", "output file")
	flag.BoolVar(&dereferencing, "deref", false, "Enable dereferencing of spec")
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	var methodFiles []string
	for _, file := range methodFilesFlag {
		info, err := os.Stat(file)
		if err != nil {
			log.Fatal("can't access method file:", err)
		}
		if info.IsDir() {
			if err := filepath.WalkDir(file, addFilesWithExt(&methodFiles, "yaml")); err != nil {
				log.Fatal(err)
			}
		} else {
			methodFiles = append(methodFiles, file)
		}
	}
	if len(methodFiles) == 0 {
		log.Fatalf("must provide at least one method file")
	}

	var schemaFiles []string
	for _, file := range schemaFilesFlag {
		info, err := os.Stat(file)
		if err != nil {
			log.Fatal("can't access schema file:", err)
		}
		if info.IsDir() {
			if err := filepath.WalkDir(file, addFilesWithExt(&schemaFiles, "yaml")); err != nil {
				log.Fatal(err)
			}
		} else {
			schemaFiles = append(schemaFiles, file)
		}
	}
	if len(schemaFilesFlag) == 0 {
		log.Fatalf("must provide at least one schema file")
	}

	sg := specgen.New()

	// Read all the files
	for _, file := range methodFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatal("can't read method file:", err)
		}
		if err := sg.AddMethods(content); err != nil {
			log.Fatalf("error in %s: %v", file, err)
		}
		log.Println("added methods from", file)
	}
	for _, file := range schemaFiles {
		content, err := os.ReadFile(file)
		if err != nil {
			log.Fatal("can't read schema file:", err)
			os.Exit(1)
		}
		if err := sg.AddSchemas(content); err != nil {
			log.Fatalf("error in %s: %v", file, err)
		}
		log.Println("added schemas from", file)
	}

	if outputFile == "" {
		log.Printf("no output file specified, just validating spec")
		err := sg.Validate()
		if err != nil {
			log.Println("spec is invalid:", err)
			os.Exit(1)
		} else {
			log.Println("spec is valid")
			os.Exit(0)
		}
	} else {
		outputBytes, err := sg.JSON()
		if err != nil {
			log.Fatal(err)
		}
		if err := os.WriteFile(outputFile, outputBytes, 0644); err != nil {
			log.Fatal("spec write failed:", err)
		}
		log.Println("wrote spec to", outputFile)
	}
}

func addFilesWithExt(list *[]string, ext string) fs.WalkDirFunc {
	ext = "." + ext
	return func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.Type().IsRegular() && strings.HasSuffix(path, ext) {
			*list = append(*list, path)
		}
		return nil
	}
}
