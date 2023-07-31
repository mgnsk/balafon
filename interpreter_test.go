package balafon_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	balafon "github.com/mgnsk/balafon"
	"github.com/mgnsk/balafon/internal/constants"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"golang.org/x/exp/slices"
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
				midi.NoteOn(0, 60, constants.DefaultVelocity),
				midi.NoteOff(0, 60),
			},
		},
		{
			":tempo 200",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(200),
			},
		},
		{
			":timesig 1 4",
			[2]uint8{1, 4},
			nil, // Nil bar.
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
			":channel 10; :assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
				midi.NoteOn(10, 60, constants.DefaultVelocity),
				midi.NoteOff(10, 60),
			},
		},
		{
			":velocity 30; :assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
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
			`:assign c 60; :bar bar :timesig 1 4; c :end; :play bar`,
			[2]uint8{1, 4},
			[][]byte{
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
				g.Expect(bars[0].TimeSig).To(Equal(tc.timesig))
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
			{":assign c 125; c##", 127},
			{":assign c 125; c#$", 125},
			{":assign c 2; c$", 1},
			{":assign c 2; c$$", 0},
			{":assign c 2; c$#", 2},
		} {
			t.Run(tc.input, func(t *testing.T) {
				g := NewWithT(t)

				it := balafon.New()

				g.Expect(it.EvalString(tc.input)).To(Succeed())

				bars := it.Flush()
				g.Expect(bars).To(HaveLen(1))
				g.Expect(bars[0].Events).To(HaveLen(2))
				g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, tc.key, constants.DefaultVelocity)))
				g.Expect(bars[0].Events[1].Message).To(BeEquivalentTo(midi.NoteOff(0, tc.key)))
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, tc := range []struct {
			input string
		}{
			{":assign c 125; c###"},
			{":assign c 2; c$$$"},
		} {
			t.Run(tc.input, func(t *testing.T) {
				g := NewWithT(t)

				it := balafon.New()

				g.Expect(it.EvalString(tc.input)).NotTo(Succeed())
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
			g.Expect(bars[0].Events).To(HaveLen(2))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, 60, tc.velocity)))
			g.Expect(bars[0].Events[1].Message).To(BeEquivalentTo(midi.NoteOff(0, 60)))
		})
	}
}

func TestStaccatoNote(t *testing.T) {
	for _, tc := range []struct {
		input string
		offAt uint32
	}{
		{":timesig 1 4; :assign c 60; c", uint32(constants.TicksPerQuarter)},
		{":timesig 1 4; :assign c 60; c`", uint32(constants.TicksPerQuarter / 2)},
		{":timesig 1 4; :assign c 60; c``", uint32(constants.TicksPerQuarter / 4)},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := balafon.New()

			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].TimeSig).To(Equal([2]uint8{1, 4}))
			g.Expect(bars[0].Events[0].Duration).To(Equal(tc.offAt))
			g.Expect(bars[0].Events[1].Pos).To(Equal(tc.offAt))
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
			g.Expect(it.EvalString(":timesig 4 4")).To(Succeed())
			g.Expect(it.EvalString(":assign k 36")).To(Succeed())
			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))

			g.Expect(bars[0].TimeSig).To(Equal([2]uint8{4, 4}))
			g.Expect(bars[0].Events).To(HaveExactElements(
				HaveField("Message", smf.MetaTempo(float64(tempo))),
				SatisfyAll(
					HaveField("Pos", uint32(0)),
					HaveField("Duration", uint32(tc.offAt)),
				),
				SatisfyAll(
					HaveField("Pos", uint32(tc.offAt)),
					HaveField("Duration", uint32(0)),
				),
			))
		})
	}
}

func TestSilence(t *testing.T) {
	t.Run("end of bar", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		err := it.EvalString(`
:timesig 4 4
:assign c 60
c
`)
		g.Expect(err).NotTo(HaveOccurred())

		bars := it.Flush()

		g.Expect(bars).To(HaveLen(1))
		g.Expect(bars[0].Cap()).To(Equal(uint32(constants.TicksPerWhole)))

		// TODO: fill with rests?
		g.Expect(bars[0].String()).To(Equal(`timesig: 4/4
events:
pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
`))
	})

	t.Run("beginning of bar", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		err := it.EvalString(`
:timesig 4 4
:assign c 60
---c
`)
		g.Expect(err).NotTo(HaveOccurred())

		bars := it.Flush()

		g.Expect(bars).To(HaveLen(1))
		g.Expect(bars[0].Cap()).To(Equal(uint32(constants.TicksPerWhole)))
		g.Expect(bars[0].String()).To(Equal(`timesig: 4/4
events:
pos: 0 dur: 960 note: - message: UnknownType
pos: 960 dur: 960 note: - message: UnknownType
pos: 1920 dur: 960 note: - message: UnknownType
pos: 2880 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
pos: 3840 dur: 0 message: NoteOff channel: 0 key: 60
`))
	})
}

