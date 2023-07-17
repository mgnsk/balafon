package ast

import (
	"io"
	"sort"
	"strconv"

	"github.com/mgnsk/balafon/internal/parser/token"
	"github.com/mgnsk/balafon/internal/tokentype"
)

// PropertyList is a list of note properties.
type PropertyList []*token.Token

func (p PropertyList) Len() int      { return len(p) }
func (p PropertyList) Swap(i, j int) { p[i], p[j] = p[j], p[i] }
func (p PropertyList) Less(i, j int) bool {
	return p[i].Type < p[j].Type
}

// Find the property with specified type.
func (p PropertyList) Find(typ token.Type) *token.Token {
	for _, tok := range p {
		if tok.Type == typ {
			return tok
		}
	}
	return nil
}

func (p PropertyList) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	for _, t := range p {
		n += ew.Write(t.Lit)
	}

	return int64(n), ew.Flush()
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
		p := make(PropertyList, len(props)+1)
		p[0] = t
		copy(p[1:], props)
		sort.Sort(p)

		return p, nil
	}

	return PropertyList{t}, nil
}
