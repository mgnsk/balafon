package balafon_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

func TestKeyAccidentalsPitch(t *testing.T) {
	for _, tc := range []struct {
		input string

		// MIDI:
		key uint8

		// MusicXML:
		alter int
		step  string
	}{
		{":assign f 65; :key G; f", 66, 1, "F"},
		{":assign f 65; :key G; f#", 66, 1, "F"}, // courtesy accidental
		{":assign f 65; :key G; f$", 64, -1, "F"},
		{":assign f 66; :key G; f", 66, 1, "F"},

		{":assign b 71; :key Dm; b", 70, -1, "B"},
		{":assign b 71; :key Dm; b$", 70, -1, "B"}, // courtesy accidental
		{":assign b 71; :key Dm; b#", 72, 1, "B"},
		{":assign b 70; :key Dm; b", 70, -1, "B"},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			{
				p := balafon.New()
				g.Expect(p.EvalString(tc.input)).To(Succeed())

				bars := p.Flush()
				g.Expect(bars).To(HaveLen(1))

				_, key, _, ok := findNote(bars[0])
				g.Expect(ok).To(BeTrue())
				g.Expect(key).To(Equal(tc.key))
			}

			{
				var buf strings.Builder
				g.Expect(balafon.ToXML(&buf, []byte(tc.input))).To(Succeed())

				g.Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("<alter>%d</alter>", tc.alter)))
				g.Expect(buf.String()).To(ContainSubstring(fmt.Sprintf("<step>%s</step>", tc.step)))
			}
		})
	}
}
