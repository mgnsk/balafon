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

func TestInvalidArguments(t *testing.T) {
	for _, input := range []string{
		"tempo",
		"tempo 1 1",
		"tempo \"string\"",
		"channel",
		"channel 0 0",
		"channel \"string\" \"string\"",
		"velocity",
		"velocity 0 0",
		"velocity \"string\" \"string\"",
		"program",
		"program 0 0",
		"program \"string\" \"string\"",
		"control",
		"control 0",
		"control \"string\"",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)
			lex := lexer.NewLexer([]byte(input))
			p := parser.NewParser()

			_, err := p.Parse(lex)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("requires"))
		})
	}
}

func TestInvalidArgumentRange(t *testing.T) {
	for _, input := range []string{
		"tempo 0",
		"tempo 65536",
		"channel 16",
		"velocity 128",
		"program 128",
		"control 0 128",
		"control 128 0",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)
			lex := lexer.NewLexer([]byte(input))
			p := parser.NewParser()

			_, err := p.Parse(lex)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("range"))
		})
	}
}
