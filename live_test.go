package balafon_test

import (
	"testing"

	"github.com/mgnsk/balafon"
	"gitlab.com/gomidi/midi/v2/drivers"
)

type out struct {
	drivers.Port
}

func (*out) Send([]byte) error {
	return nil
}

type reader struct{}

func (*reader) Read(p []byte) (int, error) {
	p[0] = 'a'
	return 1, nil
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
