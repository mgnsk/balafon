package ast

import (
	"strings"
)

// Bar is a bar.
type Bar struct {
	Name     string
	DeclList DeclList
}

func (b Bar) String() string {
	var format strings.Builder

	format.WriteString(`bar "`)
	format.WriteString(b.Name)
	format.WriteString("\" {\n")

	for _, stmt := range b.DeclList {
		format.WriteString("\t")
		format.WriteString(stmt.String())
		format.WriteString("\n")
	}

	format.WriteString("}")

	return format.String()
}

// NewBar creates a new bar.
func NewBar(name string, declList DeclList) Bar {
	return Bar{Name: name, DeclList: declList}
}
