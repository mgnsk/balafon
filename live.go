package balafon

import (
	"fmt"
	"io"
	"unicode/utf8"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

const (
	// ETX is keycode for Ctrl-C.
	ETX = '\x03'
	// EOT is keycode for Ctrl+D.
	EOT = '\x04'
)

// LiveShell is an unbuffered live shell.
type LiveShell struct {
	r                io.Reader
	it               *Interpreter
	buf              []byte
	out              drivers.Out
	exitRequestCount int
}

// NewLiveShell creates a new live shell.
func NewLiveShell(r io.Reader, it *Interpreter, out drivers.Out) *LiveShell {
	return &LiveShell{
		r:   r,
		it:  it,
		buf: make([]byte, 1),
		out: out,
	}
}

// HandleNext handles the next character from input.
func (s *LiveShell) HandleNext() error {
	_, err := s.r.Read(s.buf)
	if err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	r, _ := utf8.DecodeRune(s.buf)
	if r == EOT || r == ETX {
		if s.exitRequestCount > 0 {
			return io.EOF
		}
		s.exitRequestCount++
		fmt.Printf("press Ctrl-D or Ctrl-C again to exit\r\n")
		return nil
	}

	s.exitRequestCount = 0

	if err := s.it.EvalString(string(r)); err != nil {
		fmt.Printf("%s\r\n", err.Error())
		return nil
	}

	for _, bar := range s.it.Flush() {
		for _, ev := range bar.Events {
			if ev.Message.Is(midi.NoteOnMsg) {
				if err := s.out.Send(ev.Message); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// Run the shell.
func (s *LiveShell) Run() error {
	for {
		if err := s.HandleNext(); err != nil {
			return err
		}
	}
}
