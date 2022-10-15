package ast

import (
	"fmt"
	"strings"
)

type DeclList []fmt.Stringer

func (s DeclList) IndentString(n int) string {
	var format strings.Builder
	for _, stmt := range s {
		for i := 0; i < n; i++ {
			format.WriteString("\t")
		}
		format.WriteString(stmt.String())
		format.WriteString("\n")
	}
	return format.String()
}

func (s DeclList) String() string {
	return s.IndentString(0)
}

func NewDeclList(stmt fmt.Stringer, inner DeclList) (song DeclList) {
	return append(DeclList{stmt}, inner...)
}
