package ast

import (
	"fmt"
	"io"
	"sort"
	"strconv"

	"github.com/mgnsk/balafon/internal/constants"
	"github.com/mgnsk/balafon/internal/parser/token"
	"github.com/mgnsk/balafon/internal/tokentype"
)

// PropertyList is a list of note properties.
type PropertyList []*token.Token

func isUniqueProperty(typ token.Type) bool {
	switch typ {
	case tokentype.PropSharp,
		tokentype.PropFlat,
		tokentype.Uint,
		tokentype.PropTuplet,
		tokentype.PropLetRing:
		return true
	default:
		return false
	}
}

// Merge merges list into a copy of p and returns the merged result.
// Unique properties are overwritten while additive properties are added.
func (l PropertyList) Merge(list PropertyList) PropertyList {
	var result PropertyList
	result = append(result, l...)

	for _, prop := range list {
		if isUniqueProperty(prop.Type) {
			if idx := result.find(prop.Type); idx != -1 {
				result[idx] = prop
				continue
			}
		}
		result = append(result, prop)
	}

	return result
}

func (l PropertyList) Len() int      { return len(l) }
func (l PropertyList) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
func (l PropertyList) Less(i, j int) bool {
	return l[i].Type < l[j].Type
}

func (l PropertyList) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	for _, t := range l {
		n += ew.WriteBytes(t.Lit)
	}

	return int64(n), ew.Flush()
}

// NoteLen returns the note duration in ticks.
func (l PropertyList) NoteLen() uint32 {
	length := uint32(constants.TicksPerWhole) / uint32(l.Value())
	newLength := length
	dots := l.NumDot()
	for i := 0; i < dots; i++ {
		length /= 2
		newLength += length
	}
	if division := uint32(l.Tuplet()); division > 0 {
		newLength = newLength * 2 / division
	}
	return newLength
}

// IsSharp reports whether the list contains a sharp property.
func (l PropertyList) IsSharp() bool {
	return l.find(tokentype.PropSharp) != -1
}

// IsFlat reports whether the list contains a flat property.
func (l PropertyList) IsFlat() bool {
	return l.find(tokentype.PropFlat) != -1
}

// NumSharp returns the number of sharp signs.
func (l PropertyList) NumSharp() int {
	return l.countProps(tokentype.PropSharp)
}

// NumFlat reports the number of flat signs.
func (l PropertyList) NumFlat() int {
	return l.countProps(tokentype.PropFlat)
}

// NumStaccato reports the number of staccato properties.
func (l PropertyList) NumStaccato() int {
	return l.countProps(tokentype.PropStaccato)
}

// NumAccent reports the number of accent properties.
func (l PropertyList) NumAccent() int {
	return l.countProps(tokentype.PropAccent)
}

// NumMarcato reports the number of marcato properties.
func (l PropertyList) NumMarcato() int {
	return l.countProps(tokentype.PropMarcato)
}

// NumGhost reports the number of ghost properties.
func (l PropertyList) NumGhost() int {
	return l.countProps(tokentype.PropGhost)
}

// Value returns the note value (1th, 2th, 4th, 8th, 16th, 32th and so on).
func (l PropertyList) Value() uint8 {
	idx := l.find(tokentype.Uint)
	if idx == -1 {
		// Implicit quarter note.
		return 4
	}
	v, err := strconv.Atoi(string(l[idx].Lit))
	if err != nil {
		panic(err)
	}
	return uint8(v)
}

// NumDot reports the number of dot properties.
func (l PropertyList) NumDot() int {
	return l.countProps(tokentype.PropDot)
}

// Tuplet returns the irregular division value if the note is a tuplet.
func (l PropertyList) Tuplet() int {
	idx := l.find(tokentype.PropTuplet)
	if idx == -1 {
		return 0
	}
	// Trim the "/" prefix from tuplet token to get division number.
	v, err := strconv.Atoi(string(l[idx].Lit[1:]))
	if err != nil {
		panic(err)
	}
	return v
}

// IsLetRing reports whether the note must ring.
func (l PropertyList) IsLetRing() bool {
	return l.find(tokentype.PropLetRing) != -1
}

func (l PropertyList) has(types ...token.Type) bool {
	for _, typ := range types {
		if l.find(typ) != -1 {
			return true
		}
	}
	return false
}

func (l PropertyList) find(typ token.Type) int {
	for i, tok := range l {
		if tok.Type == typ {
			return i
		}
	}
	return -1
}

func (l PropertyList) countProps(typ token.Type) int {
	var count int
	for _, t := range l {
		if t.Type == typ {
			count++
		}
	}
	return count
}

// NewPropertyList creates a note property list.
func NewPropertyList(t *token.Token, inner interface{}) (PropertyList, error) {
	switch t.Type {
	case tokentype.Uint:
		v, err := strconv.Atoi(string(t.Lit))
		if err != nil {
			return nil, err
		}
		if err := validateNoteValue(v); err != nil {
			return nil, err
		}
	case tokentype.PropTuplet:
		v, err := strconv.Atoi(string(t.Lit[1:]))
		if err != nil {
			return nil, err
		}
		if err := validateTuplet(v); err != nil {
			return nil, err
		}
	}

	if props, ok := inner.(PropertyList); ok {
		switch t.Type {
		case tokentype.PropSharp, tokentype.PropFlat:
			if props.has(tokentype.PropSharp, tokentype.PropFlat) {
				return nil, fmt.Errorf("duplicate sharp or flat property")
			}
		}

		p := make(PropertyList, len(props)+1)
		p[0] = t
		copy(p[1:], props)
		sort.Sort(p)

		return p, nil
	}

	return PropertyList{t}, nil
}
