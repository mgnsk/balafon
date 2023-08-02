package balafon_test

import (
	"testing"
	"time"

	"github.com/mgnsk/balafon"
	"github.com/mgnsk/balafon/internal/constants"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/testdrv"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestSequencerTiming(t *testing.T) {
	for _, tc := range []struct {
		input    string
		absTicks uint32
	}{
		{
			input:    ":time 1 4; :tempo 60; :assign c 60; c",
			absTicks: uint32(constants.TicksPerQuarter),
		},
		{
			input:    ":time 2 8; :tempo 60; :assign c 60; c",
			absTicks: uint32(constants.TicksPerQuarter),
		},
		{
			input:    ":time 2 4; :tempo 120; :assign c 60; c2",
			absTicks: uint32(2 * constants.TicksPerQuarter),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()
			g.Expect(it.EvalString(tc.input)).To(Succeed())

			s := balafon.NewSequencer()
			s.AddBars(it.Flush()...)

			sm := s.Flush()
			g.Expect(sm).To(HaveLen(3))

			g.Expect(sm[0].Message.Type()).To(Equal(smf.MetaTempoMsg))
			g.Expect(sm[0].AbsTicks).To(Equal(uint32(0)))
			g.Expect(sm[0].AbsNanoseconds).To(Equal(int64(0)))

			g.Expect(sm[1].Message.Type()).To(Equal(midi.NoteOnMsg))
			g.Expect(sm[1].AbsTicks).To(Equal(uint32(0)))
			g.Expect(sm[1].AbsNanoseconds).To(Equal(int64(0)))

			g.Expect(sm[2].Message.Type()).To(Equal(midi.NoteOffMsg))
			g.Expect(sm[2].AbsTicks).To(Equal(tc.absTicks))
			g.Expect(sm[2].AbsNanoseconds).To(Equal(time.Second.Nanoseconds()))
		})
	}
}

func TestSequencerMultiTrackTiming(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	input := `
:channel 1
:assign x 42
:channel 2
:assign k 36
:tempo 60
:time 4 4
:bar test
	:channel 1
	xxxx
	:channel 2
	kkkk
:end
:play test
`

	g.Expect(it.EvalString(input)).To(Succeed())

	s := balafon.NewSequencer()
	s.AddBars(it.Flush()...)

	sm := s.Flush()
	g.Expect(sm).To(HaveLen(17))

	g.Expect(sm[15].Message.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(sm[15].AbsTicks).To(Equal(uint32(constants.TicksPerWhole)))
	g.Expect(sm[15].AbsNanoseconds).To(Equal(4 * time.Second.Nanoseconds()))

	g.Expect(sm[16].Message.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(sm[16].AbsTicks).To(Equal(uint32(constants.TicksPerWhole)))
	g.Expect(sm[16].AbsNanoseconds).To(Equal(4 * time.Second.Nanoseconds()))
}

func TestSilenceBetweenBars(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	input := `
:assign x 42
:tempo 60
:time 2 4

:bar one
	x-
:end

:bar two
	-x
:end

:play one
:play two
`

	g.Expect(it.EvalString(input)).To(Succeed())

	s := balafon.NewSequencer()
	s.AddBars(it.Flush()...)

	sm := s.Flush()

	g.Expect(sm).To(HaveExactElements(
		balafon.TrackEvent{
			Message: smf.MetaTempo(60),
		},
		balafon.TrackEvent{
			Message: smf.Message(midi.NoteOn(0, 42, constants.DefaultVelocity)),
		},
		balafon.TrackEvent{
			Message:        smf.Message(midi.NoteOff(0, 42)),
			AbsTicks:       uint32(constants.TicksPerQuarter),
			AbsNanoseconds: 1 * time.Second.Nanoseconds(),
		},
		balafon.TrackEvent{
			AbsTicks:       uint32(1 * constants.TicksPerQuarter),
			AbsNanoseconds: 1 * time.Second.Nanoseconds(),
		},
		balafon.TrackEvent{
			AbsTicks:       uint32(2 * constants.TicksPerQuarter),
			AbsNanoseconds: 2 * time.Second.Nanoseconds(),
		},
		balafon.TrackEvent{
			Message:        smf.Message(midi.NoteOn(0, 42, constants.DefaultVelocity)),
			AbsTicks:       uint32(3 * constants.TicksPerQuarter),
			AbsNanoseconds: 3 * time.Second.Nanoseconds(),
		},
		balafon.TrackEvent{
			Message:        smf.Message(midi.NoteOff(0, 42)),
			AbsTicks:       uint32(4 * constants.TicksPerQuarter),
			AbsNanoseconds: 4 * time.Second.Nanoseconds(),
		},
	))
}

func TestTempoChange(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	input := `
:assign x 42

:bar one
	:time 1 4
	:tempo 60
	x
:end

:bar two
	:time 2 4
	:tempo 120
	xx
:end

:play one
:play two
`

	g.Expect(it.EvalString(input)).To(Succeed())

	s := balafon.NewSequencer()
	s.AddBars(it.Flush()...)

	sm := s.Flush()
	g.Expect(sm).To(HaveLen(8))

	g.Expect(sm[7]).To(Equal(balafon.TrackEvent{
		Message:        smf.Message(midi.NoteOff(0, 42)),
		AbsTicks:       uint32(3 * constants.TicksPerQuarter),
		AbsNanoseconds: 2 * time.Second.Nanoseconds(),
	}))
}

func TestZeroDurationBarCollapse(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:tempo 60
:time 1 4

:bar one
	-
:end

:bar two
	:program 1
:end

:bar three
	-
:end

:play one
:play two
:play three
-
`)
	g.Expect(err).NotTo(HaveOccurred())

	s := balafon.NewSequencer()
	s.AddBars(it.Flush()...)

	sm := s.Flush()

	g.Expect(sm).To(HaveExactElements(
		balafon.TrackEvent{
			Message: smf.MetaTempo(60),
		},
		balafon.TrackEvent{},
		balafon.TrackEvent{
			Message:        smf.Message(midi.ProgramChange(0, 1)),
			AbsTicks:       uint32(constants.TicksPerQuarter),
			AbsNanoseconds: 1 * time.Second.Nanoseconds(),
		},
		balafon.TrackEvent{
			AbsTicks:       uint32(constants.TicksPerQuarter),
			AbsNanoseconds: 1 * time.Second.Nanoseconds(),
		},
		balafon.TrackEvent{
			AbsTicks:       uint32(2 * constants.TicksPerQuarter),
			AbsNanoseconds: 2 * time.Second.Nanoseconds(),
		},
	))
}
