package balafon_test

import (
	"fmt"
	"testing"

	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

func TestGetPitch(t *testing.T) {
	g := NewWithT(t)

	step, octave := balafon.GetPitch(60)
	g.Expect(step).To(Equal("C"))
	g.Expect(octave).To(Equal(uint8(4)))
}

func TestXMLConvert(t *testing.T) {
	g := NewWithT(t)

	b, err := balafon.ToXML([]byte(`
:channel 9
:assign k 36
:assign s 38

:channel 0
:assign c 60

:timesig 4 4

:bar one
	:channel 9
	[- s - s]
	[k k k k]

	:channel 0
	c2   c2
:end

:bar two
	:channel 9
	-8 s.      s8. s16
	[k    k    k  k]

	:channel 0
	c2.    c
:end

:play one
:play two
`))

	g.Expect(err).NotTo(HaveOccurred())

	fmt.Println(string(b))
}
