package ast

import (
	"fmt"
)

// Command is a control command.
type Command struct {
	Name string
	Arg  string
}

// NewCommand creates a command from string name and argument.
func NewCommand(name, arg string) (*Command, error) {
	return &Command{name, arg}, nil
}

func (c *Command) String() string {
	return fmt.Sprintf("%s %s", c.Name, c.Arg)
}
