package interpreter_test

import (
	"testing"

	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
)

func TestShell(t *testing.T) {
	g := NewWithT(t)

	sh := interpreter.NewShell()
	sh.Execute(`
channel 1
assign c 60

channel 2
assign c 60

timesig 1 4

bar "test"
	channel 1
	c

	channel 2
	c
end
play "test"
`)

	_ = g
}
