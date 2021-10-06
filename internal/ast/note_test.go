package ast_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
)

func TestValidInputs(t *testing.T) {
	type (
		match    []GomegaMatcher
		testcase struct {
			input string
			match match
		}
	)

	for _, tc := range []testcase{
		{
			"k",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4"),
			},
		},
		{
			"kk",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4 k4"),
			},
		},
		{
			"k k",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4 k4"),
			},
		},
		{
			"k k8",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4 k8"),
			},
		},
		{
			"kk8.", // Properties apply only to the previous note symbol.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4 k8."),
			},
		},
		{
			"[kk.]8", // Group properties apply to all notes in the group.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k8 k8."),
			},
		},
		{
			"[k.].", // Group properties override any duplicate properties the inner notes have.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4."),
			},
		},
		{
			"[k]",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4"),
			},
		},
		{
			"[k][k].",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4 k4."),
			},
		},
		{
			"kk[kk]kk[kk]kk",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4 k4 k4 k4 k4 k4 k4 k4 k4 k4"),
			},
		},
		{
			"[[k]]8",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k8"),
			},
		},
		{
			"k8kk16kkkk16",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k8 k4 k16 k4 k4 k4 k16"),
			},
		},
		{
			"k8 [kk]16 [kkkk]32",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k8 k16 k16 k32 k32 k32 k32"),
			},
		},
		{
			"k..", // Double dotted note.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4.."),
			},
		},
		{
			"k...", // Triple dotted note.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4..."),
			},
		},
		{
			"k4/3", // Triplet.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k4/3"),
			},
		},
		{
			"-", // Pause.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("-4"),
			},
		},
		{
			"k/3.#8\n",
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k#8./3"),
			},
		},
		{
			"[[[[[k]/3].]#]8]^\n", // Testing the ordering of properties.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k#^8./3"),
			},
		},
		{
			"[[[[[k*]/3].]$]8])\n", // Testing the ordering of properties.
			match{
				BeAssignableToTypeOf(ast.NoteList(nil)),
				ContainSubstring("k$)8./3*"),
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

func TestInvalidProperties(t *testing.T) {
	for _, input := range []string{
		"k#$", // Sharp flat note.
		"k$#",
		"k^)", // Accentuated ghost note.
		"k)^",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			lex := lexer.NewLexer([]byte(input))
			p := parser.NewParser()

			_, err := p.Parse(lex)
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

			lex := lexer.NewLexer([]byte(input))
			p := parser.NewParser()

			r, err := p.Parse(lex)
			spew.Dump(r)
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

			lex := lexer.NewLexer([]byte(input))
			p := parser.NewParser()

			_, err := p.Parse(lex)
			g.Expect(err).To(HaveOccurred())
		})
	}
}
