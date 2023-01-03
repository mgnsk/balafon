package ast_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/mgnsk/gong/internal/ast"
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
			"k8k8.",
		},
		{
			"[k.].", // Group properties are appended.
			"k..",
		},
		{
			"[k]",
			"k",
		},
		{
			"[k][k].",
			"kk.",
		},
		{
			"kk[kk]kk[kk]kk",
			"kkkkkkkkkk",
		},
		{
			"[[k]]8",
			"k8",
		},
		{
			"k8kk16kkkk16",
			"k8kk16kkkk16",
		},
		{
			"k8 [kk]16 [kkkk]32",
			"k8k16k16k32k32k32k32",
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
			"[[[[[k]/3].]#]8]^^", // Testing the ordering of properties.
			"k#^^8./3",
		},
		{
			"[[[[[k*]/3].]$].8]))", // Testing the ordering of properties.
			"k$))8../3*",
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			res, err := parse(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			wt, ok := res.(io.WriterTo)
			g.Expect(ok).To(BeTrue())
			_ = wt

			var buf bytes.Buffer
			_, err = wt.WriteTo(&buf)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(buf.String()).To(Equal(tc.expected))
		})
	}
}

func TestInvalidProperties(t *testing.T) {
	for _, input := range []string{
		"k#$", // Sharp flat note.
		"k$#",
		"k^)", // Accentuated ghost note.
		"k)^",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			_, err := parse(input)
			g.Expect(err).To(HaveOccurred())
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

func TestForbiddenDuplicateProperty(t *testing.T) {
	for _, input := range []string{
		// TODO: allow double sharp and flat?
		"k44",
		"k##",
		"k$$",
		"k/3/3",
		"k**",
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
		input   string
		len32th uint8
	}{
		{
			input:   "k", // Quarter note.
			len32th: 8,
		},
		{
			input:   "k.", // Dotted quarter note, x1.5.
			len32th: 12,
		},
		{
			input:   "k..", // Double dotted quarter note, x1.75.
			len32th: 14,
		},
		{
			input:   "k...", // Triple dotted quarter note, x1.875.
			len32th: 15,
		},
		{
			input:   "k/3", // Triplet quarter note.
			len32th: 5,     // TODO: precision
		},
		{
			input:   "k/5", // Quintuplet quarter note.
			len32th: 3,     // TODO: precision
		},
		{
			input:   "k./3", // Dotted triplet quarter note == quarter note.
			len32th: 8,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			res, err := parse(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			note := res.(ast.NodeList)[0].(ast.NoteList)[0]
			g.Expect(note.Len()).To(Equal(tc.len32th))
		})
	}
}
