package interpreter_test

import (
	"strings"
	"testing"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
)

func TestTempo(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "tempo 120"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	g.Expect(s.Messages()).To(ConsistOf(interpreter.Message{
		Tempo: 120,
	}))
}

func TestProgramChange(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "program 0"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	messages := s.Messages()

	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & ProgramChangeMsg program: 0"))
}

func TestControlChange(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "control 0 1"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	messages := s.Messages()

	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & ControlChangeMsg controller: 0 change: 1"))
}

func TestUndefinedKey(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "k"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeFalse())
	g.Expect(s.Err()).To(HaveOccurred())
	g.Expect(s.Messages()).To(BeNil())
}

func TestSharpNote(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "c=60\nc#"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	messages := s.Messages()

	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 61"))
}

func TestFlatNote(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "c=60\nc$"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	messages := s.Messages()

	g.Expect(messages).To(HaveLen(2))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 59"))
}

func TestSharpNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "c=127\nc#"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeFalse())
	g.Expect(s.Err()).To(HaveOccurred())
}

func TestFlatNoteRange(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "c=0\nc$"

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeFalse())
	g.Expect(s.Err()).To(HaveOccurred())
}

func TestNoteLengths(t *testing.T) {
	for _, tc := range []struct {
		input string
		offAt uint64
	}{
		{
			input: "k=36\nk", // Quarter note.
			offAt: uint64(constants.TicksPerQuarter),
		},
		{
			input: "k=36\nk.", // Dotted quarter note.
			offAt: uint64(constants.TicksPerQuarter * 3 / 2),
		},
		{
			input: "k=36\nk..", // Double dotted quarter note.
			offAt: uint64(constants.TicksPerQuarter * 9 / 4),
		},
		{
			input: "k=36\nk...", // Triplet dotted quarter note.
			offAt: uint64(constants.TicksPerQuarter * 27 / 8),
		},
		{
			input: "k=36\nk/5", // Quintuplet quarter note.
			offAt: uint64(constants.TicksPerQuarter * 2 / 5),
		},
		{
			input: "k=36\nk./3", // Dotted triplet quarter note == quarter note.
			offAt: uint64(constants.TicksPerQuarter),
		},
	} {
		t.Run(tc.input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			s := interpreter.NewScanner(strings.NewReader(tc.input))
			g.Expect(s.Scan()).To(BeTrue())
			g.Expect(s.Err()).NotTo(HaveOccurred())

			messages := s.Messages()

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
		"bar \"bar\"\nc=10",
		"bar \"bar\"\nbar \"forbidden\"",
		"bar \"bar\"\nplay \"forbidden\"",
		"bar \"bar\"\ntempo 120",
		"bar \"bar\"\nchannel 0",
		"bar \"bar\"\nvelocity 0",
		"bar \"bar\"\nprogram 0",
		"bar \"bar\"\ncontrol 0 0",
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			s := interpreter.NewScanner(strings.NewReader(input))
			g.Expect(s.Scan()).To(BeFalse())
			g.Expect(s.Err()).To(HaveOccurred())
			g.Expect(s.Err().Error()).To(ContainSubstring("not ended"))
		})
	}
}

func TestBar(t *testing.T) {
	g := NewGomegaWithT(t)

	input := `k=36
s=38
velocity 100
channel 10
bar "verse1"
kk8
ss8
end
play "verse1"
`

	s := interpreter.NewScanner(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	messages := s.Messages()

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
