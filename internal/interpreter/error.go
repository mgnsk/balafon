package interpreter

import "fmt"

// lineError is an error occurring on a specific line.
type lineError struct {
	nr  int
	err error
}

func (e lineError) Error() string {
	return fmt.Sprintf("[%d]: %s", e.nr, e.err)
}
