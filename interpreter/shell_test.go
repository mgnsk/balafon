package interpreter_test

import (
	"testing"
	"time"

	"github.com/mgnsk/gong/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/testdrv"
)

func TestShell(t *testing.T) {
	defer midi.CloseDriver()
	out, _ := midi.OutPort(0)
	in, _ := midi.InPort(0)

	var on, off time.Time

	stop, _ := midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		switch msg.Type() {
		case midi.NoteOnMsg:
			on = time.Now()
		case midi.NoteOffMsg:
			off = time.Now()
		}
	})
	defer stop()

	g := NewWithT(t)

	it := interpreter.New()
	g.Expect(it.Eval(`
assign c 60
timesig 1 4
tempo 60

bar "test"
	c
end
play "test"
	`)).To(Succeed())

	sh := interpreter.NewShell(out)
	g.Expect(sh.Execute(it.Flush()...)).To(Succeed())

	g.Expect(off).To(BeTemporally("~", on.Add(time.Second), 10*time.Millisecond))
}

// TODO: move some tests here from interpreter?
