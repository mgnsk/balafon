package balafon

import (
	"fmt"
	"strings"

	"github.com/mgnsk/balafon/internal/ast"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Event is a MIDI event.
type Event struct {
	Note     *ast.Note // only for note on messages and rests
	Message  smf.Message
	Pos      uint32 // in relative ticks from beginning of bar
	Duration uint32 // in ticks
}

func (e *Event) String() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("pos: %d dur: %d", e.Pos, e.Duration))

	if e.Note != nil {
		s.WriteString(" note: ")
		e.Note.WriteTo(&s)
	}

	s.WriteString(" message: ")
	s.WriteString(e.Message.String())

	return s.String()
}
