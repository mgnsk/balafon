package interpreter_test

import (
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

func TestBarDuration(t *testing.T) {
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
