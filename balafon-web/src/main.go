//go:build js && wasm

package main

import (
	"errors"
	"syscall/js"

	"github.com/mgnsk/balafon"
)

func newErrorResponse(err error) map[string]interface{} {
	return map[string]interface{}{
		"kind":    "error",
		"message": err.Error(),
	}
}

func newXMLResponse(data string) map[string]interface{} {
	return map[string]interface{}{
		"kind":    "xml",
		"message": data,
	}
}

func convert(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		return newErrorResponse(errors.New("expected 1 argument"))
	}

	// TODO: CopyBytesToGo from array buffer

	input := []byte(args[0].String())

	b, err := balafon.ToXML(input)
	if err != nil {
		return newErrorResponse(err)
	}

	return newXMLResponse(string(b))
}

func main() {
	js.Global().Set("convert", js.FuncOf(convert))
	select {} // keep running
}
