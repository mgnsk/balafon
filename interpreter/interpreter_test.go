package interpreter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mgnsk/gong/constants"
	"github.com/mgnsk/gong/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
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
			`:assign c 60; :bar "bar" :timesig 1 4; c :end; :play "bar"`,
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

			g.Expect(it.Eval(tc.input)).To(Succeed())

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

	g.Expect(it.Eval("k")).NotTo(Succeed())
}

func TestNoteAlreadyAssigned(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	g.Expect(it.Eval(":assign c 60")).To(Succeed())
	g.Expect(it.Eval(":assign c 61")).NotTo(Succeed())
}

func TestSharpFlatNote(t *testing.T) {
	for _, tc := range []struct {
		input string
		key   uint8
	}{
		{":assign c 60; c#", 61},
		{":assign c 60; c$", 59},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.Eval(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(2))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, tc.key, constants.DefaultVelocity)))
			g.Expect(bars[0].Events[1].Message).To(BeEquivalentTo(midi.NoteOff(0, tc.key)))
		})
	}
}

func TestSharpFlatNoteRange(t *testing.T) {
	for _, input := range []string{
		":assign c 127; c#",
		":assign c 0; c$",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.Eval(input)).NotTo(Succeed())
		})
	}
}

func TestAccentuatedAndGhostNote(t *testing.T) {
	for _, tc := range []struct {
		input    string
		velocity uint8
	}{
		{":velocity 100; :assign c 60; c^", 110},
		{":velocity 100; :assign c 60; c^^", 120},
		{":velocity 100; :assign c 60; c^^^", constants.MaxValue},
		{":velocity 20; :assign c 60; c)", 10},
		{":velocity 20; :assign c 60; c))", 1},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.Eval(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(2))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, 60, tc.velocity)))
			g.Expect(bars[0].Events[1].Message).To(BeEquivalentTo(midi.NoteOff(0, 60)))
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

			g.Expect(it.Eval(fmt.Sprintf(":tempo %d", tempo))).To(Succeed())
			g.Expect(it.Eval(":timesig 4 4")).To(Succeed())
			g.Expect(it.Eval(":assign k 36")).To(Succeed())
			g.Expect(it.Eval(tc.input)).To(Succeed())

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

func TestNotEmptyBar(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
:timesig 1 4

:bar "one"
	-
:end

:bar "two"
	:program 1
:end

:play "one"
:play "two"
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

		err := it.Eval(`
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

		err := it.Eval(`
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

	err := it.Eval(`
:assign c 60

:timesig 3 4

:bar "bar"
	:timesig 1 4
    c
:end

:play "bar"

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

	err := it.Eval(`
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

	g.Expect(it.Eval(":assign c 60")).To(Succeed())
	g.Expect(it.Eval(":timesig 4 4")).To(Succeed())
	g.Expect(it.Eval(":ccccc")).NotTo(Succeed())
	g.Expect(it.Eval("c")).To(Succeed())

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

:bar "test"
	:channel 1
	abcd
	:channel 2
	abcd
:end
:play "test"
`

	g.Expect(it.Eval(input)).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))
	g.Expect(bars[0].Duration(60)).To(Equal(4 * time.Second))
	g.Expect(bars[0].Events).To(HaveLen(1 + 16)) // 1 tempo, 8 note on, 8 note off

	keyPositions := map[uint8][]uint32{}
	for _, ev := range bars[0].Events[1:] { // skip first tempo msg
		var ch, key, vel uint8
		if ev.Message.GetNoteOn(&ch, &key, &vel) {
			keyPositions[key] = append(keyPositions[key], ev.Pos)
		}
	}

	g.Expect(keyPositions).To(HaveLen(4))
	for _, pos := range keyPositions {
		g.Expect(pos).To(HaveLen(2))
		g.Expect(pos[0]).To(Equal(pos[1]))
	}
}

func TestPendingGlobalCommands(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
:channel 2; :assign d 62
:channel 1; :assign c 60
:tempo 60
:timesig 1 4
:velocity 50
:program 1
:control 1 1

:bar "one"
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
:bar "two"
	:tempo 120
	c
:end

:play "one"
:play "two"

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

	err := it.Eval(`
:channel 1; :assign c 60
:tempo 120
:tempo 60
:timesig 1 4

// Tempo 60 4th rest == 1s.
-

:bar "two"
	:timesig 2 8
	c
:end

:bar "one"
	:tempo 120
	:timesig 2 8
	c
:end

:play "one"
:play "two"
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

	g.Expect(it.Eval(":assign k 36")).To(Succeed())
	g.Expect(it.Eval("k*")).To(Succeed())

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

func TestSyntaxNotAmbigous(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	g.Expect(it.Eval(`
:assign t 0
:assign e 1
:assign m 2
:assign p 3
:assign o 4
:bar "bar"
	:timesig 5 4
	tempo
:end
:play "bar"
	`)).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))
	g.Expect(bars[0].Events).To(HaveLen(10))
}

// var (
// 	testFile  []byte
// 	lineCount int
// )

// func init() {
// 	b, err := ioutil.ReadFile("../../examples/bonham")
// 	if err != nil {
// 		panic(err)
// 	}
// 	testFile = b
// 	lineCount = bytes.Count(testFile, []byte{'\n'})
// }

// // func BenchmarkInterpreter(b *testing.B) {
// // 	start := time.Now()

// // 	b.ReportAllocs()
// // 	b.ResetTimer()

// // 	var err error

// // 	for i := 0; i < b.N; i++ {
// // 		it := interpreter.New()
// // 		_, err = it.EvalAll(bytes.NewReader(testFile))
// // 	}

// // 	b.StopTimer()

// // 	if err != nil {
// // 		panic(err)
// // 	}

// // 	elapsed := time.Since(start)

// // 	linesPerNano := float64(b.N*lineCount) / float64(elapsed)

// // 	fmt.Printf("lines per second: %f\n", linesPerNano*float64(time.Second))
// // }
