package balafon_test

import (
	"errors"
	"fmt"
	"slices"
	"testing"
	"time"

	"github.com/mgnsk/balafon"
	"github.com/mgnsk/balafon/internal/constants"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestParseError(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalFile("testdata/parse_error.bal")
	g.Expect(err).To(HaveOccurred())

	var perr *balafon.ParseError
	g.Expect(errors.As(err, &perr)).To(BeTrue())
	g.Expect(perr.Error()).To(HavePrefix("testdata/parse_error.bal:1:1: error:"))
}

func TestEvalError(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalFile("testdata/eval_error.bal")
	g.Expect(err).To(HaveOccurred())

	var perr *balafon.EvalError
	g.Expect(errors.As(err, &perr)).To(BeTrue())
	g.Expect(perr.Error()).To(HavePrefix("testdata/eval_error.bal:2:1: error:"))
}

func TestCommands(t *testing.T) {
	for _, tc := range []struct {
		input    string
		timesig  [2]uint8
		messages [][]byte
	}{
		{
			":assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaMeter(4, 4),
				midi.NoteOn(0, 60, constants.DefaultVelocity),
				midi.NoteOff(0, 60),
			},
		},
		{
			":tempo 200",
			[2]uint8{4, 4},
			[][]byte{
				// Note: missing MetaMeter since there are no notes.
				smf.MetaTempo(200),
			},
		},
		{
			":time 1 4",
			[2]uint8{1, 4},
			[][]byte{
				smf.MetaMeter(1, 4),
			},
		},
		{
			":velocity 10",
			[2]uint8{4, 4},
			nil, // Nil bar.
		},
		{
			":channel 10",
			[2]uint8{4, 4},
			nil, // Nil bar.
		},
		{
			":voice 4",
			[2]uint8{4, 4},
			nil, // Nil bar.
		},
		{
			":channel 10; :assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaMeter(4, 4),
				midi.NoteOn(9, 60, constants.DefaultVelocity),
				midi.NoteOff(9, 60),
			},
		},
		{
			":velocity 30; :assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaMeter(4, 4),
				midi.NoteOn(0, 60, 30),
				midi.NoteOff(0, 60),
			},
		},
		{
			":program 0",
			[2]uint8{4, 4},
			[][]byte{
				midi.ProgramChange(0, 0),
			},
		},
		{
			":control 1 2",
			[2]uint8{4, 4},
			[][]byte{
				midi.ControlChange(0, 1, 2),
			},
		},
		{
			`:assign c 60; :bar bar :time 1 4; c :end; :play bar`,
			[2]uint8{1, 4},
			[][]byte{
				smf.MetaMeter(1, 4),
				midi.NoteOn(0, 60, constants.DefaultVelocity),
				midi.NoteOff(0, 60),
			},
		},
		{
			":start",
			[2]uint8{4, 4},
			[][]byte{
				midi.Start(),
			},
		},
		{
			":stop",
			[2]uint8{4, 4},
			[][]byte{
				midi.Stop(),
			},
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()

			switch len(tc.messages) {
			case 0:
				g.Expect(bars).To(HaveLen(0))
			default:
				g.Expect(bars).To(HaveLen(1))
				g.Expect(bars[0].Cap()).To(Equal(uint32(tc.timesig[0]) * (uint32(constants.TicksPerWhole) / uint32(tc.timesig[1]))))
				g.Expect(bars[0].Events).To(HaveLen(len(tc.messages)))
				for i, msg := range tc.messages {
					g.Expect(bars[0].Events[i].Message).To(BeEquivalentTo(msg))
				}
			}
		})
	}
}

func TestUndefinedKey(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	g.Expect(it.EvalString("k")).NotTo(Succeed())
}

func TestNoteAlreadyAssigned(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	g.Expect(it.EvalString(":assign c 60")).To(Succeed())
	g.Expect(it.EvalString(":assign c 61")).NotTo(Succeed())
}

