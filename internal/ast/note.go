package ast

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/internal/constants"
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
func NewNoteListFromGroup(notes NoteList, props PropertyList) NoteList {
	if len(props) == 0 {
		// Just a grouping of notes without properties, e.g. [cde].
		return notes
	}

	// Apply the properties to all notes, replacing duplicate
	// property types if needed.
	for _, note := range notes {
		needSort := false
		for _, p := range props {
			if idx, ok := note.Props.Find(p.Type); ok {
				note.Props[idx] = p
			} else {
				note.Props = append(note.Props, p)
				needSort = true
			}
		}
		if needSort {
			sort.Sort(note.Props)
		}
	}

	return notes
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
	dots := n.Dots()
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
	_, ok := n.Props.Find(sharpType)
	return ok
}

// IsFlat reports whether the note is flat.
func (n *Note) IsFlat() bool {
	_, ok := n.Props.Find(flatType)
	return ok
}

// IsAccent reports whether the note is accentuated.
func (n *Note) IsAccent() bool {
	_, ok := n.Props.Find(accentType)
	return ok
}

// IsGhost reports whether the note is a ghost note.
func (n *Note) IsGhost() bool {
	_, ok := n.Props.Find(ghostType)
	return ok
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (n *Note) Value() uint8 {
	i, ok := n.Props.Find(uintType)
	if !ok {
		panic("ast.Note: missing note value")
	}
	v, err := strconv.Atoi(string(n.Props[i].Lit))
	if err != nil {
		panic(err)
	}
	// TODO range validation.
	return uint8(v)
}

// Dots reports the number of dot properties in the note.
func (n *Note) Dots() uint {
	dots := uint(0)
	for _, t := range n.Props {
		if t.Type == dotType {
			dots++
		}
	}
	return dots
}

// Tuplet returns the irregular division value if the note is a tuplet.
func (n *Note) Tuplet() uint8 {
	if i, ok := n.Props.Find(tupletType); ok {
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
	_, ok := n.Props.Find(letRingType)
	return ok
}

func (n *Note) String() string {
	return fmt.Sprintf("%c%s", n.Name, n.Props)
}

// NewNote creates a note with properties.
func NewNote(symbol string, propList PropertyList) *Note {
	if _, ok := propList.Find(uintType); !ok {
		// Implicit quarter note.
		propList = append(propList, quarterNoteProp)
		sort.Sort(propList)
	}

	return &Note{
		Name:  []rune(symbol)[0],
		Props: propList,
	}
}

var quarterNoteProp = Property{
	Type: uintType,
	Lit:  []byte("4"),
}
