package ast

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/internal/parser/token"
)

// Track is a single track of notes.
type Track NoteList

// NewTrack creates a track.
func NewTrack(notes NoteList, inner interface{}) Track {
	var t Track
	t = append(t, notes...)
	if inner, ok := inner.(Track); ok {
		t = append(t, inner...)
	}
	return t
}

func (t Track) String() string {
	notes := make([]string, len(t))
	for i, note := range t {
		notes[i] = note.String()
	}
	return strings.Join(notes, " ")
}

// NoteList is a list of notes.
type NoteList []Note

// NewNoteList creates a note list by expanding the short syntax of a multi note
// and applying sorted properties to each individual note.
func NewNoteList(ident string, props interface{}) NoteList {
	var p NotePropertyList

	// Add implicit quarter note property if missing.
	switch props := props.(type) {
	case NotePropertyList:
		p = props
		if p.Find(uintType) == nil {
			p = append(p, uintToken)
		}
	default:
		p = NotePropertyList{uintToken}
	}

	sort.Stable(p)

	// Expand the short syntax of a multi note into individual notes
	// and apply the same properties to each individual note.
	notes := make(NoteList, len(ident))
	for i, r := range ident {
		notes[i] = Note{
			Name:  r,
			Props: p,
		}
	}

	return notes
}

// Note is a single note with sorted properties.
type Note struct {
	Name  rune
	Props NotePropertyList
}

// IsSharp reports whether the note is sharp.
func (n Note) IsSharp() bool {
	t := n.Props.Find(sharpType)
	return t != nil
}

// IsFlat reports whether the note is flat.
func (n Note) IsFlat() bool {
	t := n.Props.Find(flatType)
	return t != nil
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (n Note) Value() uint8 {
	t := n.Props.Find(uintType)
	if t == nil {
		panic("ast.Note: missing note value")
	}
	v, err := t.Int32Value()
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
	if t := n.Props.Find(tupletType); t != nil {
		// Extract the division number.
		// For example "3" for a triplet denoted by "/3".
		v, err := strconv.Atoi(string(t.Lit[1:]))
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

// NotePropertyList is a list of note properties. 3 types of properties exist:
// note value, dot and tuplet.
type NotePropertyList []*token.Token

func (p NotePropertyList) Len() int      { return len(p) }
func (p NotePropertyList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p NotePropertyList) Less(i, j int) bool {
	a, ok := propOrder[p[i].Type]
	if !ok {
		panic(fmt.Sprintf("NotePropertyList.Sort: invalid token type '%s'", token.TokMap.StringType(p[i].Type)))
	}
	b, ok := propOrder[p[j].Type]
	if !ok {
		panic(fmt.Sprintf("NotePropertyList.Sort: invalid token type '%s'", token.TokMap.StringType(p[j].Type)))
	}
	return a < b
}

// Find the property with specified type.
func (p NotePropertyList) Find(typ token.Type) *token.Token {
	for _, t := range p {
		if t.Type == typ {
			return t
		}
	}
	return nil
}

func (p NotePropertyList) String() string {
	var format strings.Builder
	for _, t := range p {
		format.Write(t.Lit)
	}
	return format.String()
}

// NewNotePropertyList creates a note property list.
func NewNotePropertyList(t *token.Token, inner interface{}) (NotePropertyList, error) {
	if t.Type == uintType {
		v, err := t.Int32Value()
		if err != nil {
			panic(err)
		}
		if uv := uint8(v); v < 1 || v > 128 || uv&(uv-1) != 0 {
			return nil, fmt.Errorf("note value property must be factor of 2 in range 1-128, value given: '%s'", t.IDValue())
		}
	}
	if props, ok := inner.(NotePropertyList); ok {
		for _, p := range props {
			if p.Type == t.Type && p.Type != dotType {
				return nil, fmt.Errorf("duplicate note property '%s': '%s'", token.TokMap.Id(p.Type), p.IDValue())
			}
		}
		return append(NotePropertyList{t}, props...), nil
	}
	return NotePropertyList{t}, nil
}

var (
	sharpType  = token.TokMap.Type("sharp")
	flatType   = token.TokMap.Type("flat")
	uintType   = token.TokMap.Type("uint")
	dotType    = token.TokMap.Type("dot")
	tupletType = token.TokMap.Type("tuplet")
)

var uintToken = &token.Token{
	Type: uintType,
	Lit:  []byte("4"),
}

var propOrder = map[token.Type]int{
	sharpType:  0,
	flatType:   1,
	uintType:   2,
	dotType:    3,
	tupletType: 4,
}
