package ast

import (
	"io"
)

type Node interface {
	io.WriterTo
}

type NodeList []Node

func (list NodeList) WriteTo(w io.Writer) (n int64, err error) {
	ew := newErrWriter(w)

	for i, decl := range list {
		n += int64(ew.WriteFrom(decl))

		if i < len(list)-1 {
			n += int64(ew.Write([]byte("\n")))
		}
	}

	return n, ew.Flush()
}

func NewNodeList(stmt Node, inner NodeList) (song NodeList) {
	return append(NodeList{stmt}, inner...)
}

// Must returns the result or panics if err is not nil.
func Must[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}
