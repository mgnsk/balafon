package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
)

func TestNoteAssignment(t *testing.T) {
	type (
		match    []GomegaMatcher
		testcase struct {
			input string
			match match
		}
	)

	for _, tc := range []testcase{
		{
			"c = 48\n",
			match{
				BeAssignableToTypeOf(ast.NoteAssignment{}),
				ContainSubstring("c = 48"),
			},
		},
		{
			"c=16\n",
			match{
				BeAssignableToTypeOf(ast.NoteAssignment{}),
				ContainSubstring("c = 16"),
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

func TestInvalidAssignment(t *testing.T) {
	g := NewGomegaWithT(t)

	lex := lexer.NewLexer([]byte("cc = 10\n"))
	p := parser.NewParser()

	_, err := p.Parse(lex)
	g.Expect(err).To(HaveOccurred())
}
