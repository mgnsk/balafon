package interpreter_test

import (
	"bytes"
	"fmt"
	"testing"
	"time"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/sequencer"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestCommands(t *testing.T) {
	for _, tc := range []struct {
		input   string
		timesig [2]uint8
		msg     []byte
		len32th uint8
	}{
		{
			"assign c 60; c",
			[2]uint8{4, 4},
			midi.NoteOn(0, 60, constants.DefaultVelocity),
			8,
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
			8,
		},
		{
			"velocity 30; assign c 60; c",
			[2]uint8{4, 4},
			midi.NoteOn(0, 60, 30),
			8,
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
			8,
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

			g.Expect(it.Eval(tc.input)).To(Succeed())

			song := it.Flush()

			if tc.msg == nil {
				g.Expect(song).To(BeNil())
			} else {
				g.Expect(song).NotTo(BeNil())

				bars := song.Bars()
				g.Expect(bars).To(HaveLen(1))
				g.Expect(bars[0].TimeSig).To(Equal(tc.timesig))
				g.Expect(bars[0].Events).To(HaveLen(1))
				g.Expect(bars[0].Events[0].Duration).To(BeEquivalentTo(tc.len32th))
				g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(tc.msg))
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

			bars := it.Flush().Bars()
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

			bars := it.Flush().Bars()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Events).To(HaveLen(1))
			g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(0, 60, tc.velocity)))
		})
	}
}

func ticksFrom32th(len32th uint8) uint32 {
	return uint32(len32th) * constants.TicksPerQuarter.Ticks32th()
}

func TestNoteLengths(t *testing.T) {
	for _, tc := range []struct {
		input   string
		len32th uint8
	}{
		{
			input:   "k", // Quarter note.
			len32th: 8,
		},
		{
			input:   "k.", // Dotted quarter note, x1.5.
			len32th: 12,
		},
		{
			input:   "k..", // Double dotted quarter note, x1.75.
			len32th: 14,
		},
		{
			input:   "k...", // Triple dotted quarter note, x1.875.
			len32th: 15,
		},
		{
			input:   "k/3", // Triplet quarter note.
			len32th: 5,     // TODO: precision
		},
		{
			input:   "k/5", // Quintuplet quarter note.
			len32th: 3,     // TODO: precision
		},
		{
			input:   "k./3", // Dotted triplet quarter note == quarter note.
			len32th: 8,
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

			song := it.Flush()

			t.Run("bar event duration", func(t *testing.T) {
				g := NewWithT(t)

				bars := song.Bars()
				g.Expect(bars).To(HaveLen(1))
				g.Expect(bars[0].TimeSig).To(Equal([2]uint8{4, 4}))

				events := bars[0].Events
				g.Expect(events).To(ConsistOf(
					&sequencer.Event{
						TrackNo:  0,
						Pos:      0,
						Duration: 0,
						Message:  smf.MetaTempo(float64(tempo)),
					},
					&sequencer.Event{
						TrackNo:  0,
						Pos:      0,
						Duration: tc.len32th,
						Message:  smf.Message(midi.NoteOn(0, 36, constants.DefaultVelocity)),
					},
				))
			})

			t.Run("SMF event duration", func(t *testing.T) {
				g := NewWithT(t)

				sm := song.ToSMF1()

				var buf bytes.Buffer

				_, err := sm.WriteTo(&buf)
				g.Expect(err).NotTo(HaveOccurred())

				rd := smf.ReadTracksFrom(&buf)
				g.Expect(rd.Error()).NotTo(HaveOccurred())

				var events []smf.TrackEvent

				rd.
					Only(midi.NoteOnMsg, midi.NoteOffMsg).
					Do(func(ev smf.TrackEvent) {
						events = append(events, ev)
					})

				g.Expect(events).To(ConsistOf(
					smf.TrackEvent{
						Event: smf.Event{
							Delta:   0,
							Message: smf.Message(midi.NoteOn(0, 36, constants.DefaultVelocity)),
						},
						TrackNo:         1,
						AbsTicks:        0,
						AbsMicroSeconds: 0,
					},
					smf.TrackEvent{
						Event: smf.Event{
							Delta:   ticksFrom32th(tc.len32th),
							Message: smf.Message(midi.NoteOff(0, 36)),
						},
						TrackNo:         1,
						AbsTicks:        int64(ticksFrom32th(tc.len32th)),
						AbsMicroSeconds: constants.TicksPerQuarter.Duration(float64(tempo), ticksFrom32th(tc.len32th)).Microseconds(),
					},
				))
			})
		})
	}
}

