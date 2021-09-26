package ast

import (
	"fmt"

	"github.com/mgnsk/gong/internal/token"
)

// Assignment assigns a numeric value to a string name.
type Assignment struct {
	Name  string
	Value uint32
}

// NewAssignment creates an assignment from tokens.
func NewAssignment(name string, value *token.Token) (*Assignment, error) {
	num, err := value.Int64Value()
	if err != nil {
		return nil, err
	}
	return &Assignment{
		Name:  name,
		Value: uint32(num),
	}, nil
}

func (a *Assignment) String() string {
	return fmt.Sprintf("%s = %d", a.Name, a.Value)
}
