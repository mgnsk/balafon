package ast

import (
	"io"
	"sort"
	"strconv"

	"github.com/mgnsk/balafon/constants"
	"github.com/mgnsk/balafon/internal/parser/token"
	"github.com/mgnsk/balafon/internal/tokentype"
)

// Note: some list operations may be implemented with side effects.

// NoteList is a list of notes.
type NoteList []*Note

func (l NoteList) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	for _, note := range l {
		n += ew.WriteFrom(note)
	}

	return int64(n), ew.Flush()
}

// NewNoteList creates a list of notes.
func NewNoteList(note Node, inner NoteList) (list NoteList) {
	switch n := note.(type) {
	case *Note:
		return append(NoteList{n}, inner...)
	case NoteList:
		// The first argument is NoteGroup.
		return append(n, inner...)
	default:
		panic("NewNoteList: invalid argument type")
	}
}

// NewNoteListFromGroup creates a note list from a group of notes with shared properties.
func NewNoteListFromGroup(notes NoteList, props PropertyList) (NoteList, error) {
	if len(props) == 0 {
		// Just a grouping of notes without properties, e.g. [cde].
		return notes, nil
	}

	// Apply the properties to all notes.
	for _, note := range notes {
		for _, p := range props {
			note.Props = append(note.Props, p)
		}
		sort.Sort(note.Props)
	}

	return notes, nil
}

// Note is a single note with sorted properties.
type Note struct {
	Props PropertyList
	Name  rune
}

// Len returns the note duration in ticks.
func (note *Note) Len() uint32 {
	length := uint32(constants.TicksPerWhole) / uint32(note.Value())
	newLength := length
	dots := note.NumDots()
	for i := 0; i < dots; i++ {
		length /= 2
		newLength += length
	}
	if division := uint32(note.Tuplet()); division > 0 {
		newLength = newLength * 2 / division
	}
	return newLength
}

// IsPause reports whether the note is a pause.
func (note *Note) IsPause() bool {
	return note.Name == '-'
}

// NumSharp returns the number of sharp signs.
func (note *Note) NumSharp() int {
	return note.countProps(tokentype.PropSharp)
}

// NumFlat reports the number of flat signs.
func (note *Note) NumFlat() int {
	return note.countProps(tokentype.PropFlat)
}

// NumStaccato reports the number of staccato properties.
func (note *Note) NumStaccato() int {
	return note.countProps(tokentype.PropStaccato)
}

// NumAccent reports the number of accent properties.
func (note *Note) NumAccent() int {
	return note.countProps(tokentype.PropAccent)
}

// NumMarcato reports the number of marcato properties.
func (note *Note) NumMarcato() int {
	return note.countProps(tokentype.PropMarcato)
}

// NumGhosts reports the number of ghost properties.
func (note *Note) NumGhosts() int {
	return note.countProps(tokentype.PropGhost)
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (note *Note) Value() uint8 {
	tok := note.Props.Find(tokentype.Uint)
	if tok == nil {
		// Implicit quarter note.
		return 4
	}
	v, err := strconv.Atoi(string(tok.Lit))
	if err != nil {
		panic(err)
	}
	return uint8(v)
}

// NumDots reports the number of dot properties.
func (note *Note) NumDots() int {
	return note.countProps(tokentype.PropDot)
}

// Tuplet returns the irregular division value if the note is a tuplet.
func (note *Note) Tuplet() int {
	tok := note.Props.Find(tokentype.PropTuplet)
	if tok == nil {
		return 0
	}
	// Trim the "/" prefix from tuplet token to get division number.
	v, err := strconv.Atoi(string(tok.Lit[1:]))
	if err != nil {
		panic(err)
	}
	return v
}

// IsLetRing reports whether the note must ring.
func (note *Note) IsLetRing() bool {
	return note.Props.Find(tokentype.PropLetRing) != nil
}

func (note *Note) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteRune(note.Name)
	n += ew.WriteFrom(note.Props)

	return int64(n), ew.Flush()
}

func (note *Note) countProps(typ token.Type) int {
	var count int
	for _, t := range note.Props {
		if t.Type == typ {
			count++
		}
	}
	return count
}

// NewNote creates a note with properties.
func NewNote(name rune, propList PropertyList) *Note {
	return &Note{
		Props: propList,
		Name:  name,
	}
}
