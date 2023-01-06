package ast_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/mgnsk/gong/ast"
	. "github.com/onsi/gomega"
)

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

	input1 := `

tempo 120
timesig 1 4
channel 1
velocity 20
program 1
control 1 1
assign c 60
assign d 62
bar "Bar 1"
	start
	c
	stop
end
play "Bar 1"

`

	input2 := `
tempo 120; timesig 1 4; channel 1; velocity 20;
program 1; control 1 1
assign c 60; assign d 62
bar "Bar 1" start; c; stop; end
play "Bar 1"
`

	res1, err := parse(input1)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res1).To(BeAssignableToTypeOf(ast.NodeList{}))

	var buf1 bytes.Buffer
	res1.(ast.NodeList).WriteTo(&buf1)

	g.Expect(buf1.String()).To(Equal(strings.Trim(input1, " \n")))

	res2, err := parse(input2)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res2).To(BeAssignableToTypeOf(ast.NodeList{}))
	g.Expect(res2).To(Equal(res1))

	var buf2 bytes.Buffer
	res2.(ast.NodeList).WriteTo(&buf2)
	g.Expect(buf2.String()).To(Equal(strings.Trim(input1, " \n")))
}

func TestCommandsForbiddenInBar(t *testing.T) {
	for _, input := range []string{
		"assign c 60",
		`bar "Inner" start end`,
		`play "test"`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			_, err := parse(fmt.Sprintf(`bar "Outer" %s; end`, input))
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(`got:`))
		})
	}
}
