package balafon_test

import (
	"testing"

	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

func TestTrailingNewlineIsAdded(t *testing.T) {
	g := NewWithT(t)

	input := `:assign c 60
:assign d 62`

	res, err := balafon.Format([]byte(input))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(res)).To(Equal(`:assign c 60
:assign d 62
`))
}

func TestFmtCollapseEmptyLines(t *testing.T) {
	g := NewWithT(t)

	input := `:assign c 60


:assign d 62
`

	res, err := balafon.Format([]byte(input))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(res)).To(Equal(`:assign c 60

:assign d 62
`))
}

func TestFmtBarIndent(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign c 60


	:bar bar1
   :timesig 4 4
    c.            d8 [e$ e f f#]8
	 [-CE$G]16 c2          [B$A]8
:end


`

	res, err := balafon.Format([]byte(input))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(res)).To(Equal(`:assign c 60

:bar bar1
	:timesig 4 4
	c.            d8 [e$ e f f#]8
	[-CE$G]16 c2          [B$A]8
:end
`))
}

func TestFmtCommand(t *testing.T) {
	g := NewWithT(t)

	input := `:assign  c  60;`

	res, err := balafon.Format([]byte(input))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(res)).To(Equal(":assign c 60\n"))
}

func TestFmtBarCommand(t *testing.T) {
	g := NewWithT(t)

	input := `
:bar	  a
:end
`

	res, err := balafon.Format([]byte(input))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(res)).To(Equal(":bar a\n:end\n"))
}

func TestFmtPlayCommand(t *testing.T) {
	g := NewWithT(t)

	input := `
:play	  a
`

	res, err := balafon.Format([]byte(input))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(string(res)).To(Equal(":play a\n"))
}
