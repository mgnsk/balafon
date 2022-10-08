package ast_test

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestComment(t *testing.T) {
	g := NewGomegaWithT(t)

	res, err := parse("// this is a comment")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res).To(BeNil())
}
