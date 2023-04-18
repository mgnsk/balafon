package ast

import (
	"bytes"
	"io"

	"github.com/mgnsk/balafon/internal/parser/token"
)

// Bar is a bar.
type Bar struct {
	// TODO: token not used
	Token    *token.Token
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
		n += ew.CopyFrom(stmt)
		n += ew.WriteString("\n")
	}

	n += ew.WriteString(":end")

	return int64(n), ew.Flush()
}

// NewBar creates a new bar.
func NewBar(name *token.Token, declList NodeList) Bar {
	return Bar{
		Token:    name,
		Name:     string(bytes.TrimPrefix(name.Lit, []byte(":bar "))),
		DeclList: declList,
	}
}
