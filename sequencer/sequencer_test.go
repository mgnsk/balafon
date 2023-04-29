package sequencer_test

import (
	"testing"
	"time"

	"github.com/mgnsk/balafon/constants"
	"github.com/mgnsk/balafon/interpreter"
	"github.com/mgnsk/balafon/sequencer"
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
			input:    ":timesig 1 4; :tempo 60; :assign c 60; c",
			absTicks: uint32(constants.TicksPerQuarter),
		},
		{
			input:    ":timesig 2 8; :tempo 60; :assign c 60; c",
			absTicks: uint32(constants.TicksPerQuarter),
		},
		{
			input:    ":timesig 2 4; :tempo 120; :assign c 60; c2",
			absTicks: uint32(2 * constants.TicksPerQuarter),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()
			g.Expect(it.EvalString(tc.input)).To(Succeed())

			s := sequencer.New()
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

	it := interpreter.New()

	input := `
:channel 1
:assign x 42
:channel 2
:assign k 36
:tempo 60
:timesig 4 4
:bar test
	:channel 1
	xxxx
	:channel 2
	kkkk
:end
:play test
`

	g.Expect(it.EvalString(input)).To(Succeed())

	s := sequencer.New()
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

	it := interpreter.New()

	input := `
:assign x 42
:tempo 60
:timesig 2 4

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

	s := sequencer.New()
	s.AddBars(it.Flush()...)

	sm := s.Flush()
	g.Expect(sm).To(HaveLen(5))

	g.Expect(sm[1].Message.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(sm[1].AbsTicks).To(Equal(uint32(0)))
	g.Expect(sm[1].AbsNanoseconds).To(Equal(int64(0)))

	g.Expect(sm[2].Message.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(sm[2].AbsTicks).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(sm[2].AbsNanoseconds).To(Equal(time.Second.Nanoseconds()))

	g.Expect(sm[3].Message.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(sm[3].AbsTicks).To(Equal(uint32(3 * constants.TicksPerQuarter)))
	g.Expect(sm[3].AbsNanoseconds).To(Equal(3 * time.Second.Nanoseconds()))

	g.Expect(sm[4].Message.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(sm[4].AbsTicks).To(Equal(uint32(4 * constants.TicksPerQuarter)))
	g.Expect(sm[4].AbsNanoseconds).To(Equal(4 * time.Second.Nanoseconds()))
}

func TestTempoChange(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	input := `
:assign x 42

:bar one
	:timesig 1 4
	:tempo 60
	x
:end

:bar two
	:timesig 2 4
	:tempo 120
	xx
:end

:play one
:play two
`

	g.Expect(it.EvalString(input)).To(Succeed())

	s := sequencer.New()
	s.AddBars(it.Flush()...)

	sm := s.Flush()
	g.Expect(sm).To(HaveLen(8))

	g.Expect(sm[7]).To(Equal(sequencer.TrackEvent{
		Message:        smf.Message(midi.NoteOff(0, 42)),
		AbsTicks:       uint32(3 * constants.TicksPerQuarter),
		AbsNanoseconds: 2 * time.Second.Nanoseconds(),
	}))
}
