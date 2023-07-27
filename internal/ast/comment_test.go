package ast_test

import (
	"testing"

	"github.com/mgnsk/balafon/internal/ast"
	. "github.com/onsi/gomega"
)

func TestBlockComment(t *testing.T) {
	g := NewGomegaWithT(t)

	for _, input := range []string{
		`
/* this is a single line block comment */
:assign c 60
`,
		`
/*
this is a multi line
block comment
*/
:assign c 60
`,
	} {
		res, err := parse(input)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(res).To(ConsistOf(
			BeAssignableToTypeOf(ast.BlockComment{}),
			BeAssignableToTypeOf(ast.CmdAssign{}),
		))
	}
}

// func TestLineComment(t *testing.T) {
// 	g := NewGomegaWithT(t)

// 	for _, input := range []string{
// 		`
// // this is a line comment
// :assign c 60
// `,
// 	} {
// 		res, err := parse(input)
// 		g.Expect(err).NotTo(HaveOccurred())
// 		g.Expect(res).To(ConsistOf(
// 			BeAssignableToTypeOf(ast.LineComment{}),
// 			BeAssignableToTypeOf(ast.CmdAssign{}),
// 		))
// 	}
// }
