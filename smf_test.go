package balafon_test

import (
	_ "embed"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

//go:embed examples/bonham.bal
var input []byte

func TestSMFConvert(t *testing.T) {
	g := NewWithT(t)

	song, err := balafon.Convert(input)
	g.Expect(err).NotTo(HaveOccurred())

	spew.Dump(song)

	// g.Expect(it.Eval(input)).To(Succeed())

	// bars := it.Flush()

	// _ = bars
	// _ = spew.Dump
}
