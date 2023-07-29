package ast

import (
	"io"

	"github.com/mgnsk/balafon/internal/parser/token"
)

// Bar is a bar.
type Bar struct {
	Pos      token.Pos
	Name     string
	DeclList NodeList
}

func (b Bar) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(`:bar `)
	n += ew.WriteString(b.Name)
	n += ew.WriteString("\n")

	for _, stmt := range b.DeclList {
		n += ew.WriteString("\t")
		n += ew.WriteFrom(stmt)
	}

	n += ew.WriteString(":end")

	return int64(n), ew.Flush()
}

// NewBar creates a new bar.
func NewBar(pos token.Pos, name string, declList NodeList) Bar {
	return Bar{
		Pos:      pos,
		Name:     name,
		DeclList: declList,
	}
}
