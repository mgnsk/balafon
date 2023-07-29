package ast

import (
	"io"
)

// BlockComment is a block comment.
type BlockComment struct {
	Text string
}

func (c BlockComment) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(c.Text)

	return int64(n), ew.Flush()
}

// NewBlockComment creates a new block comment.
func NewBlockComment(text string) BlockComment {
	return BlockComment{
		Text: text,
	}
}
