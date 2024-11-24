package balafon

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"github.com/mgnsk/balafon/internal/ast"
	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/parser"
)

var lf = []byte("\n")

// FormatFile formats a file.
func FormatFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	b, err := io.ReadAll(f)
	if err != nil {
		return err
	}

	result, err := Format(b)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, result, stat.Mode())
}

// Format a balafon script.
func Format(input []byte) ([]byte, error) {
	scanner := bufio.NewScanner(bytes.NewReader(input))
	p := parser.NewParser()

	var (
		isBar       bool
		barBuffer   bytes.Buffer
		output      bytes.Buffer
		isEmptyLine bool
	)

	for scanner.Scan() {
		line := scanner.Bytes()
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			if isEmptyLine {
				continue
			}
			isEmptyLine = true
		} else {
			isEmptyLine = false
		}

		if isBar && bytes.HasPrefix(line, []byte(":end")) {
			barBuffer.Write(line)
			barBuffer.Write(lf)
			isBar = false

			barBuffer.WriteTo(&output)
			barBuffer.Reset()
		} else if pref := []byte(":bar"); bytes.HasPrefix(line, pref) {
			isBar = true

			barName := bytes.TrimPrefix(line, pref)
			barName = bytes.TrimSpace(barName)

			barBuffer.Write(pref)
			barBuffer.WriteString(" ")
			barBuffer.Write(barName)
			barBuffer.Write(lf)
		} else if pref := []byte(":play"); bytes.HasPrefix(line, pref) {
			barName := bytes.TrimPrefix(line, pref)
			barName = bytes.TrimSpace(barName)

			output.Write(pref)
			output.WriteString(" ")
			output.Write(barName)
			output.Write(lf)
		} else if isBar {
			if len(line) > 0 {
				barBuffer.WriteString("\t")
			}

			if bytes.HasPrefix(line, []byte(":")) {
				// Parse the command line and print.
				node, err := p.Parse(lexer.NewLexer(line))
				if err != nil {
					return nil, err
				}
				node.(ast.NodeList).WriteTo(&barBuffer)
			} else {
				// Print note list raw.
				barBuffer.Write(line)
			}

			barBuffer.Write(lf)
		} else {
			if bytes.HasPrefix(line, []byte(":")) {
				// Parse the command line and print.
				node, err := p.Parse(lexer.NewLexer(line))
				if err != nil {
					return nil, err
				}
				node.(ast.NodeList).WriteTo(&output)
			} else {
				// Print note list raw.
				output.Write(line)
			}

			output.Write(lf)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	out := bytes.TrimSpace(output.Bytes())
	out = append(out, lf...)

	return out, nil
}
