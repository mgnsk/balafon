//go:build js && wasm

package main

import (
	"bytes"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/mgnsk/balafon"
)

func newResponse(written int, err error) map[string]interface{} {
	if err != nil {
		var (
			msg string
			pos balafon.Pos
		)

		if perr := new(balafon.ParseError); errors.As(err, &perr) {
			msg = perr.Error()
			pos = perr.ErrorToken.Pos
		} else if perr := new(balafon.EvalError); errors.As(err, &perr) {
			msg = perr.Error()
			pos = perr.Pos
		} else {
			panic(err)
		}

		return map[string]interface{}{
			"err": msg,
			"pos": map[string]interface{}{
				"offset": pos.Offset,
				"line":   pos.Line,
				"column": pos.Column,
			},
		}
	}

	return map[string]interface{}{
		"written": written,
	}
}

var buf bytes.Buffer

func convert(_ js.Value, args []js.Value) any {
	if len(args) != 2 {
		panic("expected 2 argument")
	}

	input := []byte(args[1].String())

	buf.Reset()
	if err := balafon.ToXML(&buf, input); err != nil {
		return newResponse(0, err)
	}

	if n := js.CopyBytesToJS(args[0], buf.Bytes()); n != buf.Len() {
		panic(fmt.Errorf("copied: %d, expected: %d bytes", n, buf.Len()))
	}

	return newResponse(buf.Len(), nil)
}

func main() {
	js.Global().Set("convert", js.FuncOf(convert))
	select {} // keep running
}
