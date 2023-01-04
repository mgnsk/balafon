package interpreter

import (
	"bytes"
	"fmt"

	"github.com/c-bata/go-prompt"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/sequencer"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Shell is a gong shell.
type Shell struct {
	out drivers.Out
	it  *Interpreter
	buf bytes.Buffer
}

// Execute the input.
func (s *Shell) Execute(in string) {
	if err := s.it.Eval(in); err != nil {
		fmt.Println(err)
		return
	}

	bars := s.it.Flush()

	if len(bars) == 0 {
		return
	}

	song := sequencer.New()
	for _, bar := range bars {
		song.AddBar(bar)
	}

	sm := song.ToSMF1()

	s.buf.Reset()

	if _, err := sm.WriteTo(&s.buf); err != nil {
		panic(err)
	}

	rd := smf.ReadTracksFrom(&s.buf)
	if err := rd.Error(); err != nil {
		panic(err)
	}

	if err := rd.Play(s.out); err != nil {
		fmt.Println(err)
	}
}

// Complete the input.
func (s *Shell) Complete(in prompt.Document) []prompt.Suggest {
	return nil
}

// NewShell creates a gong shell.
func NewShell(out drivers.Out) *Shell {
	return &Shell{
		out: out,
		it:  New(),
	}
}
