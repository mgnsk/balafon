package ast_test

import (
	"bytes"
	"testing"

	"github.com/mgnsk/balafon/internal/ast"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestValidCommands(t *testing.T) {
	for _, tc := range []struct {
		input string
		match types.GomegaMatcher
	}{
		{
			`:assign k 36`,
			Equal(ast.CmdAssign{Note: 'k', Key: 36}),
		},
		{
			`:tempo 120`,
			Equal(ast.CmdTempo{BPM: 120}),
		},
		{
			`:time 1 1`,
			Equal(ast.CmdTime{Num: 1, Denom: 1}),
		},
		{
			`:channel 16`,
			Equal(ast.CmdChannel{Channel: 15}),
		},
		{
			`:voice 4`,
			Equal(ast.CmdVoice{Voice: 16}),
		},
		{
			`:velocity 127`,
			Equal(ast.CmdVelocity{Velocity: 127}),
		},
		{
			`:program 127`,
			Equal(ast.CmdProgram{Program: 127}),
		},
		{
			`:control 127 127`,
			Equal(ast.CmdControl{Control: 127, Parameter: 127}),
		},
		{
			`:play chorus`,
			Equal(ast.CmdPlay{BarName: "chorus"}),
		},
		{
			`:start`,
			Equal(ast.CmdStart{}),
		},
		{
			`:stop`,
			Equal(ast.CmdStop{}),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			res, err := parse(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			var buf bytes.Buffer
			res.WriteTo(&buf)
			g.Expect(buf.String()).To(Equal(tc.input))
		})
	}
}

func TestInvalidArgumentRange(t *testing.T) {
	for _, input := range []string{
		`:assign k 128`,
		`:tempo 0`,
		`:tempo 65536`,
		`:time 0 1`,
		`:time 1 0`,
		`:time 1 129`,
		`:time 129 1`,
		`:channel 17`,
		`:voice 5`,
		`:velocity 128`,
		`:program 128`,
		`:control 0 128`,
		`:control 128 0`,
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
		`:time 4 5`,
		`:time 2 3`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			_, err := parse(input)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("range"))
		})
	}
}
