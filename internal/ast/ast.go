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

	for _, decl := range list {
		n += int64(ew.WriteFrom(decl))
	}

	return n, ew.Flush()
}

func NewNodeList(stmt Node, inner ...Node) (list NodeList) {
	return append(NodeList{stmt}, inner...)
}

// WalkNotes flattens notes, groups and properties and calls f for every new note.
func WalkNotes(node Node, props PropertyList, f func(*Note) error) error {
	switch n := node.(type) {
	case NodeList:
		for _, node := range n {
			if err := WalkNotes(node, props, f); err != nil {
				return err
			}
		}
	case NoteGroup:
		for _, node := range n.Nodes {
			if err := WalkNotes(node, n.Props.Merge(props), f); err != nil {
				return err
			}
		}
	case *Note:
		noteCopy := *n
		noteCopy.Props = n.Props.Merge(props)
		return f(&noteCopy)
	}

	return nil
}

// Must returns the result or panics if err is not nil.
func Must[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}
