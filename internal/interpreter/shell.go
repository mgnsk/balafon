package interpreter

import (
	"bytes"
	"fmt"

	"github.com/c-bata/go-prompt"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Shell is a gong shell.
type Shell struct {
	it *Interpreter
}

// Execute the input.
func (s *Shell) Execute(in string) {
	song, err := s.it.Eval(in)
	if err != nil {
		fmt.Println(err)
		return
	}

	if song.Bars().Len() == 0 {
		return
	}

	sm := song.ToSMF1()

	// sm.Logger = log.Default()

	fmt.Println(sm.String())

	var buf bytes.Buffer

	sm.WriteFile("test.mid")
	if _, err := sm.WriteTo(&buf); err != nil {
		panic(err)
	}

	fmt.Println(buf.String())

	// tr := smf.ReadTracksFrom(&buf)

	// tr.Do(func(ev smf.TrackEvent) {
	// 	spew.Dump(ev)
	// })

	s2, err := smf.ReadFrom(&buf)
	if err != nil {
		panic(err)
	}

	fmt.Println(s2.String())
	_ = s2

	// fmt.Println(s2.String())
	// // rd := smf.ReadTracksFrom(&buf)
	// // if err := rd.Error(); err != nil {
	// // 	panic(err)
	// // }

	// // rd.Do(func(ev smf.TrackEvent) {
	// // 	panic("wat")
	// // 	// fmt.Println(ev.Message.String())
	// // })
}

// Complete the input.
func (s *Shell) Complete(in prompt.Document) []prompt.Suggest {
	return nil
}

// NewShell creates a gong shell.
func NewShell() *Shell {
	return &Shell{
		it: New(),
	}
}
