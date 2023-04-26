package lint

// Error is a lint error.
type Error struct {
	// TODO
	err error
}

func (e *Error) Error() string {
	return e.err.Error()
}
