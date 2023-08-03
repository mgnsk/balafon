package balafon

import (
	"fmt"
	"strings"

	"github.com/mgnsk/balafon/internal/ast"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Channel is a MIDI channel.
type Channel uint8

func (c Channel) Uint8() uint8 {
	return uint8(c)
}

func (c Channel) Human() uint8 {
	return uint8(c) + 1
}

func NewChannelFromMIDI(ch uint8) Channel {
	return Channel(ch)
}

func NewChannelFromHuman(ch uint8) Channel {
	return Channel(ch - 1)
}

// Voice is a score voice.
type Voice uint8

func (v Voice) Uint8() uint8 {
	return uint8(v)
}

// Event is a balafon event.
type Event struct {
	Note     *ast.Note // only for note on messages and rests
	Message  smf.Message
	IsFlat   bool   // if the midi note was lowered due to key sig
	Pos      uint32 // in relative ticks from beginning of bar
	Duration uint32 // in ticks
	Voice    Voice
	Track    uint8 // track is the MIDI channel in human value
}

func (e *Event) String() string {
	var s strings.Builder

	if e.Track > 0 {
		s.WriteString(fmt.Sprintf("track: %d ", e.Track))
	}

	s.WriteString(fmt.Sprintf("pos: %d dur: %d", e.Pos, e.Duration))

	if e.Voice > 0 {
		s.WriteString(fmt.Sprintf(" voice: %d", e.Voice))
	}

	if e.Note != nil {
		s.WriteString(" note: ")
		e.Note.WriteTo(&s)
	}

	s.WriteString(" message: ")
	s.WriteString(e.Message.String())

	return s.String()
}
