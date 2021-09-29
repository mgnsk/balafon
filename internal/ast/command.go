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
	c := Command{Name: name}
	for _, arg := range args {
		switch name {
		case "tempo":
			v, err := strconv.Atoi(arg)
			if err != nil {
				return Command{}, err
			}
			if v > math.MaxUint32 {
				return Command{}, fmt.Errorf("tempo argument must be in range 0-%d", math.MaxUint32)
			}
		case "channel":
			v, err := strconv.Atoi(arg)
			if err != nil {
				return Command{}, err
			}
			if v > 15 {
				return Command{}, fmt.Errorf("channel argument must be in range 0-15")
			}
		case "velocity":
			fallthrough
		case "program":
			fallthrough
		case "control":
			v, err := strconv.Atoi(arg)
			if err != nil {
				return Command{}, err
			}
			if v > 127 {
				return Command{}, fmt.Errorf("%s argument must be in range 0-127", name)
			}
		}
		c.Args = append(c.Args, arg)
	}
	return c, nil
}
