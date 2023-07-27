package ast

import (
	"io"
	"regexp"
	"strings"
)

type Node interface {
	io.WriterTo
}

type RepeatTerminator []string

func (t RepeatTerminator) WriteTo(w io.Writer) (n int64, err error) {
	ew := newErrWriter(w)

	val := strings.Join(t, "")
	val = newlines.ReplaceAllLiteralString(val, "\n")

	n += int64(ew.WriteString(val))

	return n, ew.Flush()
}

func NewRepeatTerminator(terminator string, inner ...string) RepeatTerminator {
	return append(RepeatTerminator{terminator}, inner...)
}

type NodeList []Node

func (list NodeList) WriteTo(w io.Writer) (n int64, err error) {
	ew := newErrWriter(w)

	for _, decl := range list {
		n += int64(ew.WriteFrom(decl))
	}

	return n, ew.Flush()
}

func NewNodeList(stmt Node, inner ...Node) (song NodeList) {
	return append(NodeList{stmt}, inner...)
}

// Must returns the result or panics if err is not nil.
func Must[T any](result T, err error) T {
	if err != nil {
		panic(err)
	}
	return result
}

var newlines = regexp.MustCompile(`\n+`)
