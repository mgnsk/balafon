package ast

import (
	"errors"
	"fmt"
	"strconv"
)

// NoteAssignment assigns a MIDI key to a note.
type NoteAssignment struct {
	Note string
	Key  uint8
}

// NewNoteAssignment creates a new note assignment.
// TODO single char note
func NewNoteAssignment(note, key string) (NoteAssignment, error) {
	v, err := strconv.Atoi(key)
	if err != nil {
		return NoteAssignment{}, err
	}
	if v > 127 {
		return NoteAssignment{}, errors.New("note key must be in range 0-127")
	}
	return NoteAssignment{
		Note: note,
		Key:  uint8(v),
	}, nil
}

func (a NoteAssignment) String() string {
	return fmt.Sprintf("%s = %d", a.Note, a.Key)
}
