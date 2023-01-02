package interpreter_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestCommands(t *testing.T) {
	for _, tc := range []struct {
		input   string
		timesig [2]uint8
		msg     []byte
		dur     uint32
	}{
		{
			"assign c 60; c",
			[2]uint8{4, 4},
			midi.NoteOn(0, 60, constants.DefaultVelocity),
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"tempo 200",
			[2]uint8{4, 4},
			smf.MetaTempo(200),
			0,
		},
		{
			"timesig 1 4",
			[2]uint8{1, 4},
			nil,
			0,
		},
		{
			"channel 10; assign c 60; c",
			[2]uint8{4, 4},
			midi.NoteOn(10, 60, constants.DefaultVelocity),
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"velocity 30; assign c 60; c",
			[2]uint8{4, 4},
			midi.NoteOn(0, 60, 30),
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"program 0",
			[2]uint8{4, 4},
			midi.ProgramChange(0, 0),
			0,
		},
		{
			"control 1 2",
			[2]uint8{4, 4},
			midi.ControlChange(0, 1, 2),
			0,
		},
		{
			`assign c 60; bar "bar" timesig 1 4; c end; play "bar"`,
			[2]uint8{1, 4},
			midi.NoteOn(0, 60, constants.DefaultVelocity),
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"start",
			[2]uint8{4, 4},
			midi.Start(),
			0,
		},
		{
			"stop",
			[2]uint8{4, 4},
			midi.Stop(),
			0,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			song, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			bars := song.Bars()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].TimeSig).To(Equal(tc.timesig))

			if tc.msg == nil {
				g.Expect(bars[0].Events).To(BeEmpty())
			} else {
				g.Expect(bars[0].Events).To(HaveLen(1))
				g.Expect(bars[0].Events[0].Duration).To(BeEquivalentTo(tc.dur))
				g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(tc.msg))
			}
		})
	}
}

func TestUndefinedKey(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	_, err := it.Eval("k")
	g.Expect(err).To(HaveOccurred())
}

func TestNoteAlreadyAssigned(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	_, err := it.Eval("assign c 60")
	g.Expect(err).NotTo(HaveOccurred())

	_, err = it.Eval("assign c 61")
	g.Expect(err).To(HaveOccurred())
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

			song, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			bars := song.Bars()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(1))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, tc.key, constants.DefaultVelocity)))
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

			_, err := it.Eval(input)
			g.Expect(err).To(HaveOccurred())
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

			song, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			bars := song.Bars()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(1))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, 60, tc.velocity)))
		})
	}
}

func TestNoteLengths(t *testing.T) {
	for _, tc := range []struct {
		input string
		ticks smf.MetricTicks
	}{
		{
			input: "k", // Quarter note.
			ticks: constants.TicksPerQuarter,
		},
		{
			input: "k.", // Dotted quarter note, x1.5.
			ticks: constants.TicksPerQuarter * 3 / 2,
		},
		{
			input: "k..", // Double dotted quarter note, x1.75.
			ticks: constants.TicksPerQuarter * 7 / 4,
		},
		{
			input: "k...", // Triple dotted quarter note, x1.875.
			ticks: constants.TicksPerQuarter * 15 / 8,
		},
		{
			input: "k/3", // Triplet quarter note.
			ticks: constants.TicksPerQuarter * 2 / 3,
		},
		{
			input: "k/5", // Quintuplet quarter note.
			ticks: constants.TicksPerQuarter * 2 / 5,
		},
		{
			input: "k./3", // Dotted triplet quarter note == quarter note.
			ticks: constants.TicksPerQuarter,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			_, err := it.Eval("assign k 36")
			g.Expect(err).NotTo(HaveOccurred())

			song, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			bars := song.Bars()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(1))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, 36, constants.DefaultVelocity)))
			g.Expect(uint32(bars[0].Events[0].Duration) * 8).To(BeEquivalentTo(tc.ticks))
		})
	}
}

func TestBarScope(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	song, err := it.Eval(`
velocity 10

channel 1
assign c 60

channel 2
assign c 120

channel 1

bar "bar"
    velocity 20

    channel 2
    c
end

play "bar"

// back to channel 1.

c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := song.Bars()
	g.Expect(bars).To(HaveLen(2))
	g.Expect(bars[0].Events).To(HaveLen(1))
	g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(2, 120, 20)))

	g.Expect(bars[1].Events).To(HaveLen(1))
	g.Expect(bars[1].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(1, 60, 10)))
}

// // TODO: test bar length validation

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
