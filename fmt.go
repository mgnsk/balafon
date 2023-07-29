package balafon

import (
	"bufio"
	"bytes"
	"io"
	"os"

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

	if err := f.Close(); err != nil {
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

			_ = p
			// TODO
			// node, err := p.Parse(lexer.NewLexer(barBuffer.Bytes()))
			// barBuffer.Reset()
			// if err != nil {
			// 	var perr *ParseError
			// 	if errors.As(err, &perr) && perr.ErrorToken.Type == token.EOF {
			// 		continue
			// 	}
			// 	return nil, err
			// }

			// nodeList, ok := node.(ast.NodeList)
			// if !ok {
			// 	panic(fmt.Sprintf("expected %T, got %T", ast.NodeList{}, node))
			// }

			// nodeList.WriteTo(&output)
		} else if bytes.HasPrefix(line, []byte(":bar")) {
			isBar = true
			barBuffer.Write(line)
			barBuffer.Write(lf)
		} else if isBar {
			barBuffer.WriteString("\t")
			barBuffer.Write(line)
			barBuffer.Write(lf)
		} else {
			output.Write(line)
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
