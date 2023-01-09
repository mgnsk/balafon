package interpreter

import (
	"math"

	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Event is a MIDI event.
type Event struct {
	Message  smf.Message
	Pos      uint32 // in ticks
	Duration uint32 // in ticks
	Channel  uint8
}

// IsLetRing returns whether the Event is a note that is let ring.
func (e *Event) IsLetRing() bool {
	return e.Message.Is(midi.NoteOnMsg) && e.Duration == math.MaxUint16
}
