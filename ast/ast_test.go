package ast_test

import (
	"testing"

	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/parser"
)

func parse(input string) (interface{}, error) {
	lex := lexer.NewLexer([]byte(input))
	p := parser.NewParser()
	return p.Parse(lex)
}

func BenchmarkParser(b *testing.B) {
	lex := lexer.NewLexer([]byte(
		"[[[[[k*]/3].]$].8]))",
	))
	p := parser.NewParser()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lex.Reset()
		p.Reset()
		_, err := p.Parse(lex)
		if err != nil {
			b.Fatal(err)
		}
	}
}
