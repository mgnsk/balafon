package balafon

import (
	"fmt"

	parseError "github.com/mgnsk/balafon/internal/parser/errors"
	"github.com/mgnsk/balafon/internal/parser/token"
)

// ParseError is a parse error.
type ParseError = parseError.Error

// EvalError is an eval error.
type EvalError struct {
	Err        error
	ErrorToken *token.Token
}

func (e *EvalError) Error() string {
	text := fmt.Sprintf("%d:%d: error: ", e.ErrorToken.Pos.Line, e.ErrorToken.Pos.Column)

	// See if the error token can provide us with the filename.
	switch src := e.ErrorToken.Pos.Context.(type) {
	case token.Sourcer:
		text = src.Source() + ":" + text
	}

	return text + e.Err.Error()
}
