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
	if e.Pos.Context != nil {
		if src, ok := e.Pos.Context.(token.Sourcer); ok {
			return fmt.Sprintf("%s:%d:%d: error: %s", src.Source(), e.Pos.Line, e.Pos.Column, e.Err.Error())
		}
	}

	return fmt.Sprintf("%d:%d: error: %s", e.Pos.Line, e.Pos.Column, e.Err.Error())
}
