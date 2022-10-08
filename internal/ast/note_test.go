package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/ast"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/types"
)

func TestValidInputs(t *testing.T) {
	type (
		match    types.GomegaMatcher
		testcase struct {
			input string
			match match
		}
	)

	for _, tc := range []testcase{
		{
			"k",
			ContainSubstring("k4"),
		},
		{
			"kk",
			ContainSubstring("k4 k4"),
		},
		{
			"k k",
			ContainSubstring("k4 k4"),
		},
		{
			"k k8",
			ContainSubstring("k4 k8"),
		},
		{
			"kk8.", // Properties apply only to the previous note symbol.
			ContainSubstring("k4 k8."),
		},
		{
			"[kk.]8", // Group properties apply to all notes in the group.
			ContainSubstring("k8 k8."),
		},
		{
			"[k.].", // Group properties override any duplicate properties the inner notes have.
			ContainSubstring("k4."),
		},
		{
			"[k]",
			ContainSubstring("k4"),
		},
		{
			"[k][k].",
			ContainSubstring("k4 k4."),
		},
		{
			"kk[kk]kk[kk]kk",
			ContainSubstring("k4 k4 k4 k4 k4 k4 k4 k4 k4 k4"),
		},
		{
			"[[k]]8",
			ContainSubstring("k8"),
		},
		{
			"k8kk16kkkk16",
			ContainSubstring("k8 k4 k16 k4 k4 k4 k16"),
		},
		{
			"k8 [kk]16 [kkkk]32",
			ContainSubstring("k8 k16 k16 k32 k32 k32 k32"),
		},
		{
			"k..", // Double dotted note.
			ContainSubstring("k4.."),
		},
		{
			"k...", // Triple dotted note.
			ContainSubstring("k4..."),
		},
		{
			"k4/3", // Triplet.
			ContainSubstring("k4/3"),
		},
		{
			"-", // Pause.
			ContainSubstring("-4"),
		},
		{
			"k/3.#8",
			ContainSubstring("k#8./3"),
		},
		{
			"[[[[[k]/3].]#]8]^", // Testing the ordering of properties.
			ContainSubstring("k#^8./3"),
		},
		{
			"[[[[[k*]/3].]$]8])", // Testing the ordering of properties.
			ContainSubstring("k$)8./3*"),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			res, err := parse(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(res).To(BeAssignableToTypeOf(ast.DeclList{}))
			list := res.(ast.DeclList)
			g.Expect(list).To(HaveLen(1))
			g.Expect(list[0]).To(tc.match)
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
		"k##",
		"k$$",
		"k^^",
		"k))",
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
