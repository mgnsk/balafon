package ast

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/internal/parser/token"
)

// Command is a control command.
type Command struct {
	Name string
	Args ArgumentList
}

func (c Command) String() string {
	return fmt.Sprintf("%s %s", c.Name, c.Args)
}

// RuneArg returns the i-th argument as a rune.
func (c Command) RuneArg(i int) rune {
	if i >= len(c.Args) {
		panic("invalid argument index")
	}
	return []rune(c.Args[i].IDValue())[0]
}

// StringValueArg returns the i-th string literal value argument.
func (c Command) StringValueArg(i int) string {
	if i >= len(c.Args) {
		panic("invalid argument index")
	}
	return c.Args[i].StringValue()
}

// Uint8Arg returns the i-th argument as a uint8.
func (c Command) Uint8Arg(i int) uint8 {
	if i >= len(c.Args) {
		panic("invalid argument index")
	}
	v, err := c.Args[i].Int32Value()
	if err != nil {
		panic(err)
	}
	return uint8(v)
}

// Uint32Arg returns the i-th argument as a uint32.
func (c Command) Uint32Arg(i int) uint32 {
	if i >= len(c.Args) {
		panic("invalid argument index")
	}
	v, err := c.Args[i].Int32Value()
	if err != nil {
		panic(err)
	}
	return uint32(v)
}

// NewCommand creates a command from name and optional arguments.
func NewCommand(name string, argList interface{}) (Command, error) {
	c := Command{
		Name: name,
	}

	var args ArgumentList
	if list, ok := argList.(ArgumentList); ok {
		args = list
	}

	switch name {
	case "assign":
		if len(args) != 2 || args[0].Type != singleNoteType || args[1].Type != uintType {
			return Command{}, fmt.Errorf("command '%s' requires 1 note argument and 1 numeric argument", name)
		}
		if len(args[0].IDValue()) != 1 {
			return Command{}, fmt.Errorf("command '%s' requires first argument to be a single character note", name)
		}
	case "tempo":
		fallthrough
	case "channel":
		fallthrough
	case "velocity":
		fallthrough
	case "program":
		if len(args) != 1 || args[0].Type != uintType {
			return Command{}, fmt.Errorf("command '%s' requires 1 numeric argument", name)
		}
	case "control":
		if len(args) != 2 || args[0].Type != uintType || args[1].Type != uintType {
			return Command{}, fmt.Errorf("command '%s' requires 2 numeric arguments", name)
		}
	case "bar":
		fallthrough
	case "play":
		if len(args) != 1 || args[0].Type != stringLitType {
			return Command{}, fmt.Errorf("command '%s' requires 1 string argument", name)
		}
	case "end":
		if len(args) != 0 {
			return Command{}, fmt.Errorf("command '%s' requires 0 arguments", name)
		}
	case "start":
		fallthrough
	case "stop":
		if len(args) != 0 {
			return Command{}, fmt.Errorf("command '%s' requires 0 arguments", name)
		}
	}

	for i, arg := range args {
		switch name {
		case "assign":
			if i == 1 {
				if err := validateRange(name, arg.IDValue(), 0, 127); err != nil {
					return Command{}, err
				}
			}
		case "tempo":
			if err := validateRange(name, arg.IDValue(), 1, math.MaxUint16); err != nil {
				return Command{}, err
			}
		case "channel":
			if err := validateRange(name, arg.IDValue(), 0, 15); err != nil {
				return Command{}, err
			}
		case "velocity":
			fallthrough
		case "program":
			fallthrough
		case "control":
			if err := validateRange(name, arg.IDValue(), 0, 127); err != nil {
				return Command{}, err
			}
		}
	}

	c.Args = args

	return c, nil
}

// ArgumentList is a list of command arguments.
type ArgumentList []*token.Token

func (l ArgumentList) String() string {
	args := make([]string, len(l))
	for i, arg := range l {
		args[i] = arg.IDValue()
	}
	return strings.Join(args, " ")
}

// NewArgumentList creates an argument list.
func NewArgumentList(arg *token.Token, inner interface{}) ArgumentList {
	innerArgs, ok := inner.(ArgumentList)
	if ok {
		l := make(ArgumentList, 1, len(innerArgs)+1)
		l[0] = arg
		l = append(l, innerArgs...)
		return l
	}
	return ArgumentList{arg}
}

func validateRange(name, arg string, min, max int) error {
	v, err := strconv.Atoi(arg)
	if err != nil {
		return err
	}
	if v < min || v > max {
		return fmt.Errorf("%s argument must be in range %d-%d", name, min, max)
	}
	return nil
}
