package balafon

import (
	"bytes"

	"github.com/mgnsk/balafon/internal/ast"
	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/parser"
)

// Format a balafon script.
func Format(input []byte) ([]byte, error) {
	scanner := lexer.NewLexer(input)
	p := parser.NewParser()

	res, err := p.Parse(scanner)
	if err != nil {
		return nil, err
	}

	declList, ok := res.(ast.NodeList)
	if !ok {
		panic("invalid input, expected ast.NodeList")
	}

	var buf bytes.Buffer
	if _, err := declList.WriteTo(&buf); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
