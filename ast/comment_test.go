package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/ast"
	. "github.com/onsi/gomega"
)

func TestComments(t *testing.T) {
	g := NewGomegaWithT(t)

	for _, input := range []string{
		`
// this is a line comment
assign c 60
`,
		`
/* this is a block comment */
assign c 60
`,
		`
// this is a line comment
/*
this is a
block comment
*/
assign c 60
`,
	} {
		res, err := parse(input)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(res).To(Equal(ast.NodeList{
			ast.CmdAssign{Note: 'c', Key: 60},
		}))
	}
}
