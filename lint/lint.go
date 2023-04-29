package lint

import (
	"errors"

	parseErrors "github.com/mgnsk/balafon/internal/parser/errors"
	"github.com/mgnsk/balafon/interpreter"
)

// Lint the input script.
func Lint(filename string, script []byte) error {
	it := interpreter.New()
	err := it.Eval(script)

	var e *parseErrors.Error
	if errors.As(err, &e) {
		return &Error{
			Filename:       filename,
			Err:            err.Error(),
			ErrorToken:     e.ErrorToken,
			ErrorSymbols:   e.ErrorSymbols,
			ExpectedTokens: e.ExpectedTokens,
		}
	}

	return err
}
