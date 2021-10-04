package interpreter_test

import (
	"testing"

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

func TestSharpNote(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("assign c 60")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(BeNil())

	messages, err = it.Eval("c#")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 61"))
}

func TestFlatNote(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("assign c 60")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(BeNil())

	messages, err = it.Eval("c$")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 59"))
}

func TestSharpNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("assign c 127")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(BeNil())

	messages, err = it.Eval("c#")
	g.Expect(err).To(HaveOccurred())
	g.Expect(messages).To(BeNil())
}

func TestFlatNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	messages, err := it.Eval("assign c 0")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(messages).To(BeNil())

	messages, err = it.Eval("c$")
	g.Expect(err).To(HaveOccurred())
	g.Expect(messages).To(BeNil())
}

func TestNoteLengths(t *testing.T) {
	for _, tc := range []struct {
		input string
		offAt uint64
	}{
		{
			input: "k", // Quarter note.
			offAt: uint64(constants.TicksPerQuarter),
		},
		{
			input: "k.", // Dotted quarter note, x1.5.
			offAt: uint64(constants.TicksPerQuarter * 3 / 2),
		},
		{
			input: "k..", // Double dotted quarter note, x1.75.
			offAt: uint64(constants.TicksPerQuarter * 7 / 4),
		},
		{
			input: "k...", // Triplet dotted quarter note, x1.875.
			offAt: uint64(constants.TicksPerQuarter * 15 / 8),
		},
		{
			input: "k/5", // Quintuplet quarter note.
			offAt: uint64(constants.TicksPerQuarter * 2 / 5),
		},
		{
			input: "k./3", // Dotted triplet quarter note == quarter note.
			offAt: uint64(constants.TicksPerQuarter),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			it := interpreter.New()

			messages, err := it.Eval("assign k 36")
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(messages).To(BeNil())

			messages, err = it.Eval(tc.input)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(messages).To(HaveLen(2))
			g.Expect(messages[0].Tick).To(Equal(uint64(0)))
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
		`channel 0`,
		`velocity 0`,
		`program 0`,
		`control 0 0`,
		`start`,
		`stop`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			it := interpreter.New()

			messages, err := it.Eval(`bar "bar"`)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(messages).To(BeNil())

			messages, err = it.Eval(input)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring("not ended"))
			g.Expect(messages).To(BeNil())
		})
	}
}

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

	it := interpreter.New()

	for _, input := range []string{
		`assign k 36`,
		`assign s 38`,
		`velocity 100`,
		`channel 10`,
		`bar "verse"`,
		`[kk]8`,
		`[ss]8`,
		`end`,
	} {
		messages, err := it.Eval(input)
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(messages).To(BeNil())
	}

	messages, err := it.Eval(`play "verse"`)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(messages).To(HaveLen(8))

	g.Expect(messages[0].Tick).To(Equal(uint64(0)))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel10Msg & NoteOnMsg key: 36 velocity: 100"))

	g.Expect(messages[1].Tick).To(Equal(uint64(0)))
	g.Expect(messages[1].Msg).To(ContainSubstring("Channel10Msg & NoteOnMsg key: 38 velocity: 100"))

	g.Expect(messages[2].Tick).To(Equal(uint64(constants.TicksPerQuarter / 2)))
	g.Expect(messages[2].Msg).To(ContainSubstring("Channel10Msg & NoteOffMsg key: 36"))

	g.Expect(messages[3].Tick).To(Equal(uint64(constants.TicksPerQuarter / 2)))
	g.Expect(messages[3].Msg).To(ContainSubstring("Channel10Msg & NoteOffMsg key: 38"))

	g.Expect(messages[4].Tick).To(Equal(uint64(constants.TicksPerQuarter / 2)))
	g.Expect(messages[4].Msg).To(ContainSubstring("Channel10Msg & NoteOnMsg key: 36 velocity: 100"))

	g.Expect(messages[5].Tick).To(Equal(uint64(constants.TicksPerQuarter / 2)))
	g.Expect(messages[5].Msg).To(ContainSubstring("Channel10Msg & NoteOnMsg key: 38 velocity: 100"))

	g.Expect(messages[6].Tick).To(Equal(uint64(constants.TicksPerQuarter)))
	g.Expect(messages[6].Msg).To(ContainSubstring("Channel10Msg & NoteOffMsg key: 36"))

	g.Expect(messages[7].Tick).To(Equal(uint64(constants.TicksPerQuarter)))
	g.Expect(messages[7].Msg).To(ContainSubstring("Channel10Msg & NoteOffMsg key: 38"))
}
