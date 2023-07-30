package balafon_test

import (
	_ "embed"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

//go:embed testdata/bonham.xml
var actualXML []byte

func TestXMLConvert(t *testing.T) {
	g := NewWithT(t)

	xml, err := balafon.ToXML(input)
	g.Expect(err).NotTo(HaveOccurred())

	_ = xml
	_ = spew.Dump
	// spew.Dump(xml)
	_ = actualXML

	// var buf bytes.Buffer
	// song.WriteTo(&buf)

	// g.Expect(buf.Bytes()).To(Equal(actualSMF))
}
