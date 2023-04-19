package lint

import (
	"github.com/mgnsk/balafon/interpreter"
)

// Lint the input script.
// TODO: implement lint error that can format itself
func Lint(script []byte) error {
	it := interpreter.New()

	return it.Eval(string(script))
}
