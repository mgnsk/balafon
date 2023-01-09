package interpreter_test

import (
	"fmt"
	"sort"
	"testing"

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
			"assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(120),
				midi.NoteOn(0, 60, constants.DefaultVelocity),
			},
		},
		{
			"tempo 200",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(200),
			},
		},
		{
			"timesig 1 4",
			[2]uint8{1, 4},
			nil, // Nil bar.
		},
		{
			"velocity 10",
			[2]uint8{4, 4},
			nil, // Nil bar.
		},
		{
			"channel 10",
			[2]uint8{4, 4},
			nil, // Nil bar.
		},
		{
			"channel 10; assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(120),
				midi.NoteOn(10, 60, constants.DefaultVelocity),
			},
		},
		{
			"velocity 30; assign c 60; c",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(120),
				midi.NoteOn(0, 60, 30),
			},
		},
		{
			"program 0",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(120),
				midi.ProgramChange(0, 0),
			},
		},
		{
			"control 1 2",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(120),
				midi.ControlChange(0, 1, 2),
			},
		},
		{
			`assign c 60; bar "bar" timesig 1 4; c end; play "bar"`,
			[2]uint8{1, 4},
			[][]byte{
				smf.MetaTempo(120),
				midi.NoteOn(0, 60, constants.DefaultVelocity),
			},
		},
		{
			"start",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(120),
				midi.Start(),
			},
		},
		{
			"stop",
			[2]uint8{4, 4},
			[][]byte{
				smf.MetaTempo(120),
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

	g.Expect(it.Eval("assign c 60")).To(Succeed())
	g.Expect(it.Eval("assign c 61")).NotTo(Succeed())
}

func TestSharpFlatNote(t *testing.T) {
	for _, tc := range []struct {
		input string
		key   uint8
	}{
		{"assign c 60; c#", 61},
		{"assign c 60; c$", 59},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.Eval(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(2))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(smf.MetaTempo(120)))
			g.Expect(bars[0].Events[1].Message).To(BeEquivalentTo(midi.NoteOn(0, tc.key, constants.DefaultVelocity)))
		})
	}
}

func TestSharpFlatNoteRange(t *testing.T) {
	for _, input := range []string{
		"assign c 127; c#",
		"assign c 0; c$",
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
		{"velocity 100; assign c 60; c^", 110},
		{"velocity 100; assign c 60; c^^", 120},
		{"velocity 100; assign c 60; c^^^", constants.MaxValue},
		{"velocity 20; assign c 60; c)", 10},
		{"velocity 20; assign c 60; c))", 1},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.Eval(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(2))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(smf.MetaTempo(120)))
			g.Expect(bars[0].Events[1].Message).To(BeEquivalentTo(midi.NoteOn(0, 60, tc.velocity)))
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
			input: "k...", // Triplet dotted quarter note, x1.875.
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

			g.Expect(it.Eval(fmt.Sprintf("tempo %d", tempo))).To(Succeed())
			g.Expect(it.Eval("timesig 4 4")).To(Succeed())
			g.Expect(it.Eval("assign k 36")).To(Succeed())
			g.Expect(it.Eval(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].TimeSig).To(Equal([2]uint8{4, 4}))

			events := bars[0].Events
			g.Expect(events).To(ConsistOf(
				interpreter.Event{
					Channel:  0,
					Pos:      0,
					Duration: 0,
					Message:  smf.MetaTempo(float64(tempo)),
				},
				interpreter.Event{
					Channel:  0,
					Pos:      0,
					Duration: tc.offAt,
					Message:  smf.Message(midi.NoteOn(0, 36, constants.DefaultVelocity)),
				},
			))
		})
	}
}

func TestNotEmptyBar(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
timesig 1 4

bar "one"
	-
end

bar "two"
	program 1
end

play "one"
play "two"
-
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()

	g.Expect(bars).To(HaveLen(2))
	g.Expect(bars[0].TimeSig).To(Equal([2]uint8{1, 4}))
	g.Expect(bars[0].Events).To(HaveLen(1))
	g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(smf.MetaTempo(120)))

	g.Expect(bars[1].TimeSig).To(Equal([2]uint8{1, 4}))
	g.Expect(bars[1].Events).To(HaveLen(2))
	g.Expect(bars[1].Events[0].Message).To(BeEquivalentTo(smf.MetaTempo(120)))
	g.Expect(bars[1].Events[1].Message).To(BeEquivalentTo(midi.ProgramChange(0, 1)))
}

