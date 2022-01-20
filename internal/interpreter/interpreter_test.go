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
)

func TestTempoCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("tempo 120")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(ConsistOf(interpreter.Message{
		Tempo: 120,
	}))
}

func TestProgramChangeCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("program 0")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & ProgramChangeMsg program: 0"))
}

func TestControlChangeCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("control 0 1")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & ControlChangeMsg controller: 0 change: 1"))
}

func TestStartCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("start")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("StartMsg"))
}

func TestStopCommand(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("stop")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("StopMsg"))
}

func TestUndefinedKey(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("k")
	g.Expect(err).To(HaveOccurred())
	g.Expect(messages).To(BeNil())
}

func evalExpectNil(g *WithT, it *interpreter.Interpreter, input string) {
	messages, err := it.Eval(input)
	if err != nil {
		panic(err)
	}
	g.Expect(messages).To(BeNil())
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

	messages, err := it.Eval("c#")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 61"))
}

func TestFlatNote(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 60")

	messages, err := it.Eval("c$")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 59"))
}

func TestSharpNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 127")

	messages, err := it.Eval("c#")
	g.Expect(err).To(HaveOccurred())
	g.Expect(messages).To(BeNil())
}

func TestFlatNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "assign c 0")

	messages, err := it.Eval("c$")
	g.Expect(err).To(HaveOccurred())
	g.Expect(messages).To(BeNil())
}

func TestAccentuatedAndGhostNote(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "velocity 50")
	evalExpectNil(g, it, "assign c 60")

	messages, err := it.Eval("c^")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("velocity: 100"))

	messages, err = it.Eval("c)")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("velocity: 25"))
}

func TestAccentutedNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, "velocity 127")
	evalExpectNil(g, it, "assign c 60")

	messages, err := it.Eval("c^")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("velocity: 127"))
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

			messages, err := it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(messages).To(HaveLen(2))
			g.Expect(messages[0].Tick).To(Equal(uint32(0)))
			g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 36"))
			g.Expect(messages[1].Tick).To(Equal(tc.offAt))
			g.Expect(messages[1].Msg).To(ContainSubstring("Channel0Msg & NoteOffMsg key: 36"))
		})
	}
}

func TestCommandForbiddenInBar(t *testing.T) {
	for _, input := range []string{
		`assign c 10`,
		`bar "forbidden"`,
		`play "forbidden"`,
		`tempo 120`,
		`start`,
		`stop`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			it := interpreter.New()

			evalExpectNil(g, it, `bar "bar"`)

			messages, err := it.Eval(input)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("not ended"))
			g.Expect(messages).To(BeNil())
		})
	}
}

func TestTimeSigForbiddenOutsideBar(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("timesig 4 4")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("timesig"))
	g.Expect(messages).To(BeNil())
}

func TestInvalidBarLength(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, `bar "bar"`)
	evalExpectNil(g, it, `timesig 3 8`)

	messages, err := it.Eval("[kk]8")
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("invalid bar length"))
	g.Expect(messages).To(BeNil())
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

	messages, err := it.Eval(`play "verse"`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(4))

	g.Expect(messages[0].Tick).To(Equal(uint32(0)))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel1Msg & NoteOnMsg key: 36 velocity: 100"))

	g.Expect(messages[1].Tick).To(Equal(uint32(0)))
	g.Expect(messages[1].Msg).To(ContainSubstring("Channel2Msg & NoteOnMsg key: 38 velocity: 100"))

	g.Expect(messages[2].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(messages[2].Msg).To(ContainSubstring("Channel1Msg & NoteOffMsg key: 36"))

	g.Expect(messages[3].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(messages[3].Msg).To(ContainSubstring("Channel2Msg & NoteOffMsg key: 38"))

	messages, err = it.Eval(`play "default"`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))

	g.Expect(messages[0].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel2Msg & NoteOnMsg key: 38 velocity: 100"))

	g.Expect(messages[1].Tick).To(Equal(uint32(constants.TicksPerQuarter * 2)))
	g.Expect(messages[1].Msg).To(ContainSubstring("Channel2Msg & NoteOffMsg key: 38"))
}

func TestLetRing(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	evalExpectNil(g, it, `assign k 36`)

	messages, err := it.Eval(`k*`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(1))

	g.Expect(messages[0].Tick).To(Equal(uint32(0)))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 36"))

	// Expect the ringing note to be turned off.

	messages, err = it.Eval(`k`)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(3))

	g.Expect(messages[0].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOffMsg key: 36"))

	g.Expect(messages[1].Tick).To(Equal(uint32(constants.TicksPerQuarter)))
	g.Expect(messages[1].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 36"))

	g.Expect(messages[2].Tick).To(Equal(uint32(constants.TicksPerQuarter * 2)))
	g.Expect(messages[2].Msg).To(ContainSubstring("Channel0Msg & NoteOffMsg key: 36"))
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
