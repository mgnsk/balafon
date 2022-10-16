package ast

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/internal/parser/token"
)

// Property is a note property.
type Property struct {
	Type token.Type
	Lit  []byte
}

// PropertyList is a list of note properties.
type PropertyList []Property

func (p PropertyList) Len() int      { return len(p) }
func (p PropertyList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PropertyList) Less(i, j int) bool {
	return p[i].Type < p[j].Type
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
		v, err := strconv.Atoi(string(t.Lit))
		if err != nil {
			return nil, err
		}
		if err := validateNoteValue(v); err != nil {
			return nil, err
		}
	}

	if props, ok := inner.(PropertyList); ok {
		for _, p := range props {
			switch {
			case p.Type == t.Type && p.Type != dotType:
				return nil, fmt.Errorf("duplicate note property '%s': '%c'", token.TokMap.Id(p.Type), p.Lit)
			case t.Type == accentType && p.Type == ghostType:
				return nil, fmt.Errorf("cannot add ghost property, note already has accentuated property")
			case t.Type == ghostType && p.Type == accentType:
				return nil, fmt.Errorf("cannot add accentuated property, note already has ghost property")
			case t.Type == sharpType && p.Type == flatType:
				return nil, fmt.Errorf("cannot add flat property, note already has sharp property")
			case t.Type == flatType && p.Type == sharpType:
				return nil, fmt.Errorf("cannot add sharp property, note already has flat property")
			}
		}

		p := make(PropertyList, len(props)+1)
		p[0] = Property{Type: t.Type, Lit: t.Lit}
		copy(p[1:], props)
		sort.Sort(p)
		return p, nil
	}

	return PropertyList{Property{Type: t.Type, Lit: t.Lit}}, nil
}
