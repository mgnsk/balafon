package ast_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	. "github.com/onsi/gomega"
)

func TestComment(t *testing.T) {
	g := NewGomegaWithT(t)

	lex := lexer.NewLexer([]byte("// this is a comment\n"))
	p := parser.NewParser()

	res, err := p.Parse(lex)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(BeNil())
}
