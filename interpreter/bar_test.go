package interpreter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mgnsk/gong/constants"
	"github.com/mgnsk/gong/interpreter"
	. "github.com/onsi/gomega"
)

func TestBarCap(t *testing.T) {
	g := NewWithT(t)

	bar := interpreter.Bar{
		TimeSig: [2]uint8{4, 4},
	}

	g.Expect(bar.Cap()).To(BeEquivalentTo(constants.TicksPerWhole))
}

func TestBarLen(t *testing.T) {
	g := NewWithT(t)

	bar := interpreter.Bar{
		TimeSig: [2]uint8{4, 4},
		Events: []interpreter.Event{
			{
				Channel:  0,
				Duration: uint32(constants.TicksPerWhole),
			},
			{
				Channel:  1,
				Duration: uint32(constants.TicksPerWhole),
			},
		},
	}

	g.Expect(bar.Len()).To(BeEquivalentTo(constants.TicksPerWhole))
}

func TestBarDurationMultiTrack(t *testing.T) {
	g := NewWithT(t)

	bar := interpreter.Bar{
		TimeSig: [2]uint8{1, 4},
		Tempo:   60,
		Events: []interpreter.Event{
			{
				Channel:  0,
				Duration: uint32(constants.TicksPerQuarter),
			},
			{
				Channel:  1,
				Duration: uint32(constants.TicksPerQuarter),
			},
		},
	}

	g.Expect(bar.Duration()).To(Equal(time.Second))
}

func TestBarDurationTimeSignatures(t *testing.T) {
	for _, tc := range []struct {
		timesig string
		tempo   uint
		input   string
	}{
		{
			timesig: "1 4",
			tempo:   60,
			input:   "c",
		},
		{
			timesig: "2 8",
			tempo:   60,
			input:   "c",
		},
		{
			timesig: "2 4",
			tempo:   120,
			input:   "c2",
		},
	} {
		t.Run(fmt.Sprintf("timesig %s tempo %d", tc.timesig, tc.tempo), func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.Eval(fmt.Sprintf("timesig %s", tc.timesig))).To(Succeed())
			g.Expect(it.Eval(fmt.Sprintf("tempo %d", tc.tempo))).To(Succeed())
			g.Expect(it.Eval("assign c 60")).To(Succeed())
			g.Expect(it.Eval(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Duration()).To(Equal(time.Second))
		})
	}
}
