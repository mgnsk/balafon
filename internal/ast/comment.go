package ast

import (
	"io"

	"github.com/mgnsk/balafon/internal/parser/token"
)

// BlockComment is a block comment.
type BlockComment struct {
	Pos  token.Pos
	Text string
}

func (c BlockComment) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString("/*\n")
	n += ew.WriteString(c.Text)
	n += ew.WriteString("\n*/")

	return int64(n), ew.Flush()
}

// NewBlockComment creates a new block comment.
func NewBlockComment(pos token.Pos, text string) BlockComment {
	return BlockComment{
		Pos:  pos,
		Text: text,
	}
}
