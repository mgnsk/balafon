package balafon_test

import (
	"bytes"
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

	msg := midi.Message(buf.Bytes())
	g.Expect(msg).To(Equal(midi.NoteOn(0, 60, constants.DefaultVelocity)))
}

func BenchmarkLiveShell(b *testing.B) {
	it := balafon.New()
	if err := it.EvalString(":assign a 60"); err != nil {
		b.Fatal(err)
	}

	s := balafon.NewLiveShell(&reader{}, it, &out{})

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if err := s.HandleNext(); err != nil {
			b.Fatal(err)
		}
	}
}
