package ast_test

import (
	"strings"
	"testing"

	"github.com/mgnsk/balafon/internal/ast"
	"github.com/mgnsk/balafon/internal/constants"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestNoteList(t *testing.T) {
	type (
		match    types.GomegaMatcher
		testcase struct {
			input    string
			expected string
		}
	)

	for _, tc := range []testcase{
		{
			"k",
			"k",
		},
		{
			"kk",
			"kk",
		},
		{
			"kk8",
			"kk8",
		},
		{
			"kk8.", // Properties apply only to the previous note symbol.
			"kk8.",
		},
		{
			"[kk.]8", // Group properties apply to all notes in the group.
			"[kk.]8",
		},
		{
			"[k.].", // Group properties are appended.
			"[k.].",
		},
		{
			"[k]",
			"[k]",
		},
		{
			"[k][k].",
			"[k][k].",
		},
		{
			"kk[kk]kk[kk]kk",
			"kk[kk]kk[kk]kk",
		},
		{
			"[[k]]8",
			"[[k]]8",
		},
		{
			"k8kk16kkkk16",
			"k8kk16kkkk16",
		},
		{
			"k8 [kk]16 [kkkk]32",
			"k8[kk]16[kkkk]32",
		},
		{
			"-", // Pause.
			"-",
		},
		{
			"-8", // 8th pause.
			"-8",
		},
		{
			"k/3.#8",
			"k#8./3",
		},
		{
			"[[[[[k]/3].]#]8]>>^^``", // Testing the ordering of properties.
			"[[[[[k]/3].]#]8]``>>^^",
		},
		{
			"[[[[[k*]/3].]$].8]))", // Testing the ordering of properties.
			"[[[[[k*]/3].]$]8.]))",
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			res, err := parse(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			var buf strings.Builder
			_, err = res.WriteTo(&buf)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(buf.String()).To(Equal(tc.expected))
		})
	}
}

func TestInvalidNoteValue(t *testing.T) {
	for _, input := range []string{
		"k3",
		"k22",
		"k0",
		"k129",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			_, err := parse(input)
			g.Expect(err).To(HaveOccurred())
		})
	}
}

func TestNoteLengths(t *testing.T) {
	for _, tc := range []struct {
		input string
		offAt uint32
	}{
		{
			input: "k", // Quarter note.
			offAt: uint32(constants.TicksPerQuarter),
		},
		{
			input: "k.", // Dotted quarter note, x1.5.
			offAt: uint32(constants.TicksPerQuarter * 3 / 2),
		},
		{
			input: "k..", // Double dotted quarter note, x1.75.
			offAt: uint32(constants.TicksPerQuarter * 7 / 4),
		},
		{
			input: "k...", // Triplet dotted quarter note, x1.875.
			offAt: uint32(constants.TicksPerQuarter * 15 / 8),
		},
		{
			input: "k/5", // Quintuplet quarter note.
			offAt: uint32(constants.TicksPerQuarter * 2 / 5),
		},
		{
			input: "k./3", // Dotted triplet quarter note == quarter note.
			offAt: uint32(constants.TicksPerQuarter),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			res, err := parse(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			note := res[0].(ast.NodeList)[0]
			g.Expect(note).To(BeAssignableToTypeOf(&ast.Note{}))
			g.Expect(note.(*ast.Note).Props.NoteLen()).To(Equal(tc.offAt))
		})
	}
}
