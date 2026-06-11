package balafon_test

import (
	"bytes"
	"io"
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

type reader struct{}

func (*reader) Read(p []byte) (int, error) {
	p[0] = 'a'
	return 1, nil
}

func TestLiveShell(t *testing.T) {
	g := NewWithT(t)

	it := balafon.New()
	g.Expect(it.EvalString(":assign a 60")).To(Succeed())

	buf := &bytes.Buffer{}

	s := balafon.NewLiveShell(&reader{}, it, &out{buf: buf})
	g.Expect(s.HandleNext()).To(Succeed())
	g.Expect(midi.Message(buf.Bytes())).To(Equal(midi.NoteOn(0, 60, constants.DefaultVelocity)))
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

	s := balafon.NewLiveShell(&reader{}, it, &out{})

	b.ReportAllocs()

	for b.Loop() {
		if err := s.HandleNext(); err != nil {
			b.Fatal(err)
		}
	}
}
