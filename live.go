package balafon

import (
	"fmt"
	"io"
	"unicode/utf8"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
)

const (
	// EOT is keycode for Ctrl+D.
	EOT = 4
)

// LiveShell is an unbuffered live shell.
type LiveShell struct {
	r   io.Reader
	p   *Parser
	buf []byte
	out drivers.Out
}

// NewLiveShell creates a new live shell.
func NewLiveShell(r io.Reader, p *Parser, out drivers.Out) *LiveShell {
	return &LiveShell{
		r:   r,
		p:   p,
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

	if err := s.p.EvalString(string(r)); err != nil {
		fmt.Printf("%s\n", err.Error())
		return nil
	}

	for _, bar := range s.p.Flush() {
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
