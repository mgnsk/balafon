package parser_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/lexer"
	"github.com/mgnsk/gong/internal/parser"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/types"
)

func TestParser(t *testing.T) {
	type (
		match    []GomegaMatcher
		testcase struct {
			input string
			match match
		}
	)

	for _, tc := range []testcase{
		{
			"tempo = 120",
			match{
				Equal(&ast.Assignment{Name: "tempo", Value: 120}),
				ContainSubstring("tempo = 120"),
			},
		},
		{
			"c = 48",
			match{
				Equal(&ast.Assignment{Name: "c", Value: 48}),
				ContainSubstring("c = 48"),
			},
		},
		{
			"c=48",
			match{
				Equal(&ast.Assignment{Name: "c", Value: 48}),
				ContainSubstring("c = 48"),
			},
		},
		{
			"k",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4"),
			},
		},
		{
			"k k",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4 k4"),
			},
		},
		{
			"k k8",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4 k8"),
			},
		},
		{
			"kk4",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4 k4"),
			},
		},
		{
			"k8kk16kkkk16",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k8 k16 k16 k16 k16 k16 k16"),
			},
		},
		{
			"k.",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4."),
			},
		},
		{
			"k4.",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4."),
			},
		},
		{
			"k8.k16",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k8. k16"),
			},
		},
		{
			"kk8.",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k8. k8."),
			},
		},
		{
			"k/3",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4/3"),
			},
		},
		{
			"kkk8/3",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k8/3 k8/3 k8/3"),
			},
		},
		{
			"kkk8./3",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k8./3 k8./3 k8./3"),
			},
		},
		{
			"kkk8/3.",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				// Note properties are sorted.
				ContainSubstring("k8./3 k8./3 k8./3"),
			},
		},
		{
			"k k4/3kk8/3k4/3 kk8./3",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("k4 k4/3 k8/3 k8/3 k4/3 k8./3 k8./3"),
			},
		},
		{
			"- k4/3--8/3k4/3 --8./3",
			match{
				BeAssignableToTypeOf((*ast.Track)(nil)),
				ContainSubstring("-4 k4/3 -8/3 -8/3 k4/3 -8./3 -8./3"),
			},
		},
		{
			"bar MyRiff",
			match{
				Equal(&ast.Command{Name: "bar", Arg: "MyRiff"}),
			},
		},
		{
			"end",
			match{
				Equal(&ast.Command{Name: "end"}),
			},
		},
		{
			"play MyRiff",
			match{
				Equal(&ast.Command{Name: "play", Arg: "MyRiff"}),
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
