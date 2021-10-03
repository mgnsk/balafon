package ast

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/internal/parser/token"
)

// Note: some list operations may be implemented with side effects.

// NoteList is a list of notes.
type NoteList []Note

func (l NoteList) String() string {
	notes := make([]string, len(l))
	for i, note := range l {
		notes[i] = note.String()
	}
	return strings.Join(notes, " ")
}

// NewNoteList creates a list of notes.
func NewNoteList(note interface{}, inner interface{}) (list NoteList) {
	switch n := note.(type) {
	case Note:
		list = NoteList{n}
	case NoteList:
		list = n
	default:
		panic("NewNoteList: invalid argument type")
	}
	if innerList, ok := inner.(NoteList); ok {
		list = append(list, innerList...)
	}
	return list
}

// NewNoteListFromGroup creates a note list from a group of notes with shared properties.
func NewNoteListFromGroup(notes, props interface{}) (NoteList, error) {
	noteList, ok := notes.(NoteList)
	if !ok {
		return nil, errors.New("note group must contain notes")
	}

	propList, ok := props.(PropertyList)
	if !ok {
		// Just a grouping of notes without properties, e.g. [cde].
		return noteList, nil
	}

	// Apply the properties to all notes, replacing duplicate
	// property types if needed.
	list := make(NoteList, len(noteList))
	for i, note := range noteList {
		for _, p := range propList {
			idx, ok := note.Props.Find(p.Type)
			if ok {
				note.Props[idx] = p
			} else {
				note.Props = append(note.Props, p)
			}
		}
		sort.Sort(note.Props)
		list[i] = note
	}

	return list, nil
}

// Note is a single note with sorted properties.
type Note struct {
	Props PropertyList
	Name  rune
}

// IsSharp reports whether the note is sharp.
func (n Note) IsSharp() bool {
	_, ok := n.Props.Find(sharpType)
	return ok
}

// IsFlat reports whether the note is flat.
func (n Note) IsFlat() bool {
	_, ok := n.Props.Find(flatType)
	return ok
}

// IsAccent reports whether the note is accentuated.
func (n Note) IsAccent() bool {
	_, ok := n.Props.Find(accentType)
	return ok
}

// IsGhost reports whether the note is a ghost note.
func (n Note) IsGhost() bool {
	_, ok := n.Props.Find(ghostType)
	return ok
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (n Note) Value() uint8 {
	i, ok := n.Props.Find(uintType)
	if !ok {
		panic("ast.Note: missing note value")
	}
	v, err := n.Props[i].Int32Value()
	if err != nil {
		panic(err)
	}
	// TODO range validation.
	return uint8(v)
}

// Dots reports the number of dot properties in the note.
func (n Note) Dots() uint {
	dots := uint(0)
	for _, t := range n.Props {
		if t.Type == dotType {
			dots++
		}
	}
	return dots
}

// Tuplet returns the irregular division value if the note is a tuplet.
func (n Note) Tuplet() uint {
	if i, ok := n.Props.Find(tupletType); ok {
		// Extract the division number.
		// For example "3" for a triplet denoted by "/3".
		v, err := strconv.Atoi(string(n.Props[i].Lit[1:]))
		if err != nil {
			panic(err)
		}
		return uint(v)
	}
	return 0
}

func (n Note) String() string {
	return fmt.Sprintf("%c%s", n.Name, n.Props)
}

// NewNote creates a note with properties.
func NewNote(note string, props interface{}) Note {
	var propList PropertyList
	if list, ok := props.(PropertyList); ok {
		propList = list
	}

	if _, ok := propList.Find(uintType); !ok {
		// Implicit quarter note.
		propList = append(propList, token.Token{
			Type: uintType,
			Lit:  []byte("4"),
		})
	}

	sort.Sort(propList)

	return Note{
		Name:  []rune(note)[0],
		Props: propList,
	}
}

// PropertyList is a list of note properties.
type PropertyList []token.Token

func (p PropertyList) Len() int      { return len(p) }
func (p PropertyList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PropertyList) Less(i, j int) bool {
	a, ok := propOrder[p[i].Type]
	if !ok {
		panic(fmt.Sprintf("PropertyList.Sort: invalid token type '%s'", token.TokMap.StringType(p[i].Type)))
	}
	b, ok := propOrder[p[j].Type]
	if !ok {
		panic(fmt.Sprintf("PropertyList.Sort: invalid token type '%s'", token.TokMap.StringType(p[j].Type)))
	}
	return a < b
}

// Find the property with specified type.
func (p PropertyList) Find(typ token.Type) (int, bool) {
	for i, t := range p {
		if t.Type == typ {
			return i, true
		}
	}
	return 0, false
}

func (p PropertyList) String() string {
	var format strings.Builder
	for _, t := range p {
		format.Write(t.Lit)
	}
	return format.String()
}

// NewPropertyList creates a note property list.
func NewPropertyList(t *token.Token, inner interface{}) (PropertyList, error) {
	if t.Type == uintType {
		v, err := t.Int32Value()
		if err != nil {
			panic(err)
		}
		if uv := uint8(v); v < 1 || v > 128 || uv&(uv-1) != 0 {
			return nil, fmt.Errorf("note value property must be a power of 2 in range 1-128, value given: '%s'", t.IDValue())
		}
	}
	if props, ok := inner.(PropertyList); ok {
		for _, p := range props {
			if p.Type == t.Type && p.Type != dotType {
				return nil, fmt.Errorf("duplicate note property '%s': '%s'", token.TokMap.Id(p.Type), p.IDValue())
			}
			if t.Type == accentType && p.Type == ghostType {
				return nil, fmt.Errorf("cannot add ghost property, note already has accentuated property")
			}
			if t.Type == ghostType && p.Type == accentType {
				return nil, fmt.Errorf("cannot add accentuated property, note already has ghost property")
			}
			if t.Type == sharpType && p.Type == flatType {
				return nil, fmt.Errorf("cannot add flat property, note already has sharp property")
			}
			if t.Type == flatType && p.Type == sharpType {
				return nil, fmt.Errorf("cannot add sharp property, note already has flat property")
			}
		}
		p := make(PropertyList, len(props)+1)
		p[0] = *t
		copy(p[1:], props)
		return p, nil
	}
	return PropertyList{*t}, nil
}

var propOrder = map[token.Type]int{
	sharpType:  0,
	flatType:   1,
	accentType: 2,
	ghostType:  3,
	uintType:   4,
	dotType:    5,
	tupletType: 6,
}
