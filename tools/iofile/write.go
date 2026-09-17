package iofile

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

// Writer creates a test file.
type Writer struct {
	output io.StringWriter
	test   Test
	script bool // true after script was written
	line   int
}

func NewWriter(output io.StringWriter) *Writer {
	return &Writer{output: output}
}

// Test returns the test that was written.
func (w *Writer) Test() Test {
	return w.test
}

// Send writes a send operation (>>).
func (w *Writer) Send(text string) error {
	return w.writeMessage(true, text)
}

// Receive writes a receive operation (<<).
func (w *Writer) Receive(text string) error {
	return w.writeMessage(false, text)
}

func (w *Writer) writeMessage(send bool, text string) error {
	if w.script {
		return errors.New("can't write message after script section")
	}

	text = strings.TrimSpace(text)
	if strings.Contains(text, "\n") {
		return errors.New("invalid JSON message (contains newlines)")
	}
	data := []byte(text)
	if !json.Valid(data) {
		return errors.New("invalid JSON message")
	}
	w.line++
	w.test.Messages = append(w.test.Messages, TestMessage{Data: data, Send: send})

	op := "<<"
	if send {
		op = ">>"
	}
	_, err := w.output.WriteString(op + " " + text + "\n")
	return err
}

// Comment adds the given text as comment lines.
func (w *Writer) Comment(text string) error {
	if w.script {
		return errors.New("can't write comment after script section")
	}
	var b strings.Builder
	for line := range strings.Lines(text) {
		b.WriteString("//")
		line = strings.TrimSpace(line)
		w.test.Comment += line + "\n"
		if strings.HasPrefix(line, "speconly:") {
			w.test.SpecOnly = true
		}
		if len(line) > 0 {
			b.WriteString(" ")
			b.WriteString(line)
		}
		b.WriteString("\n")
		w.line++
	}
	_, err := w.output.WriteString(b.String())
	return err
}

// Script writes a validation script section.
// No other operations can be added after this section.
func (w *Writer) Script(text string) error {
	w.script = true
	if _, err := w.output.WriteString("--\n"); err != nil {
		return err
	}
	w.line++
	w.test.Script = unindent(text)
	w.test.scriptStartLine = w.line
	_, err := w.output.WriteString(w.test.Script)
	return err
}

// unindent removes whitespace up to the first column where text starts.
func unindent(s string) string {
	s = strings.ReplaceAll(s, "\t", "    ")
	lines := strings.Split(s, "\n")
	// remove initial blank lines
	for len(lines) > 0 && strings.TrimSpace(lines[0]) == "" {
		lines = lines[1:]
	}
	// find indentation column
	minIndent := -1
	for _, l := range lines {
		body := strings.TrimLeft(l, " \t")
		if body == "" {
			continue // blank lines don't constrain the indent
		}
		if n := len(l) - len(body); minIndent < 0 || n < minIndent {
			minIndent = n
		}
	}

	if minIndent <= 0 {
		return s
	}
	for i, l := range lines {
		if len(l) < minIndent {
			lines[i] = strings.TrimLeft(l, " \t") // short/whitespace-only line
		} else {
			lines[i] = l[minIndent:]
		}
	}
	return strings.Join(lines, "\n")
}
