package balafon_test

import (
	"bytes"
	_ "embed"
	"testing"

	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

//go:embed examples/bonham.bal
var input []byte

//go:embed testdata/bonham.mid
var actualSMF []byte

func TestSMFConvert(t *testing.T) {
	g := NewWithT(t)

	song, err := balafon.Convert(input)
	g.Expect(err).NotTo(HaveOccurred())

	var buf bytes.Buffer
	song.WriteTo(&buf)

	g.Expect(buf.Bytes()).To(Equal(actualSMF))
}
