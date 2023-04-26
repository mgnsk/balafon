package lint

import (
	"errors"

	"github.com/mgnsk/balafon/interpreter"
)

// Lint the input file. TODO: proper API.
func Lint(filename, script string) error {
	it := interpreter.New()

	err := it.Eval(script)
	var perr *interpreter.ParseError
	if errors.As(err, &perr) {
		perr.Filename = filename
		return perr
	}

	return err
}