func TestSharpFlatNote(t *testing.T) {
	t.Run("success cases", func(t *testing.T) {
		for _, tc := range []struct {
			input string
			key   uint8
		}{
			{":assign c 125; c#", 126},
			{":assign c 2; c$", 1},
		} {
			t.Run(tc.input, func(t *testing.T) {
				g := NewWithT(t)

				it := balafon.New()

				g.Expect(it.EvalString(tc.input)).To(Succeed())

				bars := it.Flush()
				g.Expect(bars).To(HaveLen(1))

				_, key, _, ok := FindNote(bars[0])
				g.Expect(ok).To(BeTrue())
				g.Expect(key).To(Equal(tc.key))
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, tc := range []struct {
			input string
		}{
			{":assign c 127; c#"},
			{":assign c 0; c$"},
		} {
			t.Run(tc.input, func(t *testing.T) {
				g := NewWithT(t)

				it := balafon.New()

				err := it.EvalString(tc.input)
				g.Expect(err).To(HaveOccurred())

				var perr *balafon.EvalError
				g.Expect(errors.As(err, &perr)).To(BeTrue())
				g.Expect(perr.Error()).To(ContainSubstring("note key must be in range"))
			})
		}
	})
}

func TestAccentuatedAndGhostNote(t *testing.T) {
	for _, tc := range []struct {
		input    string
		velocity uint8
	}{
		{":velocity 110; :assign c 60; c", 110},
		{":velocity 110; :assign c 60; c>", 115},
		{":velocity 110; :assign c 60; c>>", 120},
		{":velocity 110; :assign c 60; c>>>", 125},
		{":velocity 110; :assign c 60; c>>>>", constants.MaxValue},
		{":velocity 110; :assign c 60; c^", 120},
		{":velocity 110; :assign c 60; c^^", constants.MaxValue},
		{":velocity 110; :assign c 60; c^^^", constants.MaxValue},
		{":velocity 110; :assign c 60; c^>", 125},
		{":velocity 110; :assign c 60; c^>>", constants.MaxValue},
		{":velocity 10; :assign c 60; c", 10},
		{":velocity 10; :assign c 60; c)", 5},
		{":velocity 10; :assign c 60; c))", 0},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))

			_, _, velocity, ok := FindNote(bars[0])
			g.Expect(ok).To(BeTrue())
			g.Expect(velocity).To(Equal(velocity))
		})
	}
}

func TestStaccatoNote(t *testing.T) {
	for _, tc := range []struct {
		input string
		offAt uint32
	}{
		{":time 1 4; :assign c 60; c", uint32(constants.TicksPerQuarter)},
		{":time 1 4; :assign c 60; c`", uint32(constants.TicksPerQuarter / 2)},
		{":time 1 4; :assign c 60; c``", uint32(constants.TicksPerQuarter / 4)},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Cap()).To(Equal(uint32(1) * (uint32(constants.TicksPerWhole) / uint32(4))))
			g.Expect(bars[0].Events[1].Duration).To(Equal(tc.offAt))
			g.Expect(bars[0].Events[2].Pos).To(Equal(tc.offAt))
		})
	}
}

func TestNoteLengths(t *testing.T) {
	for _, tc := range []struct {
		input string
		offAt uint32
	}{
		{
			input: "k", // Quarter note.
			offAt: uint32(constants.TicksPerQuarter),
		},
		{
			input: "k.", // Dotted quarter note, x1.5.
			offAt: uint32(constants.TicksPerQuarter * 3 / 2),
		},
		{
			input: "k..", // Double dotted quarter note, x1.75.
			offAt: uint32(constants.TicksPerQuarter * 7 / 4),
		},
		{
			input: "k...", // Triple dotted quarter note, x1.875.
			offAt: uint32(constants.TicksPerQuarter * 15 / 8),
		},
		{
			input: "k/5", // Quintuplet quarter note.
			offAt: uint32(constants.TicksPerQuarter * 2 / 5),
		},
		{
			input: "k./3", // Dotted triplet quarter note == quarter note.
			offAt: uint32(constants.TicksPerQuarter),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			tempo := 60

			g.Expect(it.EvalString(fmt.Sprintf(":tempo %d", tempo))).To(Succeed())
			g.Expect(it.EvalString(":time 4 4")).To(Succeed())
			g.Expect(it.EvalString(":assign k 36")).To(Succeed())
			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))

			g.Expect(bars[0].Cap()).To(Equal(uint32(4) * (uint32(constants.TicksPerWhole) / uint32(4))))

			events := bars[0].Events
			g.Expect(events).To(HaveLen(4))

			g.Expect(events[0]).To(HaveField("Message", smf.MetaTempo(float64(tempo))))
			g.Expect(events[1]).To(HaveField("Message", smf.MetaMeter(4, 4)))
			g.Expect(events[2]).To(SatisfyAll(
				HaveField("Pos", uint32(0)),
				HaveField("Duration", uint32(tc.offAt)),
			))
			g.Expect(events[3]).To(SatisfyAll(
				HaveField("Pos", uint32(tc.offAt)),
				HaveField("Duration", uint32(0)),
			))
		})
	}
}

