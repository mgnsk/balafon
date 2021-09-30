package ast_test

import (
	"testing"

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
			"k\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4"),
			},
		},
		{
			"k k\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4 k4"),
			},
		},
		{
			"k k8\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4 k8"),
			},
		},
		{
			"kk4\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4 k4"),
			},
		},
		{
			"k8kk16kkkk16\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k8 k16 k16 k16 k16 k16 k16"),
			},
		},
		{
			"k.\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4."),
			},
		},
		{
			"k..\n", // Double dotted note.
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4.."),
			},
		},
		{
			"k...\n", // Triple dotted note.
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4..."),
			},
		},
		{
			"k4.\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4."),
			},
		},
		{
			"k8.k16\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k8. k16"),
			},
		},
		{
			"kk8.\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k8. k8."),
			},
		},
		{
			"k/3\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4/3"),
			},
		},
		{
			"kkk8/3\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k8/3 k8/3 k8/3"),
			},
		},
		{
			"kkk8./3\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k8./3 k8./3 k8./3"),
			},
		},
		{
			"kkk128/3.\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				// Note properties are sorted.
				ContainSubstring("k128./3 k128./3 k128./3"),
			},
		},
		{
			"k k4/3kk8/3k4/3 kk8./3\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k4 k4/3 k8/3 k8/3 k4/3 k8./3 k8./3"),
			},
		},
		{
			"- k4/3--8/3k4/3 --8./3\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("-4 k4/3 -8/3 -8/3 k4/3 -8./3 -8./3"),
			},
		},
		{
			"k/3.#8\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k#8./3"),
			},
		},
		{
			"k/3.$#8\n", // A sharp flat note! Testing the ordering.
			match{
				BeAssignableToTypeOf(ast.Track{}),
				ContainSubstring("k#$8./3"),
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

			_, err := p.Parse(lex)
			g.Expect(err).To(HaveOccurred())
		})
	}
}

func TestForbiddenDuplicateProperty(t *testing.T) {
	for _, input := range []string{
		"k##",
		"k$$",
		"k/3/3",
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
