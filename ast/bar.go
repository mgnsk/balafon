package ast

import (
	"io"

	"github.com/mgnsk/gong/internal/parser/token"
)

// Bar is a bar.
type Bar struct {
	Token    *token.Token
	Name     string
	DeclList NodeList
}

func (b Bar) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(`bar "`)
	n += ew.WriteString(b.Name)
	n += ew.WriteString("\"\n")

	for _, stmt := range b.DeclList {
		n += ew.WriteString("\t")
		n += ew.CopyFrom(stmt)
		n += ew.WriteString("\n")
	}

	n += ew.WriteString("end")

	return int64(n), ew.Flush()
}

// NewBar creates a new bar.
func NewBar(name *token.Token, declList NodeList) Bar {
	return Bar{
		Token:    name,
		Name:     name.StringValue(),
		DeclList: declList,
	}
}
