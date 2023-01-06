package interpreter_test

import (
	"testing"
	"time"

	"github.com/mgnsk/gong/interpreter"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/testdrv"
)

type testMessage struct {
	ts  time.Time
	msg midi.Message
}

var (
	out      drivers.Out
	in       drivers.In
	messages chan testMessage
)

func init() {
	out, _ = midi.OutPort(0)
	in, _ = midi.InPort(0)
	messages = make(chan testMessage, 2)
	midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		messages <- testMessage{
			ts:  time.Now(),
			msg: msg,
		}
	})
}

func TestShell(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()
	g.Expect(it.Eval(`
assign c 60
timesig 1 4
tempo 600

bar "test"
	c
end
play "test"
	`)).To(Succeed())

	sh := interpreter.NewShell(out)
	g.Expect(sh.Execute(it.Flush()...)).To(Succeed())

	on := <-messages
	off := <-messages

	g.Expect(on.msg.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(off.msg.Type()).To(Equal(midi.NoteOffMsg))

	g.Expect(off.ts).To(BeTemporally("~", on.ts.Add(time.Second/10), 10*time.Millisecond))
}

func TestUnplayableBarSkipped(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()
	g.Expect(it.Eval(`
tempo 60
timesig 1 4
velocity 25
channel 1

assign c 60
-
	`)).To(Succeed())

	sh := interpreter.NewShell(out)
	g.Expect(sh.Execute(it.Flush()...)).To(Succeed())

	g.Expect(messages).NotTo(Receive())
}
