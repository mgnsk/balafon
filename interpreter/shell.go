package interpreter

import (
	"bytes"

	"github.com/c-bata/go-prompt"
	"github.com/davecgh/go-spew/spew"
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
func (s *Shell) Execute(bars ...*Bar) error {
	// isPlayable := false
	// for _, bar := range bars {
	// 	for _, ev := range bar.Events {
	// 		if ev.Message.IsPlayable() {
	// 			isPlayable = true
	// 			break
	// 		}
	// 	}
	// }
	// if !isPlayable {
	// 	return nil
	// }

	song := sequencer.New()
	for _, bar := range bars {
		song.AddBar(bar.Export())
	}

	sm := song.ToSMF1()
	spew.Dump(sm.TempoChanges())

	s.buf.Reset()

	if _, err := sm.WriteTo(&s.buf); err != nil {
		return err
	}

	rd := smf.ReadTracksFrom(&s.buf)
	if err := rd.Error(); err != nil {
		return err
	}

	rd.Do(func(ev smf.TrackEvent) {
		// fmt.Println(ev)

	})

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
