package ast

import (
	"fmt"
)

// Command is a control command.
// TODO zero or more arguments.
type Command struct {
	Name string
	Arg  string
}

func (c Command) String() string {
	return fmt.Sprintf("%s %s", c.Name, c.Arg)
}
