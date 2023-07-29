package ast

import (
	"io"

	"github.com/mgnsk/balafon/internal/parser/token"
)

// NoteGroup is a group of notes with shared properties.
type NoteGroup struct {
	Nodes NodeList
	Props PropertyList
}

// NewNoteGroup creates a note group from a list of notes with shared properties.
func NewNoteGroup(notes NodeList, props PropertyList) (NoteGroup, error) {
	return NoteGroup{
		Nodes: notes,
		Props: props,
	}, nil
}

func (g NoteGroup) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString("[")
	n += ew.WriteFrom(g.Nodes)
	n += ew.WriteString("]")
	n += ew.WriteFrom(g.Props)

	return int64(n), ew.Flush()
}

// Note is a single note.
type Note struct {
	Pos   token.Pos
	Props PropertyList
	Name  rune
}

func (note *Note) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteRune(note.Name)
	n += ew.WriteFrom(note.Props)

	return int64(n), ew.Flush()
}

// IsPause reports whether the note is a pause.
func (note *Note) IsPause() bool {
	return note.Name == '-'
}

// NewNote creates a note with properties.
func NewNote(pos token.Pos, name rune, propList PropertyList) *Note {
	return &Note{
		Pos:   pos,
		Props: propList,
		Name:  name,
	}
}
