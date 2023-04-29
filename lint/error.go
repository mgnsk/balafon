package lint

import (
	"fmt"

	parseErrors "github.com/mgnsk/balafon/internal/parser/errors"
	"github.com/mgnsk/balafon/internal/parser/token"
)

// Error is a lint error.
type Error struct {
	Filename       string
	Err            string
	ErrorToken     *token.Token
	ErrorSymbols   []parseErrors.ErrorSymbol
	ExpectedTokens []string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s:%s", e.Filename, e.Err)
}
