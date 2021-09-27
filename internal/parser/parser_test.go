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
			"# this is a comment\n",
			match{
				BeNil(),
				// BeAssignableToTypeOf(ast.LineComment{}),
			},
		},
		{
			"tempo = 120\n",
			match{
				BeAssignableToTypeOf(ast.Assignment{}),
				ContainSubstring("tempo = 120"),
			},
		},
		{
			"c = 48\n",
			match{
				BeAssignableToTypeOf(ast.Assignment{}),
				ContainSubstring("c = 48"),
			},
		},
		{
			"c=48\n",
			match{
				BeAssignableToTypeOf(ast.Assignment{}),
				ContainSubstring("c = 48"),
			},
		},
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
			"kkk8/3.\n",
			match{
				BeAssignableToTypeOf(ast.Track{}),
				// Note properties are sorted.
				ContainSubstring("k8./3 k8./3 k8./3"),
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
			"bar MyRiff\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("bar MyRiff"),
			},
		},
		{
			"end\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("end"),
			},
		},
		{
			"play MyRiff\n",
			match{
				BeAssignableToTypeOf(ast.Command{}),
				ContainSubstring("play MyRiff"),
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
