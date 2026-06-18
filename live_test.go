package balafon_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/mgnsk/balafon"
	"github.com/mgnsk/balafon/internal/constants"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

type out struct {
	drivers.Port
	buf *bytes.Buffer
}

func (o *out) Send(b []byte) error {
	if o.buf != nil {
		o.buf.Write(b)
	}
	return nil
}

func TestLiveShell(t *testing.T) {
	t.Run("one byte input", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()
		g.Expect(it.EvalString(":assign a 60")).To(Succeed())

		buf := &bytes.Buffer{}

		s := balafon.NewLiveShell(strings.NewReader("a"), it, &out{buf: buf})
		g.Expect(s.HandleNext()).To(Succeed())
		g.Expect(midi.Message(buf.Bytes())).To(Equal(midi.NoteOn(0, 60, constants.DefaultVelocity)))
	})

	t.Run("more one byte input", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()
		g.Expect(it.EvalString(":assign a 60")).To(Succeed())

		buf := &bytes.Buffer{}

		s := balafon.NewLiveShell(strings.NewReader("ä"), it, &out{buf: buf})
		err := s.HandleNext()
		g.Expect(err).To(HaveOccurred())
		g.Expect(err.Error()).To(ContainSubstring(`token "ä"`))
	})
}

func TextLiveShellExit(t *testing.T) {
	t.Run("Ctrl-D or Ctrl-C twice in a row", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()

		buf := &bytes.Buffer{}

		s := balafon.NewLiveShell(bytes.NewReader([]byte{balafon.EOT, balafon.EOT}), it, &out{buf: buf})
		g.Expect(s.HandleNext()).To(Succeed())                          // Press again.
		g.Expect(s.HandleNext()).Error().To(MatchErrorStrictly(io.EOF)) // Exit.
	})

	t.Run("canceling the shutdown by pressing anything else", func(t *testing.T) {
		g := NewWithT(t)

		it := balafon.New()
		g.Expect(it.EvalString(":assign a 60")).To(Succeed())

		buf := &bytes.Buffer{}

		s := balafon.NewLiveShell(bytes.NewReader([]byte{balafon.EOT, 'a', balafon.EOT, balafon.EOT}), it, &out{buf: buf})
		g.Expect(s.HandleNext()).To(Succeed()) // Press again.
		g.Expect(s.HandleNext()).To(Succeed()) // Pressed 'a', shutdown canceled.
		g.Expect(midi.Message(buf.Bytes())).To(Equal(midi.NoteOn(0, 60, constants.DefaultVelocity)))
		g.Expect(s.HandleNext()).To(Succeed())                          // Press again.
		g.Expect(s.HandleNext()).Error().To(MatchErrorStrictly(io.EOF)) // Exit.
	})
}

func BenchmarkLiveShell(b *testing.B) {
	it := balafon.New()
	if err := it.EvalString(":assign a 60"); err != nil {
		b.Fatal(err)
	}

	reader := strings.NewReader("a")

	s := balafon.NewLiveShell(reader, it, &out{})

	b.ReportAllocs()

	for b.Loop() {
		reader.Seek(0, io.SeekStart)
		if err := s.HandleNext(); err != nil {
			b.Fatal(err)
		}
	}
}
