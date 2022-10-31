package ast

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/parser/token"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Note: some list operations may be implemented with side effects.

// NoteList is a list of notes.
type NoteList []*Note

func (l NoteList) String() string {
	notes := make([]string, len(l))
	for i, note := range l {
		notes[i] = note.String()
	}
	return strings.Join(notes, " ")
}

// NewNoteList creates a list of notes.
func NewNoteList(note fmt.Stringer, inner NoteList) (list NoteList) {
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
	Name  rune
}

// Ticks returns the note duration in ticks.
func (n *Note) Ticks() smf.MetricTicks {
	value := n.Value()
	length := constants.TicksPerWhole / smf.MetricTicks(value)
	newLength := length
	dots := n.NumDots()
	for i := uint(0); i < dots; i++ {
		length /= 2
		newLength += length
	}
	if division := n.Tuplet(); division > 0 {
		newLength = newLength * 2 / smf.MetricTicks(division)
	}
	return newLength
}

// IsPause reports whether the note is a pause.
func (n *Note) IsPause() bool {
	return n.Name == '-'
}

// IsSharp reports whether the note is sharp.
func (n *Note) IsSharp() bool {
	_, ok := n.Props.Find(typeSharp)
	return ok
}

// IsFlat reports whether the note is flat.
func (n *Note) IsFlat() bool {
	_, ok := n.Props.Find(typeFlat)
	return ok
}

// NumAccents reports the number of accent properties in the note.
func (n *Note) NumAccents() uint {
	return n.countProps(typeAccent)
}

// NumGhosts reports the number of ghost properties in the note.
func (n *Note) NumGhosts() uint {
	return n.countProps(typeGhost)
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (n *Note) Value() uint8 {
	i, ok := n.Props.Find(typeUint)
	if !ok {
		// Implicit quarter note.
		return 4
	}
	v, err := strconv.Atoi(string(n.Props[i].Lit))
	if err != nil {
		panic(err)
	}
	// TODO range validation.
	return uint8(v)
}

// NumDots reports the number of dot properties in the note.
func (n *Note) NumDots() uint {
	return n.countProps(typeDot)
}

// Tuplet returns the irregular division value if the note is a tuplet.
func (n *Note) Tuplet() uint8 {
	if i, ok := n.Props.Find(typeTuplet); ok {
		// Extract the division number.
		// For example "3" for a triplet denoted by "/3".
		v, err := strconv.Atoi(string(n.Props[i].Lit[1:]))
		if err != nil {
			panic(err)
		}
		return uint8(v)
	}
	return 0
}

// IsLetRing reports whether the note must ring.
func (n *Note) IsLetRing() bool {
	_, ok := n.Props.Find(typeLetRing)
	return ok
}

func (n *Note) String() string {
	return fmt.Sprintf("%c%s", n.Name, n.Props)
}

func (n *Note) countProps(targetType token.Type) uint {
	num := uint(0)
	for _, t := range n.Props {
		if t.Type == targetType {
			num++
		}
	}
	return num
}

// NewNote creates a note with properties.
func NewNote(symbol string, propList PropertyList) *Note {
	return &Note{
		Name:  []rune(symbol)[0],
		Props: propList,
	}
}
