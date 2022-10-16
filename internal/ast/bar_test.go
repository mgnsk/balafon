package ast_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
)

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

	input1 := `assign c 60
assign d 62
bar "Bar 1" {
	start
	c4
	stop
}
play "Bar 1"`

	input2 := `assign c 60; assign d 62
bar "Bar 1" { start; c; stop }
play "Bar 1"
`

	res1, err := parse(input1)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(fmt.Sprint(res1)).To(Equal(input1))

	res2, err := parse(input2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(fmt.Sprint(res2)).To(Equal(input1))

	g.Expect(res2).To(Equal(res1))
}
