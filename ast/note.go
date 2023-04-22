package ast

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/mgnsk/balafon/constants"
	"github.com/mgnsk/balafon/internal/parser/token"
)

// Note: some list operations may be implemented with side effects.

// NoteList is a list of notes.
// TODO why pointer note?
type NoteList []*Note

func (l NoteList) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	for _, note := range l {
		n += ew.WriteFrom(note)
	}

	return int64(n), ew.Flush()
}

// func (l NoteList) String() string {
// 	notes := make([]string, len(l))
// 	for i, note := range l {
// 		notes[i] = note.String()
// 	}
// 	return strings.Join(notes, " ")
// }

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

	// Apply the properties to all notes, replacing duplicate
	// property types if needed.
	for _, note := range notes {
		for _, p := range props {
			switch p.Type {
			case typeAccent, typeGhost, typeDot:
			default:
				// must be overwritten
				// error if exists??
				if _, ok := note.Props.Find(p.Type); ok {
					// TODO: similar error in property.go
					return nil, fmt.Errorf("duplicate note property '%s': '%c'", token.TokMap.Id(p.Type), p.Lit)
				}
				// if idx, ok := note.Props.Find(p.Type); ok {
				// 	note.Props[idx] = p
				// }
			}
			note.Props = append(note.Props, p)
		}
		// TODO
		sort.Sort(note.Props)
	}

	return notes, nil
}

// Note is a single note with sorted properties.
type Note struct {
	Props PropertyList
	Token *token.Token
	Name  rune
}

// Len returns the note duration in ticks.
func (note *Note) Len() uint32 {
	length := uint32(constants.TicksPerWhole) / uint32(note.Value())
	newLength := length
	dots := note.NumDots()
	for i := uint(0); i < dots; i++ {
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

// IsSharp reports whether the note is sharp.
func (note *Note) IsSharp() bool {
	_, ok := note.Props.Find(typeSharp)
	return ok
}

// IsFlat reports whether the note is flat.
func (note *Note) IsFlat() bool {
	_, ok := note.Props.Find(typeFlat)
	return ok
}

// NumAccents reports the number of accent properties in the note.
func (note *Note) NumAccents() uint {
	return note.countProps(typeAccent)
}

// NumGhosts reports the number of ghost properties in the note.
func (note *Note) NumGhosts() uint {
	return note.countProps(typeGhost)
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (note *Note) Value() uint8 {
	i, ok := note.Props.Find(typeUint)
	if !ok {
		// Implicit quarter note.
		return 4
	}
	v, err := strconv.Atoi(string(note.Props[i].Lit))
	if err != nil {
		panic(err)
	}
	// TODO range validation.
	return uint8(v)
}

// NumDots reports the number of dot properties in the note.
func (note *Note) NumDots() uint {
	return note.countProps(typeDot)
}

// Tuplet returns the irregular division value if the note is a tuplet.
func (note *Note) Tuplet() uint8 {
	if i, ok := note.Props.Find(typeTuplet); ok {
		// Extract the division number.
		// For example "3" for a triplet denoted by "/3".
		v, err := strconv.Atoi(string(note.Props[i].Lit[1:]))
		if err != nil {
			panic(err)
		}
		return uint8(v)
	}
	return 0
}

// IsLetRing reports whether the note must ring.
func (note *Note) IsLetRing() bool {
	_, ok := note.Props.Find(typeLetRing)
	return ok
}

func (note *Note) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteRune(note.Name)
	n += ew.WriteFrom(note.Props)

	return int64(n), ew.Flush()
}

// func (note *Note) String() string {
// 	return fmt.Sprintf("%c%s", note.Name, note.Props)
// }

func (note *Note) countProps(targetType token.Type) uint {
	num := uint(0)
	for _, t := range note.Props {
		if t.Type == targetType {
			num++
		}
	}
	return num
}

// NewNote creates a note with properties.
func NewNote(symbol *token.Token, propList PropertyList) *Note {
	return &Note{
		Props: propList,
		Token: symbol,
		Name:  []rune(symbol.IDValue())[0],
	}
}
