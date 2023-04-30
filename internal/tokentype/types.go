//go:generate go run tokentypes-gen/main.go

package tokentype

import (
	"github.com/mgnsk/balafon/internal/parser/token"
)

// Type is a language token type.
type Type struct {
	ID   string
	Type token.Type
}
