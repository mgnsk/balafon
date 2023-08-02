package balafon_test

import (
	"bytes"
	"testing"

	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

func TestXMLSharpNotes(t *testing.T) {
	t.Run("natural base note plus sharp", func(t *testing.T) {
		g := NewWithT(t)

		var buf bytes.Buffer
		err := balafon.ToXML(&buf, []byte(`
:assign c 60
c#
`))

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(buf.String()).To(ContainSubstring("<alter>1</alter>"))
		g.Expect(buf.String()).To(ContainSubstring("<step>C</step>"))
		g.Expect(buf.String()).To(ContainSubstring("<octave>4</octave>"))
	})

	t.Run("sharp base note", func(t *testing.T) {
		g := NewWithT(t)

		var buf bytes.Buffer
		err := balafon.ToXML(&buf, []byte(`
:assign c 61
c
`))

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(buf.String()).To(ContainSubstring("<alter>1</alter>"))
		g.Expect(buf.String()).To(ContainSubstring("<step>C</step>"))
		g.Expect(buf.String()).To(ContainSubstring("<octave>4</octave>"))
	})
}

func TestXMLChords(t *testing.T) {
	t.Run("single voice can chord", func(t *testing.T) {
		g := NewWithT(t)

		var buf bytes.Buffer
		err := balafon.ToXML(&buf, []byte(`
:assign c 60
:bar one
	c
	c
:end
:play one
`))

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(buf.String()).To(ContainSubstring("<chord>"))
	})

	t.Run("multiple voice cannot chord", func(t *testing.T) {
		g := NewWithT(t)

		var buf bytes.Buffer
		err := balafon.ToXML(&buf, []byte(`

	:assign c 60
	:bar one
	:voice 1
	c
	:voice 2
	c
	:end
	:play one
	`))

		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(buf.String()).NotTo(ContainSubstring("<chord>"))
	})
}