func TestBarScope(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
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

	bars := it.Flush().Bars()
	g.Expect(bars).To(HaveLen(2))
	g.Expect(bars[0].Events).To(HaveLen(1))
	g.Expect(bars[0].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(2, 120, 20)))

	g.Expect(bars[1].Events).To(HaveLen(1))
	g.Expect(bars[1].Events[0].Message).To(BeEquivalentTo(midi.NoteOn(1, 60, 10)))
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

// Expect time signature to be 1 4 in second bar.

c
`)
	g.Expect(err).NotTo(HaveOccurred())

	bars := it.Flush().Bars()
	g.Expect(bars).To(HaveLen(2))
	g.Expect(bars[0].TimeSig).To(Equal([2]uint8{1, 4}))
	g.Expect(bars[0].Len()).To(BeEquivalentTo(8))

	g.Expect(bars[1].TimeSig).To(Equal([2]uint8{1, 4}))
	g.Expect(bars[0].Len()).To(BeEquivalentTo(8))
}

func TestSMFBarAutoFill(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
assign c 60

tempo 60
timesig 4 4

c
c
`)
	g.Expect(err).NotTo(HaveOccurred())

	song := it.Flush()

	g.Expect(song.Bars()).To(HaveLen(2))

	sm := song.ToSMF1()

	var buf bytes.Buffer

	_, err = sm.WriteTo(&buf)
	g.Expect(err).NotTo(HaveOccurred())

	rd := smf.ReadTracksFrom(&buf)
	g.Expect(rd.Error()).NotTo(HaveOccurred())

	var events []smf.TrackEvent

	rd.
		Only(midi.NoteOnMsg, midi.NoteOffMsg).
		Do(func(ev smf.TrackEvent) {
			events = append(events, ev)
		})

	// To assert sanity.
	g.Expect(events).To(ConsistOf(
		smf.TrackEvent{
			Event: smf.Event{
				Delta:   0 * uint32(constants.TicksPerQuarter),
				Message: smf.Message(midi.NoteOn(0, 60, constants.DefaultVelocity)),
			},
			TrackNo:         1,
			AbsTicks:        0 * int64(constants.TicksPerQuarter),
			AbsMicroSeconds: (0 * time.Second).Microseconds(),
		},
		smf.TrackEvent{
			Event: smf.Event{
				Delta:   1 * uint32(constants.TicksPerQuarter),
				Message: smf.Message(midi.NoteOff(0, 60)),
			},
			TrackNo:         1,
			AbsTicks:        1 * int64(constants.TicksPerQuarter),
			AbsMicroSeconds: (1 * time.Second).Microseconds(),
		},
		smf.TrackEvent{
			Event: smf.Event{
				Delta:   3 * uint32(constants.TicksPerQuarter),
				Message: smf.Message(midi.NoteOn(0, 60, constants.DefaultVelocity)),
			},
			TrackNo:         1,
			AbsTicks:        4 * int64(constants.TicksPerQuarter),
			AbsMicroSeconds: (4 * time.Second).Microseconds(),
		},
		smf.TrackEvent{
			Event: smf.Event{
				Delta:   1 * uint32(constants.TicksPerQuarter),
				Message: smf.Message(midi.NoteOff(0, 60)),
			},
			TrackNo:         1,
			AbsTicks:        5 * int64(constants.TicksPerQuarter),
			AbsMicroSeconds: (5 * time.Second).Microseconds(),
		},
	))
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

func TestBarTempo(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	err := it.Eval(`
channel 1
assign c 60

channel 2
assign d 62

tempo 60
timesig 1 4

bar "test"
	channel 1
	c
end

// Calling play should flush the buffered tempo command.

play "test"
d
`)
	g.Expect(err).NotTo(HaveOccurred())

	song := it.Flush()

	g.Expect(song.Bars()).To(ConsistOf(
		&sequencer.Bar{
			Number:  0,
			TimeSig: [2]uint8{1, 4},
			Events: sequencer.Events{
				&sequencer.Event{
					TrackNo:  0,
					Pos:      0,
					Duration: 0,
					Message:  smf.MetaTempo(60),
				},
				&sequencer.Event{
					TrackNo:  1,
					Pos:      0,
					Duration: 8,
					Message:  smf.Message(midi.NoteOn(1, 60, constants.DefaultVelocity)),
				},
			},
		},
		&sequencer.Bar{
			Number:  1,
			TimeSig: [2]uint8{1, 4},
			Events: sequencer.Events{
				&sequencer.Event{
					TrackNo:  2,
					Pos:      0,
					Duration: 8,
					Message:  smf.Message(midi.NoteOn(2, 62, constants.DefaultVelocity)),
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
