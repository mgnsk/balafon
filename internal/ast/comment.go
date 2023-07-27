package ast

import (
	"io"

	"github.com/mgnsk/balafon/internal/parser/token"
)

// LineComment is a line comment.
type LineComment struct {
	Pos  token.Pos
	Text string
}

func (c LineComment) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString("// ")
	n += ew.WriteString(c.Text)

	return int64(n), ew.Flush()
}

// NewLineComment creates a new line comment.
func NewLineComment(pos token.Pos, text string) LineComment {
	return LineComment{
		Pos:  pos,
		Text: text,
	}
}

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
