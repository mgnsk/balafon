//go:build js && wasm

package main

import (
	"bytes"
	"errors"
	"fmt"
	"syscall/js"

	"github.com/mgnsk/balafon"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/webmididrv"
)

func newConvertResponse(written int, err error) map[string]interface{} {
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
		return newConvertResponse(0, err)
	}

	if n := js.CopyBytesToJS(args[0], buf.Bytes()); n != buf.Len() {
		panic(fmt.Errorf("copied: %d, expected: %d bytes", n, buf.Len()))
	}

	return newConvertResponse(buf.Len(), nil)
}

func listPorts(_ js.Value, _ []js.Value) any {
	outs, err := drivers.Outs()
	if err != nil {
		return map[string]interface{}{
			"err": err.Error(),
		}
	}

	ports := make([]interface{}, len(outs))

	for i, out := range outs {
		ports[i] = map[string]interface{}{
			"number": out.Number(),
			"name":   out.String(),
		}
	}

	return map[string]interface{}{
		"ports": ports,
	}
}

var out drivers.Out

func selectPort(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		panic("expected 1 argument")
	}

	if out != nil {
		if err := out.Close(); err != nil {
			js.Global().Get("console").Call("error", err.Error())
		}
		out = nil
	}

	port, err := drivers.OutByNumber(args[0].Int())
	if err != nil {
		return map[string]interface{}{
			"err": err.Error(),
		}
	}

	out = port

	return map[string]interface{}{}
}

func play(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		panic("expected 1 argument")
	}

	it := balafon.New()
	if err := it.EvalString(args[0].String()); err != nil {
		return map[string]interface{}{
			"err": err.Error(),
		}
	}

	s := balafon.NewSequencer()
	s.AddBars(it.Flush()...)

	events := s.Flush()

	p := balafon.NewPlayer(out)

	go func() {
		if err := p.Play(events...); err != nil {
			js.Global().Get("console").Call("error", err.Error())
		}
	}()

	return map[string]interface{}{}
}

func main() {
	js.Global().Set("convert", js.FuncOf(convert))
	js.Global().Set("listPorts", js.FuncOf(listPorts))
	js.Global().Set("selectPort", js.FuncOf(selectPort))
	js.Global().Set("play", js.FuncOf(play))

	js.Global().Call("resolveStartedPromise")

	select {} // keep running
}
