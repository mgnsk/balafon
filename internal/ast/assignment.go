package ast

import (
	"fmt"
	"strconv"
)

// Assignment assigns a numeric value to a string name.
type Assignment struct {
	Left  string
	Right string
}

// Uint32Value parses the right side of Assignment as uint32.
func (a Assignment) Uint32Value() uint32 {
	v, err := strconv.Atoi(a.Right)
	if err != nil {
		panic(err)
	}
	// TODO range
	return uint32(v)
}

func (a Assignment) String() string {
	return fmt.Sprintf("%s = %s", a.Left, a.Right)
}
