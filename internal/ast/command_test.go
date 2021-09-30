package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
)

func TestValidCommands(t *testing.T) {
	type (
		match    []GomegaMatcher
		testcase struct {
			input string
			match match
		}
	)

	for _, tc := range []testcase{
		{
			"bar \"Chorus0\"\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring(`bar "Chorus0"`),
			},
		},
		{
			"bar \"Chorus1\"\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring(`bar "Chorus1"`),
			},
		},
		{
			"end\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("end"),
			},
		},
		{
			"play \"chorus\"\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring(`play "chorus"`),
			},
		},
		{
			"play \"Chorus0\"\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring(`play "Chorus0"`),
			},
		},
		{
			"play \"Chorus1\"\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring(`play "Chorus1"`),
			},
		},
		{
			"tempo 120\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("tempo 120"),
			},
		},
		{
			"channel 0\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("channel 0"),
			},
		},
		{
			"velocity 50\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("velocity 50"),
			},
		},
		{
			"program 0\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("program 0"),
			},
		},
		{
			"control 0 1\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("control 0 1"),
			},
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			lex := lexer.NewLexer([]byte(tc.input))
			p := parser.NewParser()

			res, err := p.Parse(lex)
			g.Expect(err).NotTo(HaveOccurred())

			for _, match := range tc.match {
				g.Expect(res).To(match)
			}
		})
	}
}

func TestTempoRange(t *testing.T) {
	g := NewGomegaWithT(t)

	for _, input := range []string{"tempo 0\n", "tempo 65536\n"} {
		lex := lexer.NewLexer([]byte(input))
		p := parser.NewParser()

		_, err := p.Parse(lex)
		g.Expect(err).To(HaveOccurred())
	}
}

func TestChannelRange(t *testing.T) {
	g := NewGomegaWithT(t)

	lex := lexer.NewLexer([]byte("channel 16\n"))
	p := parser.NewParser()

	_, err := p.Parse(lex)
	g.Expect(err).To(HaveOccurred())
}

func TestVelocityRange(t *testing.T) {
	g := NewGomegaWithT(t)

	lex := lexer.NewLexer([]byte("velocity 128\n"))
	p := parser.NewParser()

	_, err := p.Parse(lex)
	g.Expect(err).To(HaveOccurred())
}

func TestProgramRange(t *testing.T) {
	g := NewGomegaWithT(t)

	lex := lexer.NewLexer([]byte("program 128\n"))
	p := parser.NewParser()

	_, err := p.Parse(lex)
	g.Expect(err).To(HaveOccurred())
}

func TestControlRange(t *testing.T) {
	g := NewGomegaWithT(t)

	for _, input := range []string{"control 0 128\n", "control 128 0\n"} {
		lex := lexer.NewLexer([]byte(input))
		p := parser.NewParser()

		_, err := p.Parse(lex)
		g.Expect(err).To(HaveOccurred())
	}
}
