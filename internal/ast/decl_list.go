package ast

import (
	"fmt"
	"strings"
)

type DeclList []fmt.Stringer

func (declList DeclList) String() string {
	var format strings.Builder

	for i, decl := range declList {
		format.WriteString(decl.String())
		if i < len(declList)-1 {
			format.WriteString("\n")
		}
	}

	return format.String()
}

func NewDeclList(stmt fmt.Stringer, inner DeclList) (song DeclList) {
	return append(DeclList{stmt}, inner...)
}
