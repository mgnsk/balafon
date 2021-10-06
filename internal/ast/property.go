package ast

import (
	"fmt"
	"strings"

	"github.com/mgnsk/gong/internal/parser/token"
)

// PropertyList is a list of note properties.
type PropertyList []token.Token

func (p PropertyList) Len() int      { return len(p) }
func (p PropertyList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PropertyList) Less(i, j int) bool {
	a, ok := propOrder[p[i].Type]
	if !ok {
		panic(fmt.Sprintf("PropertyList: invalid token type '%s'", token.TokMap.StringType(p[i].Type)))
	}
	b, ok := propOrder[p[j].Type]
	if !ok {
		panic(fmt.Sprintf("PropertyList: invalid token type '%s'", token.TokMap.StringType(p[j].Type)))
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
				return nil, fmt.Errorf("duplicate note property '%s': '%s'", token.TokMap.Id(p.Type), p.IDValue())
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
		p[0] = *t
		copy(p[1:], props)
		return p, nil
	}

	return PropertyList{*t}, nil
}

var (
	sharpType   = token.TokMap.Type("sharp")
	flatType    = token.TokMap.Type("flat")
	accentType  = token.TokMap.Type("accent")
	ghostType   = token.TokMap.Type("ghost")
	uintType    = token.TokMap.Type("uint")
	dotType     = token.TokMap.Type("dot")
	tupletType  = token.TokMap.Type("tuplet")
	letRingType = token.TokMap.Type("letRing")
)

var propOrder = map[token.Type]int{
	sharpType:   0,
	flatType:    1,
	accentType:  2,
	ghostType:   3,
	uintType:    4,
	dotType:     5,
	tupletType:  6,
	letRingType: 7,
}
