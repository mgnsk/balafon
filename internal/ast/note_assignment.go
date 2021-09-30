package ast

import (
	"errors"
	"fmt"

	"github.com/mgnsk/gong/internal/parser/token"
)

// NoteAssignment assigns a MIDI key to a note.
type NoteAssignment struct {
	Note rune
	Key  uint8
}

// NewNoteAssignment creates a new note assignment.
func NewNoteAssignment(name, key *token.Token) (NoteAssignment, error) {
	rs := []rune(string(name.Lit))
	if len(rs) != 1 {
		return NoteAssignment{}, errors.New("note must be a single character")
	}

	v, err := key.Int32Value()
	if err != nil {
		return NoteAssignment{}, err
	}

	if v > 127 {
		return NoteAssignment{}, errors.New("note key must be in range 0-127")
	}

	return NoteAssignment{
		Note: rs[0],
		Key:  uint8(v),
	}, nil
}

func (a NoteAssignment) String() string {
	return fmt.Sprintf("%c = %d", a.Note, a.Key)
}