func TestTimeSignature(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:assign c 60

:timesig 3 4

:bar bar
	:timesig 1 4
    c
:end

:play bar

/* Expect time signature to be restored to 3 4 in next bar. */
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(2))
	g.Expect(bars[0].TimeSig).To(Equal([2]uint8{1, 4}))
	g.Expect(bars[0].Cap()).To(BeEquivalentTo(constants.TicksPerQuarter))

	g.Expect(bars[1].TimeSig).To(Equal([2]uint8{3, 4}))
	g.Expect(bars[1].Cap()).To(BeEquivalentTo(3 * constants.TicksPerQuarter))
}

func TestBarTooLong(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:assign c 60
:tempo 60
/* Default timesig 4 4. */

ccccc
`)
	g.Expect(err).To(HaveOccurred())
}

func TestFlushSkipsTooLongBar(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	g.Expect(it.EvalString(":assign c 60")).To(Succeed())
	g.Expect(it.EvalString(":timesig 4 4")).To(Succeed())
	g.Expect(it.EvalString(":ccccc")).NotTo(Succeed())
	g.Expect(it.EvalString("c")).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))

	g.Expect(bars[0].String()).To(Equal(`timesig: 4/4
events:
pos: 0 dur: 960 note: c message: NoteOn channel: 0 key: 60 velocity: 100
pos: 960 dur: 0 message: NoteOff channel: 0 key: 60
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
:timesig 4 4

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
	g.Expect(bars[0].Events).To(HaveLen(1 + 16)) // 1 tempo, 8 note on, 8 note off

	pos := make([]uint32, 16)
	for i, ev := range bars[0].Events[1:] { // skip first tempo msg
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
:timesig 1 4
:velocity 50
:program 1
:control 1 1

:bar one
	:tempo 120
	:timesig 2 8
	:velocity 25

	:program 2
	:control 1 2

	/* on channel 1: */
	c

	:channel 2
	d
:end

/* timesig 1 4 */
:bar two
	:tempo 120
	c
:end

:play one
:play two

/*
Channel is 1, timesig 1 4, velocity 50 but tempo is 120.
Only timesig, velocity and channel are local to bars.
tempo, program, control, start, stop are global commands.
*/
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(3))

	g.Expect(bars[0].String()).To(Equal(`timesig: 2/8
events:
pos: 0 dur: 0 message: MetaTempo bpm: 60.00
pos: 0 dur: 0 message: ProgramChange channel: 1 program: 1
pos: 0 dur: 0 message: ControlChange channel: 1 controller: 1 value: 1
pos: 0 dur: 0 message: MetaText text: "timesig 1 4"
pos: 0 dur: 0 message: MetaTempo bpm: 120.00
pos: 0 dur: 0 message: ProgramChange channel: 1 program: 2
pos: 0 dur: 0 message: ControlChange channel: 1 controller: 1 value: 2
pos: 0 dur: 0 message: MetaText text: "on channel 1:"
pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 25
pos: 0 dur: 960 note: d message: NoteOn channel: 2 key: 62 velocity: 25
pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
pos: 960 dur: 0 message: NoteOff channel: 2 key: 62
`))

	g.Expect(bars[1].String()).To(Equal(`timesig: 1/4
events:
pos: 0 dur: 0 message: MetaTempo bpm: 120.00
pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 50
pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
`))

	g.Expect(bars[2].String()).To(Equal(`timesig: 1/4
events:
pos: 0 dur: 0 message: MetaText text: "Channel is 1, timesig 1 4, velocity 50 but tempo is 120.\nOnly timesig, velocity and channel are local to bars.\ntempo, program, control, start, stop are global commands."
pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 50
pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
`))
}

func TestTempoIsGlobal(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()

	err := it.EvalString(`
:channel 1; :assign c 60
:tempo 120
:tempo 60
:timesig 1 4

/* Tempo 60 4th rest == 1s. */
-

:bar two
	:timesig 2 8
	c
:end

:bar one
	:tempo 120
	:timesig 2 8
	c
:end

:play one
:play two
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()

	g.Expect(bars).To(HaveLen(4))

	g.Expect(bars[0].String()).To(Equal(`timesig: 1/4
events:
pos: 0 dur: 0 message: MetaTempo bpm: 120.00
pos: 0 dur: 0 message: MetaTempo bpm: 60.00
pos: 0 dur: 0 message: MetaText text: "Tempo 60 4th rest == 1s."
pos: 0 dur: 960 note: - message: UnknownType
`))

	g.Expect(bars[1].String()).To(Equal(`timesig: 2/8
events:
pos: 0 dur: 0 message: MetaTempo bpm: 120.00
pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 100
pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
`))

	g.Expect(bars[2].String()).To(Equal(`timesig: 2/8
events:
pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 100
pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
`))

	g.Expect(bars[3].String()).To(Equal(`timesig: 1/4
events:
pos: 0 dur: 960 note: c message: NoteOn channel: 1 key: 60 velocity: 100
pos: 960 dur: 0 message: NoteOff channel: 1 key: 60
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
	g.Expect(bars[0].String()).To(Equal(`timesig: 4/4
events:
pos: 0 dur: 960 note: k* message: NoteOn channel: 0 key: 36 velocity: 100
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
