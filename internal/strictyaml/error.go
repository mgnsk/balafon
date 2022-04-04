package strictyaml

import (
	"fmt"

	"github.com/goccy/go-yaml"
)

// Error is a YAML unmarshaling error.
type Error struct {
	src     []byte
	path    *yaml.Path
	message string
}

// NewError creates a new YAML error.
func NewError(src []byte, path *yaml.Path, message string) *Error {
	return &Error{
		src:     src,
		path:    path,
		message: message,
	}
}

func (e *Error) Error() string {
	src, err := e.path.AnnotateSource(e.src, true)
	if err != nil {
		panic(err)
	}

	return fmt.Sprintf("%s:\n%s", e.message, string(src))
}
