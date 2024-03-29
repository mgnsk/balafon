package balafon

import (
	"fmt"

	parseError "github.com/mgnsk/balafon/internal/parser/errors"
	"github.com/mgnsk/balafon/internal/parser/token"
)

// ParseError is a parse error.
type ParseError = parseError.Error

// Pos is a token position.
type Pos = token.Pos

// EvalError is an eval error.
type EvalError struct {
	Err error
	Pos Pos
}

func (e *EvalError) Error() string {
	text := fmt.Sprintf("%d:%d: error: ", e.Pos.Line, e.Pos.Column)

	// See if the error token can provide us with the filename.
	switch src := e.Pos.Context.(type) {
	case token.Sourcer:
		text = src.Source() + ":" + text
	}

	return text + e.Err.Error()
}
