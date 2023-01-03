package interpreter_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/testdrv"
)

func TestShell(t *testing.T) {
	defer midi.CloseDriver()
	out, _ := midi.OutPort(0)
	in, _ := midi.InPort(0)

	start := time.Now()

	stop, _ := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		fmt.Printf("walltime: %s, timestamp: %d, msg: %s\n", time.Since(start), timestampms, msg.String())
	})

	defer stop()

	g := NewWithT(t)

	sh := interpreter.NewShell(out)
	sh.Execute(`
	assign c 60
	timesig 1 4
	tempo 60

	bar "test"
		c
	end
	play "test"
	`)

	_ = g
}

// TODO: move some tests here from interpreter?
