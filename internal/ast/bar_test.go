package ast_test

import (
	"testing"

	"github.com/davecgh/go-spew/spew"
	. "github.com/onsi/gomega"
)

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

	input := `assign c 60
    bar "Bar 1" {
        start
        c
    }
    `

	res, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())

	spew.Dump(res)

}
