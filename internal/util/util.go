package util

import (
	"fmt"
	"io"
	"os"
)

// Open returns os.Stdin when name is '-'
// or opens the file otherwise.
func Open(name string) (io.ReadCloser, error) {
	if name == "-" {
		return os.Stdin, nil
	} else if name == "" {
		return nil, fmt.Errorf("file argument or '-' for stdin required")
	}

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	return f, nil
}

// HandleExit catches a panic and exits with 1 if the panic contained an error.
func HandleExit() {
	if e := recover(); e != nil {
		if err, ok := e.(error); ok {
			fmt.Println(err)
			os.Exit(1)
		}
		panic(e)
	}
}
