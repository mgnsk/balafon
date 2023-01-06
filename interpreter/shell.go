package interpreter

import (
	"bytes"

	"github.com/c-bata/go-prompt"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/sequencer"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Shell is a gong shell.
type Shell struct {
	out drivers.Out
	buf bytes.Buffer
}

// Execute the bars.
func (s *Shell) Execute(bars ...sequencer.Bar) error {
	if len(bars) == 0 {
		return nil
	}

	song := sequencer.New()
	for _, bar := range bars {
		song.AddBar(bar)
	}

	sm := song.ToSMF1()

	s.buf.Reset()

	if _, err := sm.WriteTo(&s.buf); err != nil {
		return err
	}

	rd := smf.ReadTracksFrom(&s.buf)
	if err := rd.Error(); err != nil {
		return err
	}

	return rd.Play(s.out)
}

// Complete the input.
func (s *Shell) Complete(in prompt.Document) []prompt.Suggest {
	return nil
}

// NewShell creates a gong shell.
func NewShell(out drivers.Out) *Shell {
	return &Shell{
		out: out,
	}
}
