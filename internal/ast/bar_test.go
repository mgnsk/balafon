package ast_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/mgnsk/balafon/internal/ast"
	. "github.com/onsi/gomega"
)

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

	input1 := `

:tempo 120
:timesig 1 4
:channel 1
:velocity 20
:program 1
:control 1 1
:assign c 60
:assign d 62
:bar bar1
	:start
	c
	:stop
:end
:play bar1

`

	input2 := `
:tempo 120; :timesig 1 4; :channel 1; :velocity 20;
:program 1; :control 1 1
:assign c 60; :assign d 62
:bar bar1 :start; c; :stop; :end
:play bar1
`

	input1Clean := strings.Trim(input1, " \n")

	res1, err := parse(input1)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res1).To(BeAssignableToTypeOf(ast.NodeList{}))

	var buf1 bytes.Buffer
	res1.(ast.NodeList).WriteTo(&buf1)
	g.Expect(buf1.String()).To(Equal(input1Clean))

	res2, err := parse(input2)
	g.Expect(err).NotTo(HaveOccurred())

	var buf2 bytes.Buffer
	res2.(ast.NodeList).WriteTo(&buf2)
	g.Expect(buf2.String()).To(Equal(input1Clean))
}

func TestBarIdentifierAllowedNumeric(t *testing.T) {
	for _, input := range []string{
		`:bar bar c :end`,
		`:bar 1 c :end`,
		`:play 1`,
		`:bar 1a c :end`,
		`:play 1a`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			_, err := parse(input)
			g.Expect(err).NotTo(HaveOccurred())
		})
	}
}
