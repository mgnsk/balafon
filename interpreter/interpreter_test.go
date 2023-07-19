package interpreter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mgnsk/balafon/internal/constants"
	"github.com/mgnsk/balafon/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"golang.org/x/exp/slices"
)

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

			it := interpreter.New()

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

	it := interpreter.New()

	g.Expect(it.EvalString("k")).NotTo(Succeed())
}

func TestNoteAlreadyAssigned(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

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

				it := interpreter.New()

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

				it := interpreter.New()

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

			it := interpreter.New()

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

			it := interpreter.New()

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

			it := interpreter.New()

			tempo := 60

			g.Expect(it.EvalString(fmt.Sprintf(":tempo %d", tempo))).To(Succeed())
			g.Expect(it.EvalString(":timesig 4 4")).To(Succeed())
			g.Expect(it.EvalString(":assign k 36")).To(Succeed())
			g.Expect(it.EvalString(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].TimeSig).To(Equal([2]uint8{4, 4}))
			g.Expect(bars[0].Events).To(ConsistOf(
				interpreter.Event{
					Pos:      0,
					Duration: 0,
					Message:  smf.MetaTempo(float64(tempo)),
				},
				interpreter.Event{
					Pos:      0,
					Duration: tc.offAt,
					Message:  smf.Message(midi.NoteOn(0, 36, constants.DefaultVelocity)),
				},
				interpreter.Event{
					Pos:      tc.offAt,
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(0, 36)),
				},
			))
		})
	}
}

