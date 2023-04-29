package shell

import (
	"fmt"
	"io"
	"unicode/utf8"

	"github.com/mgnsk/balafon/interpreter"
	"gitlab.com/gomidi/midi/v2/smf"
)

const (
	// Keycode for Ctrl+D.
	EOT = 4

	defaultReso = 16

	gridBG    = "🟦"
	beatBG    = "⭕"
	currentBG = "🔴"
)

// EventHandler handles a live MIDI message.
type EventHandler func(smf.Message) error

// LiveShell is an unbuffered live shell.
type LiveShell struct {
	r       io.Reader
	it      *interpreter.Interpreter
	handler EventHandler
}

// NewLiveShell creates a new live shell.
func NewLiveShell(r io.Reader, it *interpreter.Interpreter, handler EventHandler) *LiveShell {
	return &LiveShell{
		r:       r,
		it:      it,
		handler: handler,
	}
}

// Run the shell.
func (s *LiveShell) Run() error {
	input := make([]byte, 1)

	for {
		_, err := s.r.Read(input)
		if err != nil {
			return fmt.Errorf("error reading from stdin: %w", err)
		}

		r, _ := utf8.DecodeRune(input)
		if r == EOT {
			return nil
		}

		if err := s.it.EvalString(string(r)); err != nil {
			fmt.Println(err)
			continue
		}

		for _, bar := range s.it.Flush() {
			for _, ev := range bar.Events {
				if err := s.handler(ev.Message); err != nil {
					return fmt.Errorf("error handling MIDI message: %w", err)
				}
			}
		}
	}
}
