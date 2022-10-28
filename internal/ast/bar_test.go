package ast_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mgnsk/gong/internal/ast"
	. "github.com/onsi/gomega"
)

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

	input1 := `

assign c 60
assign d 62
bar "Bar 1" {
	start
	c4
	stop
}
play "Bar 1"

`

	input2 := `assign c 60; assign d 62
bar "Bar 1" { start; c; stop }
play "Bar 1"
`

	res1, err := parse(input1)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res1).To(BeAssignableToTypeOf(ast.DeclList{}))

	g.Expect(fmt.Sprint(res1)).To(Equal(strings.Trim(input1, " \n")))

	res2, err := parse(input2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res2).To(BeAssignableToTypeOf(ast.DeclList{}))
	g.Expect(res2).To(Equal(res1))

	g.Expect(fmt.Sprint(res2)).To(Equal(strings.Trim(input1, " \n")))
}

func TestNestedBarForbidden(t *testing.T) {
	g := NewGomegaWithT(t)

	input := `
bar "Bar 1" {
    bar "Bar 2" {
        start
    }
}
    `

	_, err := parse(input)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(HaveSuffix(`got: "bar"`))
}