func TestSilence(t *testing.T) {
	t.Run("end of bar", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		err := it.EvalString(`
:time 4 4
:assign c 60
c
`)
		g.Expect(err).NotTo(HaveOccurred())

		bars := it.Flush()

		g.Expect(bars).To(HaveLen(1))
		g.Expect(bars[0].Cap()).To(Equal(uint32(constants.TicksPerWhole)))

		// TODO: fill with rests?
		g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))
	})

	t.Run("beginning of bar", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		err := it.EvalString(`
:time 4 4
:assign c 60
---c
`)
		g.Expect(err).NotTo(HaveOccurred())

		bars := it.Flush()

		g.Expect(bars).To(HaveLen(1))
		g.Expect(bars[0].Cap()).To(Equal(uint32(constants.TicksPerWhole)))
		g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 note: - message: UnknownType
track: 1 pos: 960 dur: 960 note: - message: UnknownType
track: 1 pos: 1920 dur: 960 note: - message: UnknownType
track: 1 pos: 2880 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 1 pos: 3840 dur: 0 message: NoteOff channel: 0 key: 60
`))
	})
}

func TestTimeSignature(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:assign c 60

:time 3 4

:bar bar
	:time 1 4
    c
:end

:play bar

/* Expect time signature to be restored to 3 4 in next bar. */
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(2))

	g.Expect(bars[0].Cap()).To(Equal(uint32(1) * (uint32(constants.TicksPerWhole) / uint32(4))))
	g.Expect(bars[0].Cap()).To(BeEquivalentTo(constants.TicksPerQuarter))

	g.Expect(bars[1].Cap()).To(Equal(uint32(3) * (uint32(constants.TicksPerWhole) / uint32(4))))
	g.Expect(bars[1].Cap()).To(BeEquivalentTo(3 * constants.TicksPerQuarter))
}

func TestBarTooLong(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:assign c 60
:tempo 60
/* Default time 4 4. */

ccccc
`)
	g.Expect(err).To(HaveOccurred())
}

func TestFlushSkipsTooLongBar(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	g.Expect(it.EvalString(":assign c 60")).To(Succeed())
	g.Expect(it.EvalString(":time 4 4")).To(Succeed())
	g.Expect(it.EvalString("ccccc")).NotTo(Succeed())
	g.Expect(it.EvalString("c")).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))

	g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))
}

func TestMultiTrackNotesAreSortedPairs(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	input := `
:channel 1
:assign a 0
:assign b 1
:assign c 2
:assign d 3

:channel 2
:assign a 0
:assign b 1
:assign c 2
:assign d 3

:tempo 60
:time 4 4

:bar test
	:channel 1
	abcd
	:channel 2
	abcd
:end
:play test
`

	g.Expect(it.EvalString(input)).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))
	g.Expect(bars[0].Duration(60)).To(Equal(4 * time.Second))
	g.Expect(bars[0].Events).To(HaveLen(4 + 16)) // 2 tempo, 2 timesig, 8 note on, 8 note off

	pos := make([]uint32, 16)
	for i, ev := range bars[0].Events[4:] { // skip meta messages
		pos[i] = ev.Pos
	}

	g.Expect(slices.IsSorted(pos)).To(BeTrue())

	for i := 0; i < len(pos)-1; i += 2 {
		g.Expect(pos[i]).To(Equal(pos[i+1]))
	}
}

