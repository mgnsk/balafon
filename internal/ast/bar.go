package ast

import (
	"strings"
)

// Bar is a bar.
type Bar struct {
	Name       string
	Statements StmtList
}

func (b Bar) String() string {
	var format strings.Builder

	format.WriteString(`bar "`)
	format.WriteString(b.Name)
	format.WriteString("\" {\n")

	for _, stmt := range b.Statements {
		format.WriteString("\t")
		format.WriteString(stmt.String())
		format.WriteString("\n")
	}

	format.WriteString("}")

	return format.String()
}

// NewBar creates a new bar.
func NewBar(name string, stmtList StmtList) Bar {
	return Bar{Name: name, Statements: stmtList}
}
