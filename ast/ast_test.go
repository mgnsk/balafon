package ast_test

import (
	"testing"

	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/parser"
	. "github.com/onsi/gomega"
)

func parse(input string) (interface{}, error) {
	lex := lexer.NewLexer([]byte(input))
	p := parser.NewParser()
	return p.Parse(lex)
}

func TestSyntaxNotAmbigous(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign t 0
:assign e 1
:assign m 2
:assign p 3
:assign o 4
:bar bar
	:timesig 5 4
	tempo
:end
:play bar
`

	_, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestBarNotAmbigous(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign b 0
:assign a 1
:assign r 2
:bar bar
	bar
:end
:play bar
`

	_, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestPlayNotAmbigous(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign p 0
:assign l 1
:assign a 2
:assign y 3
:bar play
	play
:end
:play play
`

	_, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())
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
