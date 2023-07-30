package main

import (
	"encoding/base64"
	"errors"
	"strings"
	"syscall/js"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/balafon"
)

func newErrorResponse(err error) map[string]interface{} {
	return map[string]interface{}{
		"kind":    "error",
		"message": err.Error(),
	}
}

func newSMFResponse(data string) map[string]interface{} {
	return map[string]interface{}{
		"kind":    "smf",
		"message": data,
	}
}

func convert(_ js.Value, args []js.Value) any {
	if len(args) != 1 {
		return newErrorResponse(errors.New("expected 1 argument"))
	}

	// TODO: CopyBytesToGo from array buffer

	input := []byte(args[0].String())

	spew.Dump(string(input))
	song, err := balafon.Convert(input)
	if err != nil {
		return newErrorResponse(err)
	}

	var buf strings.Builder
	bw := base64.NewEncoder(base64.StdEncoding, &buf)

	if _, err := song.WriteTo(bw); err != nil {
		return newErrorResponse(err)
	}

	if err := bw.Close(); err != nil {
		return newErrorResponse(err)
	}

	return newSMFResponse(buf.String())
}

func main() {
	js.Global().Set("convert", js.FuncOf(convert))
	select {} // keep running
}