func TestTimeSignature(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
assign c 60

timesig 3 4

bar "bar"
    timesig 1 4
    c
end

play "bar"

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
assign c 60
tempo 60
// Default timesig 4 4.

ccccc
`)
	g.Expect(err).To(HaveOccurred())
}

func TestFlushSkipsTooLongBar(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	g.Expect(it.Eval("assign c 60")).To(Succeed())
	g.Expect(it.Eval("timesig 4 4")).To(Succeed())
	g.Expect(it.Eval("ccccc")).NotTo(Succeed())
	g.Expect(it.Eval("c")).To(Succeed())

	bars := it.Flush()

	g.Expect(bars).To(ConsistOf(&interpreter.Bar{
		TimeSig: [2]uint8{4, 4},
		Tempo:   120,
		Events: []interpreter.Event{
			{
				Message: smf.MetaTempo(120),
			},
			{
				Duration: uint32(constants.TicksPerQuarter),
				Message:  smf.Message(midi.NoteOn(0, 60, constants.DefaultVelocity)),
			},
		},
	}))
}

func TestMultiTrackNotesAreSorted(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	input := `
channel 1
assign x 42
channel 2
assign k 36
timesig 4 4
bar "test"
	channel 1
	xxxx
	channel 2
	kkkk
end
play "test"
`

	g.Expect(it.Eval(input)).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))
	g.Expect(bars[0].Events).To(HaveLen(9))

	positions := make([]int, len(bars[0].Events))
	for i, ev := range bars[0].Events {
		positions[i] = int(ev.Pos)
	}

	g.Expect(sort.IntsAreSorted(positions)).To(BeTrue())
}

func TestPendingGlobalCommands(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
channel 2; assign d 62
channel 1; assign c 60
tempo 60
timesig 1 4
velocity 50
program 1
control 1 1

bar "one"
	tempo 120
	timesig 2 8
	velocity 25

	program 2
	control 1 2

	// on channel 1:
	c

	channel 2
	d
end

// timesig 1 4
bar "two"
	tempo 120
	c
end

play "one"
play "two"

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
			Tempo:   120,
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(60)},
				{Message: smf.Message(midi.ProgramChange(1, 1))},
				{Message: smf.Message(midi.ControlChange(1, 1, 1))},
				{Message: smf.MetaTempo(120)},
				{Message: smf.Message(midi.ProgramChange(1, 2))},
				{Message: smf.Message(midi.ControlChange(1, 1, 2))},
				{
					Channel:  1,
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, 25)),
				},
				{
					Channel:  2,
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(2, 62, 25)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Tempo:   120,
			Events: []interpreter.Event{
				{
					Message: smf.MetaTempo(120),
				},
				{
					Channel:  1,
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, 50)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Tempo:   120,
			Events: []interpreter.Event{
				{
					Message: smf.MetaTempo(120),
				},
				{
					Channel:  1,
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, 50)),
				},
			},
		},
	))
}

func TestTempoIsGlobal(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
channel 1; assign c 60
tempo 120
tempo 60
timesig 1 4

// Tempo 60 4th rest == 1s.
-

bar "two"
	timesig 2 8
	c
end

bar "one"
	tempo 120
	timesig 2 8
	c
end

play "one"
play "two"
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush()

	g.Expect(bars).To(ConsistOf(
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Tempo:   60,
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(120)},
				{Message: smf.MetaTempo(60)},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{2, 8},
			Tempo:   120,
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(120)},
				{
					Channel:  1,
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, constants.DefaultVelocity)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{2, 8},
			Tempo:   120,
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(120)},
				{
					Channel:  1,
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, constants.DefaultVelocity)),
				},
			},
		},
		&interpreter.Bar{
			TimeSig: [2]uint8{1, 4},
			Tempo:   120,
			Events: []interpreter.Event{
				{Message: smf.MetaTempo(120)},
				{
					Channel:  1,
					Pos:      0,
					Duration: uint32(constants.TicksPerQuarter),
					Message:  smf.Message(midi.NoteOn(1, 60, constants.DefaultVelocity)),
				},
			},
		},
	))
}

// func TestLetRing(t *testing.T) {
// 	g := NewWithT(t)

// 	it := interpreter.New()

// 	evalExpectNil(g, it, `assign k 36`)

// 	ms, err := it.Eval(`k*`)
// 	g.Expect(err).NotTo(HaveOccurred())
// 	g.Expect(ms).To(HaveLen(1))

// 	g.Expect(ms[0].Pos).To(Equal(uint8(0)))
// 	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 36, 127))))

// 	// Expect the ringing note to be turned off.
// 	// TODO

// 	ms, err = it.Eval(`k`)
// 	g.Expect(err).NotTo(HaveOccurred())
// 	g.Expect(ms).To(HaveLen(1))

// 	g.Expect(ms[0].Pos).To(Equal(uint8(0)))
// 	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 36, 127))))

// 	// g.Expect(ms[2].Pos).To(Equal(uint32(constants.TicksPerQuarter * 2)))
// 	// g.Expect(ms[2].Message).To(Equal(smf.Message(midi.NoteOff(0, 36))))
// }

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
