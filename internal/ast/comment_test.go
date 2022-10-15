package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/ast"
	. "github.com/onsi/gomega"
)

func TestLineComment(t *testing.T) {
	g := NewGomegaWithT(t)

	res, err := parse("// this is a line comment\n")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(ast.StmtList{ast.LineComment(" this is a line comment")}))
}

func TestBlockComment(t *testing.T) {
	g := NewGomegaWithT(t)

	res, err := parse("/*this is a block comment*/")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(ast.StmtList{ast.BlockComment("this is a block comment")}))

	res, err = parse("/*\nthis is a\nblock comment\n*/")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(Equal(ast.StmtList{ast.BlockComment("\nthis is a\nblock comment\n")}))
}
