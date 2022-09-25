package interpreter_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestProgramChangeCommand(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("program 0")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.ProgramChange(0, 0))))
}

func TestControlChangeCommand(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("control 0 1")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.ControlChange(0, 0, 1))))
}

func TestStartCommand(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("start")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.Start())))
}

func TestStopCommand(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("stop")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.Stop())))
}

func TestUndefinedKey(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("k")
	g.Expect(err).To(HaveOccurred())
	g.Expect(ms).To(BeNil())
}

func evalExpectNil(g *WithT, it *interpreter.Interpreter, input string) {
	ms, err := it.Eval(input)
	if err != nil {
		panic(err)
	}
	g.Expect(ms).To(BeNil())
}

func TestNoteAlreadyAssigned(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 60")

	_, err := it.Eval("assign c 61")
	g.Expect(err).To(HaveOccurred())
}

func TestSharpNote(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c#")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 61, 127))))
}

func TestFlatNote(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c$")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 59, 127))))
}

func TestSharpNoteRange(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 127")

	ms, err := it.Eval("c#")
	g.Expect(err).To(HaveOccurred())
	g.Expect(ms).To(BeNil())
}

func TestFlatNoteRange(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 0")

	ms, err := it.Eval("c$")
	g.Expect(err).To(HaveOccurred())
	g.Expect(ms).To(BeNil())
}

func TestAccentuatedAndGhostNote(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "velocity 50")
	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c^")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 60, 100))))

	ms, err = it.Eval("c)")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 60, 25))))
}

func TestAccentutedNoteRange(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "velocity 127")
	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c^")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 60, 127))))
}

func TestNoteLengths(t *testing.T) {
	for _, tc := range []struct {
		input    string
		duration smf.MetricTicks
	}{
		{
			input:    "k", // Quarter note.
			duration: constants.TicksPerQuarter,
		},
		{
			input:    "k.", // Dotted quarter note, x1.5.
			duration: constants.TicksPerQuarter * 3 / 2,
		},
		{
			input:    "k..", // Double dotted quarter note, x1.75.
			duration: constants.TicksPerQuarter * 7 / 4,
		},
		{
			input:    "k...", // Triplet dotted quarter note, x1.875.
			duration: constants.TicksPerQuarter * 15 / 8,
		},
		{
			input:    "k/5", // Quintuplet quarter note.
			duration: constants.TicksPerQuarter * 2 / 5,
		},
		{
			input:    "k./3", // Dotted triplet quarter note == quarter note.
			duration: constants.TicksPerQuarter,
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			evalExpectNil(g, it, "assign k 36")

			ms, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ms).To(HaveLen(1))
			g.Expect(ms[0].Pos).To(Equal(uint8(0)))
			g.Expect(ms[0].Duration).To(Equal(uint8(tc.duration.Ticks32th())))
			g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 36, 127))))
		})
	}
}

func TestCommandForbiddenInBar(t *testing.T) {
	for _, input := range []string{
		`bar "forbidden"`,
		`play "forbidden"`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewWithT(t)

			it := interpreter.New()

			evalExpectNil(g, it, `bar "bar"`)

			ms, err := it.Eval(input)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("not ended"))
			g.Expect(ms).To(BeNil())
		})
	}
}

func TestInvalidBarLength(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, `bar "bar"`)
	evalExpectNil(g, it, `timesig 3 8`)

	ms, err := it.Eval("[kk]8")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("invalid bar length"))
	g.Expect(ms).To(BeNil())
}

func TestTempoAndTimeSigInBar(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign k 36")
	evalExpectNil(g, it, "assign s 38")
	evalExpectNil(g, it, `bar "bar"`)
	evalExpectNil(g, it, `tempo 200`)
	evalExpectNil(g, it, `timesig 1 4`)
	evalExpectNil(g, it, `k`)
	evalExpectNil(g, it, `s`)
	evalExpectNil(g, it, `end`)

	ms, err := it.Eval(`play "bar"`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(4))

	// TODO: consistof in any order

	g.Expect(ms[0].Pos).To(Equal(uint8(0)))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.Message(smf.MetaTempo(200)))))

	g.Expect(ms[1].Pos).To(Equal(uint8(0)))
	g.Expect(ms[1].Message).To(Equal(smf.Message(midi.Message(smf.MetaMeter(1, 4)))))

	g.Expect(ms[2].Pos).To(Equal(uint8(0)))
	g.Expect(ms[2].Message).To(Equal(smf.Message(midi.NoteOn(0, 36, 127))))

	g.Expect(ms[3].Pos).To(Equal(uint8(0)))
	g.Expect(ms[3].Message).To(Equal(smf.Message(midi.NoteOn(0, 38, 127))))

	// g.Expect(ms[4].Pos).To(Equal(uint32(constants.TicksPerQuarter)))
	// g.Expect(ms[4].Message).To(Equal(midi.NoteOff(0, 36)))

	// g.Expect(ms[5].Pos).To(Equal(uint32(constants.TicksPerQuarter)))
	// g.Expect(ms[5].Message).To(Equal(midi.NoteOff(0, 38)))
}

