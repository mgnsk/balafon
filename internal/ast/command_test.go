package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/ast"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestValidCommands(t *testing.T) {
	for _, tc := range []struct {
		input string
		match types.GomegaMatcher
	}{
		{
			`assign k 36`,
			Equal(ast.Song{ast.CmdAssign{'k', 36}}),
		},
		{
			`tempo 120`,
			Equal(ast.Song{ast.CmdTempo(120)}),
		},
		{
			`timesig 1 1`,
			Equal(ast.Song{ast.CmdTimeSig{1, 1}}),
		},
		{
			`channel 15`,
			Equal(ast.Song{ast.CmdChannel(15)}),
		},
		{
			`velocity 127`,
			Equal(ast.Song{ast.CmdVelocity(127)}),
		},
		{
			`program 127`,
			Equal(ast.Song{ast.CmdProgram(127)}),
		},
		{
			`control 127 127`,
			Equal(ast.Song{ast.CmdControl{127, 127}}),
		},
		{
			`end`,
			Equal(ast.Song{ast.CmdEnd{}}),
		},
		{
			`play "chorus"`,
			Equal(ast.Song{ast.CmdPlay("chorus")}),
		},
		{
			`start`,
			Equal(ast.Song{ast.CmdStart{}}),
		},
		{
			`stop`,
			Equal(ast.Song{ast.CmdStop{}}),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			res, err := parse(tc.input)
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

			_, err := parse(input)
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

			_, err := parse(input)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("range"))
		})
	}
}
