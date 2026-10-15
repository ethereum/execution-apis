/*
Package iofile implements the RPC test file format (.io file).

Test files define exchanges of JSON-RPC messages. The syntax is line-based:

  - Lines that begin with ">>" are messages sent to the server.
  - A line starting with "<<" is a received message expected by the client.
  - Comment lines are introduced by "//".

# Validation Script

Test files can optionally contain a 'validation script' section at the end. The script
section is introduced by a line containing "--" and nothing else. All remaining text in
the file is JavaScript code.

The test harness executes the validation script after message exchanges with the server.
If the script throws an exception, the test is considered to have failed.
Within the script, the `messages` variable contains an array of message objects.
Each element of `messages` is an object where

  - `messages[i].send` is set to the RPC request for send (>>) lines.
  - `messages[i].expected` is the expected message for receive (<<) lines.
  - `messages[i].response` is the message received from the server

Note the `console` module is available for use in validation scripts.
*/
package iofile

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Test represents a
type Test struct {
	Name     string
	Comment  string
	SpecOnly bool
	Messages []TestMessage
	Script   string // validation script

	scriptStartLine int
}

type TestMessage struct {
	Data json.RawMessage `json:"data"`
	// if true, the message is a Send (>>), otherwise it's a receive (<<)
	Send bool `json:"send"`
}

// Receives returns the data of all receive (<<) lines.
func (t *Test) Receives() (m []json.RawMessage) {
	for _, msg := range t.Messages {
		if !msg.Send {
			m = append(m, msg.Data)
		}
	}
	return m
}

// This is maximum length of a single test file line.
const maxLineLength = 50 * 1024 * 1024

// Load reads a test file.
func Load(name string, r io.Reader) (Test, error) {
	var (
		rdr        = bufio.NewReader(r)
		scan       = bufio.NewScanner(rdr)
		buf        = make([]byte, 0, 64*1024)
		inHeader   = true
		test       = Test{Name: name}
		lineNumber = 0
	)
	scan.Buffer(buf, maxLineLength)
	for scan.Scan() {
		lineNumber++
		line := strings.TrimSpace(scan.Text())
		switch {
		case len(line) == 0:
			continue

		case strings.HasPrefix(line, "//"):
			if !inHeader {
				continue // ignore comments after requests
			}
			text := strings.TrimPrefix(strings.TrimPrefix(line, "//"), " ")
			test.Comment += text + "\n"
			if strings.HasPrefix(text, "speconly:") {
				test.SpecOnly = true
			}

		case strings.HasPrefix(line, ">>") || strings.HasPrefix(line, "<<"):
			inHeader = false
			data := strings.TrimSpace(line[2:])
			if !json.Valid([]byte(data)) {
				return test, fmt.Errorf("invalid JSON in line %q", line)
			}
			test.Messages = append(test.Messages, TestMessage{
				Data: json.RawMessage(data),
				Send: strings.HasPrefix(line, ">>"),
			})

		case strings.HasPrefix(line, "--"):
			if strings.TrimSpace(line) != "--" {
				return test, fmt.Errorf("invalid script section divider on line %d (%q)", lineNumber, line)
			}
			// end of messages section, remainder is script
			test.Script = slurpLines(scan)
			test.scriptStartLine = lineNumber

		default:
			return test, fmt.Errorf("invalid test line: %q", line)
		}
	}
	return test, scan.Err()
}

// slurpLines reads all lines of text up to EOF from the given scanner.
func slurpLines(scan *bufio.Scanner) string {
	var text strings.Builder
	for scan.Scan() {
		text.Write(scan.Bytes())
		text.WriteByte('\n')
	}
	return text.String()
}

type TestLogger interface {
	Logf(format string, args ...any)
}

// LoadDirectory walks the given directory looking for *.io files to load.
func LoadDirectory(logger TestLogger, root string, re *regexp.Regexp) ([]Test, error) {
	var tests []Test
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logger.Logf("unable to walk path: %s", err)
			return err
		}
		if info.IsDir() {
			return nil
		}
		if fname := info.Name(); !strings.HasSuffix(fname, ".io") {
			return nil
		}
		pathname := strings.TrimSuffix(strings.TrimPrefix(path, root+"/"), ".io")
		if !re.MatchString(pathname) {
			fmt.Println("skip", pathname)
			return nil // skip
		}
		fd, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fd.Close()
		test, err := Load(pathname, fd)
		if err != nil {
			return fmt.Errorf("invalid test %s: %v", info.Name(), err)
		}
		tests = append(tests, test)
		return nil
	})
	return tests, err
}
