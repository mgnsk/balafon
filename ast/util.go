package ast

import (
	"bufio"
	"io"
	"strconv"
)

type errWriter struct {
	w   *bufio.Writer
	err error
}

func newErrWriter(w io.Writer) *errWriter {
	// bufio.NewWriter returns the underlying writer
	// if it is already a *bufio.Writer.
	return &errWriter{w: bufio.NewWriter(w)}
}

func (w *errWriter) Flush() error {
	if w.err != nil {
		return w.err
	}
	return w.w.Flush()
}

func (w *errWriter) Write(b []byte) int {
	if w.err != nil {
		return 0
	}
	n, err := w.w.Write(b)
	if err != nil {
		w.err = err
	}
	return n
}

func (w *errWriter) WriteByte(b byte) int {
	if w.err != nil {
		return 0
	}
	err := w.w.WriteByte(b)
	if err != nil {
		w.err = err
	}
	return 1

}
func (w *errWriter) WriteRune(r rune) int {
	if w.err != nil {
		return 0
	}
	n, err := w.w.WriteRune(r)
	if err != nil {
		w.err = err
	}
	return n
}

func (w *errWriter) WriteString(s string) int {
	if w.err != nil {
		return 0
	}
	n, err := w.w.WriteString(s)
	if err != nil {
		w.err = err
	}
	return n
}

func (w *errWriter) WriteInt(i int) int {
	return w.WriteString(strconv.Itoa(i))
}

func (w *errWriter) CopyFrom(wt io.WriterTo) int {
	if w.err != nil {
		return 0
	}
	n, err := wt.WriteTo(w.w)
	if err != nil {
		w.err = err
	}
	return int(n)
}
