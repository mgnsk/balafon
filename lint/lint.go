package lint

import (
	"errors"

	"github.com/mgnsk/balafon"
	parseErrors "github.com/mgnsk/balafon/internal/parser/errors"
)

// Lint the input script.
func Lint(filename string, script []byte) error {
	it := balafon.New()
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
