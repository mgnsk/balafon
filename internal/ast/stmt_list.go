package ast

import (
	"fmt"
	"strings"
)

type StmtList []fmt.Stringer

func (s StmtList) IndentString(n int) string {
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

func (s StmtList) String() string {
	return s.IndentString(0)
}

func NewStmtList(stmt fmt.Stringer, inner StmtList) (song StmtList) {
	return append(StmtList{stmt}, inner...)
}
