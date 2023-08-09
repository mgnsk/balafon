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

			p := balafon.New()
			g.Expect(p.EvalString(tc.input)).To(Succeed())

			s := balafon.NewSequencer()
			s.AddBars(p.Flush()...)

			sm := s.Flush()
			g.Expect(sm).To(HaveLen(4))

			g.Expect(sm[0].Message.Type()).To(Equal(smf.MetaTimeSigMsg))
			g.Expect(sm[0].AbsTicks).To(Equal(uint32(0)))
			g.Expect(sm[0].AbsNanoseconds).To(Equal(int64(0)))

			g.Expect(sm[1].Message.Type()).To(Equal(smf.MetaTempoMsg))
			g.Expect(sm[1].AbsTicks).To(Equal(uint32(0)))
			g.Expect(sm[1].AbsNanoseconds).To(Equal(int64(0)))

			g.Expect(sm[2].Message.Type()).To(Equal(midi.NoteOnMsg))
			g.Expect(sm[2].AbsTicks).To(Equal(uint32(0)))
			g.Expect(sm[2].AbsNanoseconds).To(Equal(int64(0)))

			g.Expect(sm[3].Message.Type()).To(Equal(midi.NoteOffMsg))
			g.Expect(sm[3].AbsTicks).To(Equal(tc.absTicks))
			g.Expect(sm[3].AbsNanoseconds).To(Equal(time.Second.Nanoseconds()))
		})
	}
}

func TestSequencerMultiTrackTiming(t *testing.T) {
	g := NewWithT(t)

	p := balafon.New()

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

	g.Expect(p.EvalString(input)).To(Succeed())

	s := balafon.NewSequencer()
	s.AddBars(p.Flush()...)

	sm := s.Flush()
	g.Expect(sm).To(HaveLen(20))

	g.Expect(sm[18].Message.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(sm[18].AbsTicks).To(Equal(uint32(constants.TicksPerWhole)))
	g.Expect(sm[18].AbsNanoseconds).To(Equal(4 * time.Second.Nanoseconds()))

	g.Expect(sm[19].Message.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(sm[19].AbsTicks).To(Equal(uint32(constants.TicksPerWhole)))
	g.Expect(sm[19].AbsNanoseconds).To(Equal(4 * time.Second.Nanoseconds()))
}

func TestSilenceBetweenBars(t *testing.T) {
	g := NewWithT(t)

	p := balafon.New()

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

	g.Expect(p.EvalString(input)).To(Succeed())

	s := balafon.NewSequencer()
	s.AddBars(p.Flush()...)

	sm := s.Flush()
	g.Expect(sm.String()).To(Equal(`pos: 0 ns: 0 message: MetaTempo bpm: 60.00
pos: 0 ns: 0 message: MetaTimeSig meter: 2/4
pos: 0 ns: 0 message: NoteOn channel: 0 key: 42 velocity: 100
pos: 960 ns: 1000000000 message: NoteOff channel: 0 key: 42
pos: 960 ns: 1000000000 message: UnknownType
pos: 1920 ns: 2000000000 message: MetaTimeSig meter: 2/4
pos: 1920 ns: 2000000000 message: UnknownType
pos: 2880 ns: 3000000000 message: NoteOn channel: 0 key: 42 velocity: 100
pos: 3840 ns: 4000000000 message: NoteOff channel: 0 key: 42
`))
}

func TestTempoChange(t *testing.T) {
	g := NewWithT(t)

	p := balafon.New()

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

	g.Expect(p.EvalString(input)).To(Succeed())

	s := balafon.NewSequencer()
	s.AddBars(p.Flush()...)

	sm := s.Flush()
	g.Expect(sm).To(HaveLen(10))

	g.Expect(sm[9]).To(Equal(balafon.TrackEvent{
		Message:        smf.Message(midi.NoteOff(0, 42)),
		AbsTicks:       uint32(3 * constants.TicksPerQuarter),
		AbsNanoseconds: 2 * time.Second.Nanoseconds(),
	}))
}

func TestZeroDurationBarCollapse(t *testing.T) {
	g := NewWithT(t)

	p := balafon.New()

	err := p.EvalString(`
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
	s.AddBars(p.Flush()...)

	sm := s.Flush()
	g.Expect(sm.String()).To(Equal(`pos: 0 ns: 0 message: MetaTempo bpm: 60.00
pos: 0 ns: 0 message: MetaTimeSig meter: 1/4
pos: 0 ns: 0 message: UnknownType
pos: 960 ns: 1000000000 message: ProgramChange channel: 0 program: 1
pos: 960 ns: 1000000000 message: MetaTimeSig meter: 1/4
pos: 960 ns: 1000000000 message: UnknownType
pos: 1920 ns: 2000000000 message: MetaTimeSig meter: 1/4
pos: 1920 ns: 2000000000 message: UnknownType
`))
}
