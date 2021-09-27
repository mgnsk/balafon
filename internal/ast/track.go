package ast

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/internal/token"
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
type NoteList []*Note

// NewNoteList creates a note list by expanding the short syntax of a multi note
// and applying sorted properties to each individual note.
func NewNoteList(ident string, props interface{}) NoteList {
	var p PropertyList

	switch props := props.(type) {
	case PropertyList:
		props.Sort()
		if len(props) == 0 || props[0].Type != token.TokMap.Type("uint64") {
			// Implicit quarter note.
			p = append(PropertyList{
				&token.Token{
					Type: token.TokMap.Type("uint64"),
					Lit:  []byte("4"),
				},
			}, props...)
		} else {
			p = props
		}
	default:
		p = PropertyList{
			&token.Token{
				Type: token.TokMap.Type("uint64"),
				Lit:  []byte("4"),
			},
		}
	}

	// Expand the short syntax of a multi note into individual notes
	// and apply the properties to each individual note.
	notes := make(NoteList, len(ident))
	for i, r := range ident {
		n := &Note{
			Name:  string(r),
			Props: p,
		}
		notes[i] = n
	}

	return notes
}

// Note is a single note with sorted properties.
type Note struct {
	Name  string
	Props PropertyList
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (n *Note) Value() uint8 {
	v, err := n.Props[0].Int32Value()
	if err != nil {
		panic(err)
	}

	return uint8(v)
}

// IsDot reports whether the not is a dotted note.
func (n *Note) IsDot() bool {
	return len(n.Props) >= 2 && n.Props[1].Type == token.TokMap.Type("dot")
}

// Tuplet returns the irregular division value if the note is a tuplet.
func (n *Note) Tuplet() uint8 {
	// If the note is a dotted note, the tuplet property is at index 2,
	// otherwise it is at index 1.
	for _, t := range n.Props {
		if t.Type == token.TokMap.Type("tuplet") {
			// Extract the division number.
			// For example "3" for a triplet denoted by "/3".
			v, err := strconv.Atoi(string(t.Lit[1:]))
			if err != nil {
				panic(err)
			}
			return uint8(v)
		}
	}
	return 0
}

func (n *Note) String() string {
	values := make([]string, len(n.Props))
	for i, p := range n.Props {
		values[i] = p.String()
	}
	return fmt.Sprintf("%s%s", n.Name, n.Props)
}

// PropertyList is a list of note properties. 3 types of properties exist:
// note value, dot and tuplet.
type PropertyList []*token.Token

// Sort the properties in the order of value, dot, tuplet.
func (p PropertyList) Sort() {
	sort.Slice(p, func(i, j int) bool {
		return p[i].Type < p[j].Type
	})
}

func (p PropertyList) String() string {
	var format strings.Builder
	for _, t := range p {
		format.Write(t.Lit)
	}
	return format.String()
}

// NewPropertyList creates a note property list.
func NewPropertyList(t *token.Token, inner interface{}) PropertyList {
	if props, ok := inner.(PropertyList); ok {
		return append(PropertyList{t}, props...)
	}
	return PropertyList{t}
}
