package shell

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/mgnsk/balafon/interpreter"
	"gitlab.com/gomidi/midi/v2"
)

const (
	// EOT is keycode for Ctrl+D.
	EOT = 4
)

// Out is the interface for an opened MIDI output port.
type Out interface {
	Send(data []byte) error
}

// LiveShell is an unbuffered live shell.
type LiveShell struct {
	r   io.Reader
	it  *interpreter.Interpreter
	buf []byte
	out Out
}

// NewLiveShell creates a new live shell.
func NewLiveShell(r io.Reader, it *interpreter.Interpreter, out Out) *LiveShell {
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
	if r == EOT {
		return io.EOF
	}

	if err := s.it.EvalString(string(r)); err != nil {
		fmt.Printf("%s\n", err.Error())
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
