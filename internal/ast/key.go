package ast

import (
	"io"
)

// CmdKey is a key change command.
type CmdKey struct {
	Key string
}

func (c CmdKey) WriteTo(w io.Writer) (int64, error) {
	ew := newErrWriter(w)
	var n int

	n += ew.WriteString(`:key `)
	n += ew.WriteString(c.Key)

	return int64(n), ew.Flush()
}

// NewCmdKey creates a key change command.
func NewCmdKey(key string) (CmdKey, error) {
	return CmdKey{
		Key: key,
	}, nil
}
