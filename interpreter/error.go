package interpreter

import (
	"fmt"

	"github.com/mgnsk/balafon/internal/parser/token"
)

// ParseError is a parse error.
type ParseError struct {
	Filename       string
	Msg            string
	ErrorToken     *token.Token
	ExpectedTokens []string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s:%s", e.Filename, e.Msg)
}
