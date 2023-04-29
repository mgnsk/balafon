package ast

import (
	"io"
)

// Bar is a bar.
type Bar struct {
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
		n += ew.WriteString("\n")
	}

	n += ew.WriteString(":end")

	return int64(n), ew.Flush()
}

// NewBar creates a new bar.
func NewBar(name string, declList NodeList) Bar {
	return Bar{
		Name:     name,
		DeclList: declList,
	}
}
