package interpreter_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func TestProgramChangeCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("program 0")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(midi.ProgramChange(0, 0)))
}

func TestControlChangeCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("control 0 1")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(midi.ControlChange(0, 0, 1)))
}

func TestStartCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("start")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(midi.Start()))
}

func TestStopCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("stop")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))
	g.Expect(ms[0].Message).To(Equal(midi.Stop()))
}

func TestUndefinedKey(t *testing.T) {
	g := NewGomegaWithT(t)

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
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 60")

	_, err := it.Eval("assign c 61")
	g.Expect(err).To(HaveOccurred())
}

func TestSharpNote(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c#")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(2))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(0, 61, 127)))
}

func TestFlatNote(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c$")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(2))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(0, 59, 127)))
}

func TestSharpNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 127")

	ms, err := it.Eval("c#")
	g.Expect(err).To(HaveOccurred())
	g.Expect(ms).To(BeNil())
}

func TestFlatNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 0")

	ms, err := it.Eval("c$")
	g.Expect(err).To(HaveOccurred())
	g.Expect(ms).To(BeNil())
}

func TestAccentuatedAndGhostNote(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "velocity 50")
	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c^")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(2))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(0, 60, 100)))

	ms, err = it.Eval("c)")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(2))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(0, 60, 25)))
}

func TestAccentutedNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "velocity 127")
	evalExpectNil(g, it, "assign c 60")

	ms, err := it.Eval("c^")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(2))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(0, 60, 127)))
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
			g := NewGomegaWithT(t)

			it := interpreter.New()

			evalExpectNil(g, it, "assign k 36")

			ms, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(ms).To(HaveLen(2))
			g.Expect(ms[0].Tick).To(Equal(uint32(0)))
			g.Expect(ms[0].Message).To(Equal(midi.NoteOn(0, 36, 127)))
			g.Expect(ms[1].Tick).To(Equal(tc.offAt))
			g.Expect(ms[1].Message).To(Equal(midi.NoteOff(0, 36)))
		})
	}
}

func TestCommandForbiddenInBar(t *testing.T) {
	for _, input := range []string{
		`assign c 10`,
		`bar "forbidden"`,
		`play "forbidden"`,
		`start`,
		`stop`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			it := interpreter.New()

			evalExpectNil(g, it, `bar "bar"`)

			ms, err := it.Eval(input)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("not ended"))
			g.Expect(ms).To(BeNil())
		})
	}
}

func TestTimeSigForbiddenOutsideBar(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	ms, err := it.Eval("timesig 4 4")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("timesig"))
	g.Expect(ms).To(BeNil())
}

func TestInvalidBarLength(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, `bar "bar"`)
	evalExpectNil(g, it, `timesig 3 8`)

	ms, err := it.Eval("[kk]8")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("invalid bar length"))
	g.Expect(ms).To(BeNil())
}

func TestTempoAndTimeSigInBar(t *testing.T) {
	g := NewGomegaWithT(t)

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
	g.Expect(ms).To(HaveLen(6))

	// TODO: consistof in any order

	g.Expect(ms[0].Tick).To(Equal(uint32(0)))
	g.Expect(ms[0].Message).To(Equal(midi.Message(smf.MetaTempo(200))))

	g.Expect(ms[1].Tick).To(Equal(uint32(0)))
	g.Expect(ms[1].Message).To(Equal(midi.Message(smf.MetaMeter(1, 4))))

	g.Expect(ms[2].Tick).To(Equal(uint32(0)))
	g.Expect(ms[2].Message).To(Equal(midi.NoteOn(0, 36, 127)))

	g.Expect(ms[3].Tick).To(Equal(uint32(0)))
	g.Expect(ms[3].Message).To(Equal(midi.NoteOn(0, 38, 127)))

	g.Expect(ms[4].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(ms[4].Message).To(Equal(midi.NoteOff(0, 36)))

	g.Expect(ms[5].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(ms[5].Message).To(Equal(midi.NoteOff(0, 38)))
}

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

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
	g.Expect(ms).To(HaveLen(4))

	g.Expect(ms[0].Tick).To(Equal(uint32(0)))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(1, 36, 100)))

	g.Expect(ms[1].Tick).To(Equal(uint32(0)))
	g.Expect(ms[1].Message).To(Equal(midi.NoteOn(2, 38, 100)))

	g.Expect(ms[2].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(ms[2].Message).To(Equal(midi.NoteOff(1, 36)))

	g.Expect(ms[3].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(ms[3].Message).To(Equal(midi.NoteOff(2, 38)))

	ms, err = it.Eval(`play "default"`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(2))

	g.Expect(ms[0].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(2, 38, 100)))

	g.Expect(ms[1].Tick).To(Equal(uint32(constants.TicksPerQuarter * 2)))
	g.Expect(ms[1].Message).To(Equal(midi.NoteOff(2, 38)))
}

func TestLetRing(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, `assign k 36`)

	ms, err := it.Eval(`k*`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(1))

	g.Expect(ms[0].Tick).To(Equal(uint32(0)))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOn(0, 36, 127)))

	// Expect the ringing note to be turned off.

	ms, err = it.Eval(`k`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ms).To(HaveLen(3))

	g.Expect(ms[0].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(ms[0].Message).To(Equal(midi.NoteOff(0, 36)))

	g.Expect(ms[1].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(ms[1].Message).To(Equal(midi.NoteOn(0, 36, 127)))

	g.Expect(ms[2].Tick).To(Equal(uint32(constants.TicksPerQuarter * 2)))
	g.Expect(ms[2].Message).To(Equal(midi.NoteOff(0, 36)))
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
