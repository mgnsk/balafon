package scanner_test

import (
	"strings"
	"testing"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/scanner"
	. "github.com/onsi/gomega"
)

func TestTempo(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "tempo 120"

	s := scanner.New(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	g.Expect(s.Messages()).To(ConsistOf(scanner.Message{
		Tempo: 120,
	}))
}

func TestProgramChange(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "program 0"

	s := scanner.New(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	messages := s.Messages()

	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & ProgramChangeMsg program: 0"))
}

func TestControlChange(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "control 0 1"

	s := scanner.New(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeTrue())
	g.Expect(s.Err()).NotTo(HaveOccurred())

	messages := s.Messages()

	g.Expect(messages).To(HaveLen(1))
	g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & ControlChangeMsg controller: 0 change: 1"))
}

// TODO change bnf to accept only single char instead of multiNote for assignment
func TestInvalidAssignment(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "cc = 120"

	s := scanner.New(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeFalse())
	g.Expect(s.Err()).To(HaveOccurred())
	g.Expect(s.Messages()).To(BeNil())
}

func TestUndefinedKey(t *testing.T) {
	g := NewGomegaWithT(t)

	input := "k"

	s := scanner.New(strings.NewReader(input))
	g.Expect(s.Scan()).To(BeFalse())
	g.Expect(s.Err()).To(HaveOccurred())
	g.Expect(s.Messages()).To(BeNil())
}

func TestNoteLengths(t *testing.T) {
	g := NewGomegaWithT(t)

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
			input: "k=36\nk/5", // Quintuplet quarter note.
			offAt: uint64(constants.TicksPerQuarter * 2 / 5),
		},
		{
			input: "k=36\nk./3", // Dotted triplet quarter note == quarter note.
			offAt: uint64(constants.TicksPerQuarter),
		},
	} {
		s := scanner.New(strings.NewReader(tc.input))
		g.Expect(s.Scan()).To(BeTrue())
		g.Expect(s.Err()).NotTo(HaveOccurred())

		messages := s.Messages()

		g.Expect(messages).To(HaveLen(2))

		g.Expect(messages[0].Tick).To(Equal(uint64(0)))
		g.Expect(messages[0].Msg).To(ContainSubstring("Channel0Msg & NoteOnMsg key: 36"))

		g.Expect(messages[1].Tick).To(Equal(tc.offAt))
		g.Expect(messages[1].Msg).To(ContainSubstring("Channel0Msg & NoteOffMsg key: 36"))
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

	s := scanner.New(strings.NewReader(input))
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