func TestBar(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	for _, input := range []string{
		`velocity 127`,
		`channel 1`,
		`assign x 36`,
		`channel 2`,
		`assign x 38`,
		``,
		`bar "verse"`,
		`velocity 100`,
		`channel 1`,
		`x`,
		`channel 2`,
		`x`,
		`end`,
		// The channel is still 2 and velocity 100.
		`bar "default"`,
		`x`,
		`end`,
	} {
		evalExpectNil(g, it, input)
	}

	ms, err := it.Eval(`play "verse"`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(2))

	g.Expect(ms[0].Pos).To(Equal(uint8(0)))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(1, 36, 100))))

	g.Expect(ms[1].Pos).To(Equal(uint8(0)))
	g.Expect(ms[1].Message).To(Equal(smf.Message(midi.NoteOn(2, 38, 100))))

	// TODO
	// g.Expect(ms[2].Pos).To(Equal(uint32(constants.TicksPerQuarter)))
	// g.Expect(ms[2].Message).To(Equal(midi.NoteOff(1, 36)))

	// g.Expect(ms[3].Pos).To(Equal(uint32(constants.TicksPerQuarter)))
	// g.Expect(ms[3].Message).To(Equal(midi.NoteOff(2, 38)))

	ms, err = it.Eval(`play "default"`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))

	g.Expect(ms[0].Pos).To(Equal(uint8(0)))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(2, 38, 100))))

	// g.Expect(ms[1].Pos).To(Equal(uint32(constants.TicksPerQuarter * 2)))
	// g.Expect(ms[1].Message).To(Equal(midi.NoteOff(2, 38)))
}

func TestLetRing(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, `assign k 36`)

	ms, err := it.Eval(`k*`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))

	g.Expect(ms[0].Pos).To(Equal(uint8(0)))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 36, 127))))

	// Expect the ringing note to be turned off.
	// TODO

	ms, err = it.Eval(`k`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))

	g.Expect(ms[0].Pos).To(Equal(uint8(0)))
	g.Expect(ms[0].Message).To(Equal(smf.Message(midi.NoteOn(0, 36, 127))))

	// g.Expect(ms[2].Pos).To(Equal(uint32(constants.TicksPerQuarter * 2)))
	// g.Expect(ms[2].Message).To(Equal(smf.Message(midi.NoteOff(0, 36))))
}

func TestEvalAll(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()

	input := `
		velocity 100
		channel 1
		assign x 36
		channel 2
		assign x 38

		bar "verse"
		channel 1
		x
		channel 2
		x
		end
        play "verse"
    `
	// TODO bar not filled error

	song, err := it.EvalAll(strings.NewReader(input))
	g.Expect(err).NotTo(HaveOccurred())

	spew.Dump(song)

}

var (
	testFile  []byte
	lineCount int
)

func init() {
	b, err := ioutil.ReadFile("../../examples/bonham")
	if err != nil {
		panic(err)
	}
	testFile = b
	lineCount = bytes.Count(testFile, []byte{'\n'})
}

func BenchmarkInterpreter(b *testing.B) {
	start := time.Now()

	b.ReportAllocs()
	b.ResetTimer()

	var err error

	for i := 0; i < b.N; i++ {
		it := interpreter.New()
		_, err = it.EvalAll(bytes.NewReader(testFile))
	}

	b.StopTimer()

	if err != nil {
		panic(err)
	}

	elapsed := time.Since(start)

	linesPerNano := float64(b.N*lineCount) / float64(elapsed)

	fmt.Printf("lines per second: %f\n", linesPerNano*float64(time.Second))
}
