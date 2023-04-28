package shell_test

import (
	"bytes"
	"io"
	"reflect"
	"sync"
	"testing"
	"unicode/utf8"
	"unsafe"

	"github.com/c-bata/go-prompt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/balafon/shell"
	. "github.com/onsi/gomega"
)

type consoleParser struct {
	lines chan []byte
}

func newConsoleParser() *consoleParser {
	return &consoleParser{
		lines: make(chan []byte, 1),
	}
}

func (*consoleParser) Setup() error { return nil }

func (*consoleParser) TearDown() error { return nil }

func (*consoleParser) GetWinSize() *prompt.WinSize {
	return &prompt.WinSize{
		Row: 10,
		Col: 80,
	}
}

func (p *consoleParser) Read() ([]byte, error) {
	return <-p.lines, nil
}

func (p *consoleParser) Write(b []byte) {
	p.lines <- b
}

type posixWriter struct {
	prompt.VT100Writer
	w io.Writer
}

func (pw *posixWriter) Flush() error {
	// Create an addressable copy.
	rs := reflect.ValueOf(pw.VT100Writer)
	rs2 := reflect.New(rs.Type()).Elem()
	rs2.Set(rs)
	rf := rs2.FieldByName("buffer")
	rf = reflect.NewAt(rf.Type(), unsafe.Pointer(rf.UnsafeAddr())).Elem()

	b := rf.Bytes()

	spew.Dump(b)

	if _, err := pw.w.Write(b); err != nil {
		return err
	}

	return nil
}

func newPosixWriter(w io.Writer) prompt.ConsoleWriter {
	return &posixWriter{
		w: w,
	}
}

func TestBufferedPrompt(t *testing.T) {
	g := NewWithT(t)

	var (
		wg         sync.WaitGroup
		buf        bytes.Buffer
		mockWriter = newPosixWriter(&buf)
		mockParser = newConsoleParser()
	)

	wg.Add(1)
	go func() {
		defer wg.Done()

		pt := shell.NewBufferedPrompt(
			mockParser,
			mockWriter,
			func(in string) {
				panic("wait")
				buf.WriteString(in)
			},
			func(in prompt.Document) []prompt.Suggest {
				return nil
			},
		)

		pt.Run()
	}()

	_ = g

	// mockParser.Write([]byte("hello"))

	out := make([]byte, 1)
	g.Expect(utf8.EncodeRune(out, shell.EOT)).To(Equal(1))
	mockParser.Write(out)

	wg.Wait()

}