func TestZeroDurationBarCollapse(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.EvalString(`
:timesig 1 4

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
-
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()

	g.Expect(bars).To(HaveLen(2))
	g.Expect(bars[0].TimeSig).To(Equal([2]uint8{1, 4}))
	g.Expect(bars[0].Events).To(HaveLen(0))

	g.Expect(bars[1].TimeSig).To(Equal([2]uint8{1, 4}))
	g.Expect(bars[1].Events).To(HaveLen(1))
	g.Expect(bars[1].Events[0].Message).To(BeEquivalentTo(midi.ProgramChange(0, 1)))
}

func TestSilence(t *testing.T) {
	t.Run("end of bar", func(t *testing.T) {
		g := NewWithT(t)

		it := interpreter.New()

		err := it.EvalString(`
:timesig 4 4
:assign c 60
c
`)
		g.Expect(err).NotTo(HaveOccurred())

		bars := it.Flush()

		g.Expect(bars).To(HaveLen(1))
		g.Expect(bars[0].Cap()).To(Equal(uint32(constants.TicksPerWhole)))
		g.Expect(bars[0].Events).To(ConsistOf(
			interpreter.Event{
				Message:  smf.Message(midi.NoteOn(0, 60, constants.DefaultVelocity)),
				Pos:      0,
				Duration: uint32(constants.TicksPerQuarter),
			},
			interpreter.Event{
				Message:  smf.Message(midi.NoteOff(0, 60)),
				Pos:      uint32(constants.TicksPerQuarter),
				Duration: 0,
			},
		))
	})

	t.Run("beginning of bar", func(t *testing.T) {
		g := NewWithT(t)

		it := interpreter.New()

		err := it.EvalString(`
:timesig 4 4
:assign c 60
---c
`)
		g.Expect(err).NotTo(HaveOccurred())

		bars := it.Flush()

		g.Expect(bars).To(HaveLen(1))
		g.Expect(bars[0].Cap()).To(Equal(uint32(constants.TicksPerWhole)))
		g.Expect(bars[0].Events).To(ConsistOf(
			interpreter.Event{
				Message:  smf.Message(midi.NoteOn(0, 60, constants.DefaultVelocity)),
				Pos:      uint32(3 * constants.TicksPerQuarter),
				Duration: uint32(constants.TicksPerQuarter),
			},
			interpreter.Event{
				Message:  smf.Message(midi.NoteOff(0, 60)),
				Pos:      uint32(constants.TicksPerWhole),
				Duration: 0,
			},
		))
	})
}

func TestTimeSignature(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.EvalString(`
:assign c 60

:timesig 3 4

:bar bar
	:timesig 1 4
    c
:end

:play bar

// Expect time signature to be restored to 3 4 in next bar.
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

	it := interpreter.New()

	err := it.EvalString(`
:assign c 60
:tempo 60
// Default timesig 4 4.

ccccc
`)
	g.Expect(err).To(HaveOccurred())
}

func TestFlushSkipsTooLongBar(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	g.Expect(it.EvalString(":assign c 60")).To(Succeed())
	g.Expect(it.EvalString(":timesig 4 4")).To(Succeed())
	g.Expect(it.EvalString(":ccccc")).NotTo(Succeed())
	g.Expect(it.EvalString("c")).To(Succeed())

	bars := it.Flush()

	g.Expect(bars).To(ConsistOf(&interpreter.Bar{
		TimeSig: [2]uint8{4, 4},
		Events: []interpreter.Event{
			{
				Message:  smf.Message(midi.NoteOn(0, 60, constants.DefaultVelocity)),
				Pos:      0,
				Duration: uint32(constants.TicksPerQuarter),
			},
			{
				Message:  smf.Message(midi.NoteOff(0, 60)),
				Pos:      uint32(constants.TicksPerQuarter),
				Duration: 0,
			},
		},
	}))
}

func TestMultiTrackNotesAreSortedPairs(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

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

	it := interpreter.New()

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

	// on channel 1:
	c

	:channel 2
	d
:end

// timesig 1 4
:bar two
	:tempo 120
	c
:end

:play one
:play two

// Channel is 1, timesig 1 4, velocity 50 but tempo is 120.
// Only timesig, velocity and channel are local to bars.
// tempo, program, control, start, stop are global commands.
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()

	g.Expect(bars).To(ConsistOf(
		&interpreter.Bar{
			TimeSig: [2]uint8{2, 8},
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(60)},
				{Message: smf.Message(midi.ProgramChange(1, 1))},
				{Message: smf.Message(midi.ControlChange(1, 1, 1))},
				{Message: smf.MetaTempo(120)},
				{Message: smf.Message(midi.ProgramChange(1, 2))},
				{Message: smf.Message(midi.ControlChange(1, 1, 2))},
				{
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, 25)),
				},
				{
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(2, 62, 25)),
				},
				{
					Pos:      uint32(constants.TicksPerQuarter),
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(1, 60)),
				},
				{
					Pos:      uint32(constants.TicksPerQuarter),
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(2, 62)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Events: []interpreter.Event{
				{
					Message: smf.MetaTempo(120),
				},
				{
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, 50)),
				},
				{
					Pos:      uint32(constants.TicksPerQuarter),
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(1, 60)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Events: []interpreter.Event{
				{
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, 50)),
				},
				{
					Pos:      uint32(constants.TicksPerQuarter),
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(1, 60)),
				},
			},
		},
	))
}

func TestTempoIsGlobal(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.EvalString(`
:channel 1; :assign c 60
:tempo 120
:tempo 60
:timesig 1 4

// Tempo 60 4th rest == 1s.
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

	g.Expect(bars).To(ConsistOf(
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(120)},
				{Message: smf.MetaTempo(60)},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{2, 8},
			Events: []interpreter.Event{
				{
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, constants.DefaultVelocity)),
				},
				{
					Pos:      uint32(constants.TicksPerQuarter),
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(1, 60)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{2, 8},
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(120)},
				{
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, constants.DefaultVelocity)),
				},
				{
					Pos:      uint32(constants.TicksPerQuarter),
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(1, 60)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Events: []interpreter.Event{
				{
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, constants.DefaultVelocity)),
				},
				{
					Pos:      uint32(constants.TicksPerQuarter),
					Duration: 0,
					Message:  smf.Message(midi.NoteOff(1, 60)),
				},
			},
		},
	))
}

func TestLetRing(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	g.Expect(it.EvalString(":assign k 36")).To(Succeed())
	g.Expect(it.EvalString("k*")).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))
	g.Expect(bars[0].Events).To(ConsistOf(
		interpreter.Event{
			Message:  smf.Message(midi.NoteOn(0, 36, constants.DefaultVelocity)),
			Pos:      0,
			Duration: uint32(constants.TicksPerQuarter),
		},
		// No note off event.
	))
}
