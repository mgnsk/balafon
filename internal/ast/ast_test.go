package ast_test

import (
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
)

// parse the input with an appended newline.
func parse(input string) (interface{}, error) {
	lex := lexer.NewLexer([]byte(input + "\n"))
	p := parser.NewParser()
	return p.Parse(lex)
}
