package interpreter_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func init() {
	format.UseStringerRepresentation = true
}

func TestCommands(t *testing.T) {
	for _, tc := range []struct {
		input   string
		msg     []byte
		timesig [2]uint8
		dur     uint32
	}{
		{
			"assign c 60; c",
			midi.NoteOn(0, 60, constants.DefaultVelocity),
			[2]uint8{4, 4},
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"tempo 200",
			smf.MetaTempo(200),
			[2]uint8{4, 4},
			0,
		},
		{
			"timesig 1 4",
			smf.MetaMeter(1, 4),
			[2]uint8{1, 4},
			0,
		},
		{
			"channel 10; assign c 60; c",
			midi.NoteOn(10, 60, constants.DefaultVelocity),
			[2]uint8{4, 4},
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"velocity 30; assign c 60; c",
			midi.NoteOn(0, 60, 30),
			[2]uint8{4, 4},
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"program 0",
			midi.ProgramChange(0, 0),
			[2]uint8{4, 4},
			0,
		},
		{
			"control 1 2",
			midi.ControlChange(0, 1, 2),
			[2]uint8{4, 4},
			0,
		},
		{
			`bar "bar" { assign c 60; c }; play "bar"`,
			midi.NoteOn(0, 60, constants.DefaultVelocity),
			[2]uint8{4, 4},
			constants.TicksPerQuarter.Ticks32th(),
		},
		{
			"start",
			midi.Start(),
			[2]uint8{4, 4},
			0,
		},
		{
			"stop",
			midi.Stop(),
			[2]uint8{4, 4},
			0,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			s, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(s.Bars()).To(HaveLen(1))
			bar := s.Bars()[0]
			g.Expect(bar.Events).To(HaveLen(1))
			g.Expect(bar.Events[0].Message).To(BeEquivalentTo(tc.msg))
			g.Expect(bar.TimeSig).To(Equal(tc.timesig))

			var d uint32
			for _, ev := range bar.Events {
				d += uint32(ev.Duration)
			}

			g.Expect(d).To(Equal(tc.dur))
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

			s, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(s.Bars()).To(HaveLen(1))
			bar := s.Bars()[0]
			g.Expect(bar.Events).To(HaveLen(1))
			g.Expect(bar.Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, tc.key, constants.DefaultVelocity)))
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
		{"velocity 100; assign c 60; c^", 127},
		{"velocity 100; assign c 60; c)", 50},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			s, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(s.Bars()).To(HaveLen(1))
			bar := s.Bars()[0]
			g.Expect(bar.Events).To(HaveLen(1))
			g.Expect(bar.Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, 60, tc.velocity)))
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
			input: "k...", // Triplet dotted quarter note, x1.875.
			ticks: constants.TicksPerQuarter * 15 / 8,
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

			s, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())

			g.Expect(s.Bars()).To(HaveLen(1))
			bar := s.Bars()[0]
			g.Expect(bar.Events).To(HaveLen(1))
			g.Expect(bar.Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, 36, constants.DefaultVelocity)))

			g.Expect(uint32(bar.Events[0].Duration) * 8).To(BeEquivalentTo(tc.ticks))
		})
	}
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

// // func TestEvalAll(t *testing.T) {
// // 	g := NewWithT(t)

// // 	it := interpreter.New()

// // 	input := `
// // 		velocity 100
// // 		channel 1
// // 		assign x 36
// // 		channel 2
// // 		assign x 38

// // 		bar "verse"
// // 		channel 1
// // 		x
// // 		channel 2
// // 		x
// // 		end
// //         play "verse"
// //     `
// // 	// TODO bar not filled error

// // 	song, err := it.EvalAll(strings.NewReader(input))
// // 	g.Expect(err).NotTo(HaveOccurred())

// // 	spew.Dump(song)

// // }

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
