package ast

import (
	"github.com/mgnsk/gong/internal/parser/token"
)

type LineComment string

func (c LineComment) String() string {
	return "//" + string(c)
}

func NewLineComment(text *token.Token) LineComment {
	return LineComment(text.Lit[2 : len(text.Lit)-1])
}

type BlockComment string

func (c BlockComment) String() string {
	return "/*" + string(c) + "*/"
}

func NewBlockComment(text *token.Token) BlockComment {
	return BlockComment(text.Lit[2 : len(text.Lit)-2])
}
