package ast

import (
	"io"

	"github.com/mgnsk/gong/internal/parser/token"
)

type LineComment string

func (c LineComment) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString("//")
	n += ew.WriteString(string(c))

	return int64(n), ew.Flush()
}

// func (c LineComment) String() string {
// 	return "//" + string(c)
// }

func NewLineComment(text *token.Token) LineComment {
	return LineComment(text.Lit[2 : len(text.Lit)-1])
}

type BlockComment string

func (c BlockComment) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString("/*")
	n += ew.WriteString(string(c))
	n += ew.WriteString("*/")

	return int64(n), ew.Flush()
}

// func (c BlockComment) String() string {
// 	return "/*" + string(c) + "*/"
// }

func NewBlockComment(text *token.Token) BlockComment {
	return BlockComment(text.Lit[2 : len(text.Lit)-2])
}
