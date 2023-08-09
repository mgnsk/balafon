package ast_test

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

var scales = []string{
	// Major scales.
	"C",
	"G",
	"D",
	"A",
	"E",
	"B",
	"F#",

	"F",
	"Bb",
	"Eb",
	"Ab",
	"Db",
	"Gb",

	// Minor scales.
	"Am",
	"Em",
	"Bm",
	"F#m",
	"C#m",
	"G#m",
	"D#m",

	"Dm",
	"Gm",
	"Cm",
	"Fm",
	"Bbm",
	"Ebm",
}

func TestKey(t *testing.T) {
	for _, scale := range scales {
		t.Run(scale, func(t *testing.T) {
			g := NewWithT(t)

			input := fmt.Sprintf(":key %s", scale)
			nodeList, err := parse(input)
			g.Expect(err).NotTo(HaveOccurred())

			var buf strings.Builder
			_, err = nodeList.WriteTo(&buf)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(buf.String()).To(Equal(fmt.Sprintf(":key %s", scale)))
		})
	}
}
