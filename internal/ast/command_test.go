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
	for _, tc := range []struct {
		input string
		match GomegaMatcher
	}{
		{
			`assign k 36`,
			Equal(ast.CmdAssign{'k', 36}),
		},
		{
			`tempo 120`,
			Equal(ast.CmdTempo(120)),
		},
		{
			`timesig 1 1`,
			Equal(ast.CmdTimeSig{1, 1}),
		},
		{
			`channel 15`,
			Equal(ast.CmdChannel(15)),
		},
		{
			`velocity 127`,
			Equal(ast.CmdVelocity(127)),
		},
		{
			`program 127`,
			Equal(ast.CmdProgram(127)),
		},
		{
			`control 127 127`,
			Equal(ast.CmdControl{127, 127}),
		},
		{
			`bar "chorus"`,
			Equal(ast.CmdBar("chorus")),
		},
		{
			`end`,
			Equal(ast.CmdEnd{}),
		},
		{
			`play "chorus"`,
			Equal(ast.CmdPlay("chorus")),
		},
		{
			`start`,
			Equal(ast.CmdStart{}),
		},
		{
			`stop`,
			Equal(ast.CmdStop{}),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			lex := lexer.NewLexer([]byte(tc.input))
			p := parser.NewParser()

			res, err := p.Parse(lex)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(res).To(tc.match)
		})
	}
}

func TestInvalidArgumentRange(t *testing.T) {
	for _, input := range []string{
		`assign k 128`,
		`tempo 0`,
		`tempo 65536`,
		`timesig 0 1`,
		`timesig 1 0`,
		`timesig 1 129`,
		`timesig 129 1`,
		`channel 16`,
		`velocity 128`,
		`program 128`,
		`control 0 128`,
		`control 128 0`,
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

func TestInvalidTimeSig(t *testing.T) {
	for _, input := range []string{
		`timesig 4 5`,
		`timesig 2 3`,
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