func TestPendingGlobalCommands(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:channel 2; :assign d 62
:channel 1; :assign c 60
:tempo 60
:time 1 4
:velocity 50
:program 1
:control 1 1

:bar one
	:key Cm
	:tempo 120
	:time 2 8
	:velocity 25

	:program 2
	:control 1 2

	/* on channel 1: */
	:voice 1
	c

	:channel 2
	:voice 2
	d
:end

:bar two
	/* time is 1 4 */
	c
:end

:play one
:play two

/*
Channel is 1, voice is 1, velocity 50
but time is 1 4, tempo is 120 and key is Cm.
Only velocity, channel and voice are local to bars.
time, tempo, key, program, control, start, stop
are global commands.
*/
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(3))

	// Note: time 1 4 does not exist here. It is overridden by 2 8.
	// time 1 4 is reintroduced in the next bar because
	// bar captures the current timesig at the moment of
	// :bar evaluation instead of :play evaluation.
	g.Expect(bars[0].String()).To(Equal(`time: 2/8
events:
track: 1 pos: 0 dur: 0 message: MetaTempo bpm: 60.00
track: 2 pos: 0 dur: 0 message: MetaTempo bpm: 60.00
track: 1 pos: 0 dur: 0 message: ProgramChange channel: 0 program: 1
track: 1 pos: 0 dur: 0 message: ControlChange channel: 0 controller: 1 value: 1
track: 1 pos: 0 dur: 0 message: MetaKeySig key: CMin
track: 1 pos: 0 dur: 0 message: MetaTempo bpm: 120.00
track: 2 pos: 0 dur: 0 message: MetaTempo bpm: 120.00
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 2/8
track: 2 pos: 0 dur: 0 message: MetaTimeSig meter: 2/8
track: 1 pos: 0 dur: 0 message: ProgramChange channel: 0 program: 2
track: 1 pos: 0 dur: 0 message: ControlChange channel: 0 controller: 1 value: 2
track: 1 pos: 0 dur: 0 message: MetaText text: "on channel 1:"
track: 1 pos: 0 dur: 960 voice: 1 note: c message: NoteOn channel: 0 key: 60 velocity: 25
track: 2 pos: 0 dur: 960 voice: 2 note: d message: NoteOn channel: 1 key: 62 velocity: 25
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
track: 2 pos: 960 dur: 0 message: NoteOff channel: 1 key: 62
`))

	// Note: time is 1 4 now.
	g.Expect(bars[1].String()).To(Equal(`time: 1/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 1/4
track: 2 pos: 0 dur: 0 message: MetaTimeSig meter: 1/4
track: 1 pos: 0 dur: 0 message: MetaText text: "time is 1 4"
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 50
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))

	g.Expect(bars[2].String()).To(Equal(`time: 1/4
events:
track: 1 pos: 0 dur: 0 message: MetaText text: "Channel is 1, voice is 1, velocity 50\nbut time is 1 4, tempo is 120 and key is Cm.\nOnly velocity, channel and voice are local to bars.\ntime, tempo, key, program, control, start, stop\nare global commands."
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 1/4
track: 2 pos: 0 dur: 0 message: MetaTimeSig meter: 1/4
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 50
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))
}

func TestGlobalCommands(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:channel 1; :assign c 60
:tempo 60
:time 1 4

/* Tempo 60 4th rest == 1s. */
-

:bar two
	c
:end

:bar one
	:tempo 120
	:time 2 8
	c
:end

:play one
:play two
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()

	g.Expect(bars).To(HaveLen(4))

	g.Expect(bars[0].String()).To(Equal(`time: 1/4
events:
track: 1 pos: 0 dur: 0 message: MetaTempo bpm: 60.00
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 1/4
track: 1 pos: 0 dur: 0 message: MetaText text: "Tempo 60 4th rest == 1s."
track: 1 pos: 0 dur: 960 note: - message: UnknownType
`))

	g.Expect(bars[1].String()).To(Equal(`time: 2/8
events:
track: 1 pos: 0 dur: 0 message: MetaTempo bpm: 120.00
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 2/8
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))

	g.Expect(bars[2].String()).To(Equal(`time: 1/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 1/4
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))

	g.Expect(bars[3].String()).To(Equal(`time: 1/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 1/4
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))
}

func TestLetRing(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	g.Expect(it.EvalString(":assign k 36")).To(Succeed())
	g.Expect(it.EvalString("k*")).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))
	// No note off event.
	g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 note: k* message: NoteOn channel: 0 key: 36 velocity: 100
`))
}

