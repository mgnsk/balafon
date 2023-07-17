package interpreter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mgnsk/balafon/internal/constants"
	"github.com/mgnsk/balafon/interpreter"
	. "github.com/onsi/gomega"
)

func TestBarCap(t *testing.T) {
	g := NewWithT(t)

	bar := interpreter.Bar{
		TimeSig: [2]uint8{4, 4},
	}

	g.Expect(bar.Cap()).To(BeEquivalentTo(constants.TicksPerWhole))
}

func TestBarDurationMultiTrack(t *testing.T) {
	g := NewWithT(t)

	bar := interpreter.Bar{
		TimeSig: [2]uint8{1, 4},
		Events: []interpreter.Event{
			{
				Duration: uint32(constants.TicksPerQuarter),
			},
			{
				Duration: uint32(constants.TicksPerQuarter),
			},
		},
	}

	g.Expect(bar.Duration(60)).To(Equal(time.Second))
}

func TestBarDurationTimeSignatures(t *testing.T) {
	for _, tc := range []struct {
		timesig string
		input   string
	}{
		{
			timesig: "1 4",
			input:   "c",
		},
		{
			timesig: "2 8",
			input:   "c",
		},
	} {
		t.Run(fmt.Sprintf(":timesig %s", tc.timesig), func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.EvalString(fmt.Sprintf(":timesig %s", tc.timesig))).To(Succeed())
			g.Expect(it.EvalString(":assign c 60")).To(Succeed())
			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Duration(60)).To(Equal(time.Second))
		})
	}
}

func TestEmptyBarIsInvalid(t *testing.T) {
	g := NewWithT(t)

	input := `
:bar mybar
:timesig 1 1
:velocity 1
:channel 1
:end
	`

	it := interpreter.New()
	g.Expect(it.EvalString(input)).NotTo(Succeed())
}
