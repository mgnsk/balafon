package balafon_test

import (
	_ "embed"
	"testing"

	"github.com/aymanbagabas/go-udiff"
	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

//go:embed examples/bonham.bal
var input []byte

//go:embed testdata/bonham.txt
var actualSMF string

func TestSMFConvert(t *testing.T) {
	g := NewWithT(t)

	song, err := balafon.ToSMF(input)
	g.Expect(err).NotTo(HaveOccurred())

	// os.WriteFile("testdata/bonham.txt", []byte(song.String()), 0644)

	diff := udiff.Unified("bonham.txt", "generated.txt", actualSMF, song.String())

	g.Expect(diff).To(BeEmpty())
}
