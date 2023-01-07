package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/ast"
	. "github.com/onsi/gomega"
)

func TestLineComment(t *testing.T) {
	g := NewGomegaWithT(t)

	res, err := parse(`
	// this is a line comment

	assign c 60
`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(ast.NodeList{
		ast.LineComment(" this is a line comment"),
		ast.CmdAssign{Note: 'c', Key: 60},
	}))
}

func TestBlockComment(t *testing.T) {
	g := NewGomegaWithT(t)

	res, err := parse(`
/* this is a block comment */
assign c 60
`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(ast.NodeList{
		ast.BlockComment(" this is a block comment "),
		ast.CmdAssign{Note: 'c', Key: 60},
	}))

	res, err = parse(`
/*
this is a
block comment
*/
`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(ast.NodeList{ast.BlockComment("\nthis is a\nblock comment\n")}))
}
