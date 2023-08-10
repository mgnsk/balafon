package balafon

import (
	"io"

	"github.com/mgnsk/balafon/internal/ast"
	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/parser"
)

// ToXML2 converts a balafon script to MusicXML.
func ToXML2(w io.Writer, input []byte) error {
	p := parser.NewParser()

	res, err := p.Parse(lexer.NewLexer(input))
	if err != nil {
		return err
	}

	declList, ok := res.(ast.NodeList)
	if !ok {
		panic("invalid input, expected ast.NodeList")
	}

	bp := New()
	if _, err := bp.parse(declList); err != nil {
		return err
	}

	return nil
}
