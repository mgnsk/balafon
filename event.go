package balafon

import (
	"gitlab.com/gomidi/midi/v2/smf"
)

// Event is a MIDI event.
type Event struct {
	Message  smf.Message
	Pos      uint32 // in relative ticks from beginning of bar
	Duration uint32 // in ticks
}
