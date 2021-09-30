package ast

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// Command is a control command.
type Command struct {
	Name string
	Args []string
}

func (c Command) String() string {
	return fmt.Sprintf("%s %s", c.Name, strings.Join(c.Args, " "))
}

// Uint8Args parses the command arguments as uint8 slice.
func (c Command) Uint8Args() []uint8 {
	args := make([]uint8, 0, 2)
	for _, a := range c.Args {
		v, err := strconv.Atoi(a)
		if err != nil {
			panic(err)
		}
		args = append(args, uint8(v))
	}
	return args
}

// Uint32Args parses the command arguments as uint32 slice.
func (c Command) Uint32Args() []uint32 {
	args := make([]uint32, 0, 2)
	for _, a := range c.Args {
		v, err := strconv.Atoi(a)
		if err != nil {
			panic(err)
		}
		args = append(args, uint32(v))
	}
	return args
}

// NewCommand creates a command from name and optional arguments.
func NewCommand(name string, args ...string) (Command, error) {
	c := Command{
		Name: name,
		Args: make([]string, len(args)),
	}
	for i, arg := range args {
		switch name {
		case "tempo":
			if err := validateRange(name, arg, 1, math.MaxUint16); err != nil {
				return Command{}, err
			}
		case "channel":
			if err := validateRange(name, arg, 0, 15); err != nil {
				return Command{}, err
			}
		case "velocity":
			fallthrough
		case "program":
			fallthrough
		case "control":
			if err := validateRange(name, arg, 0, 127); err != nil {
				return Command{}, err
			}
		}
		c.Args[i] = arg
	}
	return c, nil
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
