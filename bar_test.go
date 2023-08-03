package balafon_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/balafon"
	"github.com/mgnsk/balafon/internal/constants"
	. "github.com/onsi/gomega"
)

func TestBarCapTimeSignatures(t *testing.T) {
	for _, tc := range []struct {
		timesig  [2]uint8
		capacity uint32
	}{
		{
			timesig:  [2]uint8{1, 4},
			capacity: uint32(constants.TicksPerQuarter),
		},
		{
			timesig:  [2]uint8{4, 4},
			capacity: uint32(constants.TicksPerWhole),
		},
	} {
		g := NewWithT(t)

		bar := balafon.Bar{}
		bar.SetTimeSig(tc.timesig[0], tc.timesig[1])
		g.Expect(bar.Cap()).To(Equal(tc.capacity))
	}
}

func TestZeroDurationBar(t *testing.T) {
	for _, tc := range []struct {
		input string
		dur   time.Duration
	}{
		{
			input: ":time 1 1; :tempo 60; :bar 1 :program 1; :end; :play 1",
			dur:   0,
		},
		{
			input: ":time 1 1; :tempo 60; :assign c 60; :bar 1 c :end; :play 1",
			dur:   4 * time.Second,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Duration(60)).To(Equal(tc.dur))
		})
	}
}

func TestBarDurationMultiTrack(t *testing.T) {
	g := NewWithT(t)

	bar := balafon.Bar{
		Events: []balafon.Event{
			{
				Duration: uint32(constants.TicksPerQuarter),
			},
			{
				Duration: uint32(constants.TicksPerQuarter),
			},
		},
	}

	bar.SetTimeSig(1, 4)

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
		t.Run(fmt.Sprintf(":time %s", tc.timesig), func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			g.Expect(it.EvalString(fmt.Sprintf(":time %s", tc.timesig))).To(Succeed())
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
:time 1 1
:velocity 1
:channel 1
:end
	`

	// TODO
	it := balafon.New()
	err := it.EvalString(input)

	spew.Dump(it.Flush())
	_ = g
	_ = t
	_ = err
	// g.Expect(it.EvalString(input)).NotTo(Succeed())
}