func TestCommandsForbiddenInBar(t *testing.T) {
	for _, input := range []string{
		":assign c 60",
		`:bar inner :start :end`,
		`:play test`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			it := balafon.New()

			err := it.EvalString(fmt.Sprintf(`:bar outer %s; :end`, input))
			g.Expect(err).To(HaveOccurred())

			var perr *balafon.EvalError
			g.Expect(errors.As(err, &perr)).To(BeTrue())
			g.Expect(perr.Error()).To(HaveSuffix("not allowed in bar"))
		})
	}
}

func TestVoice(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	g.Expect(it.EvalString(":assign k 36")).To(Succeed())
	g.Expect(it.EvalString(":voice 1; k")).To(Succeed())
	g.Expect(it.EvalString(":voice 2; k")).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(2))

	g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 voice: 1 note: k message: NoteOn channel: 0 key: 36 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 36
`))

	g.Expect(bars[1].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 voice: 2 note: k message: NoteOn channel: 0 key: 36 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 36
`))
}

func TestChannelHumanValue(t *testing.T) {
	for _, tc := range []struct {
		humanCh uint8
		midiCh  uint8
	}{
		{1, 0},
		{16, 15},
	} {
		t.Run(fmt.Sprintf(":channel %d", tc.humanCh), func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			g.Expect(it.EvalString(fmt.Sprintf(":channel %d", tc.humanCh))).To(Succeed())
			g.Expect(it.EvalString(":assign c 60")).To(Succeed())
			g.Expect(it.EvalString("c")).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))

			ch, _, _, ok := FindNote(bars[0])
			g.Expect(ok).To(BeTrue())

			g.Expect(ch).To(Equal(tc.midiCh))
		})
	}
}

func TestDefaultChannel(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	g.Expect(it.EvalString(":assign c 60")).To(Succeed())
	g.Expect(it.EvalString("c")).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))

	ch, _, _, ok := FindNote(bars[0])
	g.Expect(ok).To(BeTrue())
	g.Expect(ch).To(Equal(uint8(0)))
}

func TestMetaEventsOnAllChannels(t *testing.T) {
	t.Run("single default channel", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		g.Expect(it.EvalString(":assign c 60")).To(Succeed())
		g.Expect(it.EvalString(":tempo 100")).To(Succeed())
		g.Expect(it.EvalString("c")).To(Succeed())

		bars := it.Flush()
		g.Expect(bars).To(HaveLen(1))

		g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTempo bpm: 100.00
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))
	})

	t.Run("default channel", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		g.Expect(it.EvalString("/* channel 1 */")).To(Succeed())
		g.Expect(it.EvalString(":channel 2")).To(Succeed())
		g.Expect(it.EvalString(":assign c 60")).To(Succeed())
		g.Expect(it.EvalString("c")).To(Succeed())

		bars := it.Flush()
		g.Expect(bars).To(HaveLen(1))

		g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaText text: "channel 1"
track: 2 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 2 pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 100
track: 2 pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
`))
	})

	t.Run("multiple channels", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		g.Expect(it.EvalString(`
:channel 1; :assign c 60
:channel 2; :assign c 60
:tempo 100
:bar one
	:channel 1; c
	:channel 2; c
:end
:play one
`)).To(Succeed())

		bars := it.Flush()
		g.Expect(bars).To(HaveLen(1))

		g.Expect(bars[0].String()).To(Equal(`time: 4/4
events:
track: 1 pos: 0 dur: 0 message: MetaTempo bpm: 100.00
track: 2 pos: 0 dur: 0 message: MetaTempo bpm: 100.00
track: 1 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 2 pos: 0 dur: 0 message: MetaTimeSig meter: 4/4
track: 1 pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
track: 2 pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 100
track: 1 pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
track: 2 pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
`))
	})
}
