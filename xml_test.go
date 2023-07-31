package balafon_test

import (
	"encoding/xml"
	"fmt"
	"testing"

	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

func TestXMLConvert(t *testing.T) {
	g := NewWithT(t)

	doc, err := balafon.ToXML(input)
	g.Expect(err).NotTo(HaveOccurred())

	b, err := xml.MarshalIndent(doc, "", "    ")
	g.Expect(err).NotTo(HaveOccurred())

	fmt.Println(string(b))
}
