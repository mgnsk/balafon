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
	messages = make(chan testMessage, 4)
	midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		messages <- testMessage{
			ts:  time.Now(),
			msg: msg,
		}
	})
}

func TestShellTempo(t *testing.T) {
	g := NewWithT(t)

	it := interpreter.New()
	g.Expect(it.Eval(`
assign c 60
tempo 60

bar "one"
	timesig 1 4
	c
end

bar "two"
	timesig 2 4
	tempo 120
	c2
end

play "two"
play "one"
	`)).To(Succeed())

	sh := interpreter.NewShell(out)

	g.Expect(sh.Execute(it.Flush()...)).To(Succeed())

	var msgs []testMessage
L:
	for {
		select {
		case msg := <-messages:
			msgs = append(msgs, msg)
		default:
			break L
		}
	}

	g.Expect(msgs).To(HaveLen(4))
	g.Expect(msgs[0].msg.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(msgs[1].msg.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(msgs[2].msg.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(msgs[3].msg.Type()).To(Equal(midi.NoteOffMsg))

	g.Expect(msgs[1].ts).To(BeTemporally("~", msgs[0].ts.Add(1*time.Second), 10*time.Millisecond))
	g.Expect(msgs[2].ts).To(BeTemporally("~", msgs[1].ts, 10*time.Millisecond))
	g.Expect(msgs[3].ts).To(BeTemporally("~", msgs[2].ts.Add(1*time.Second), 10*time.Millisecond))
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
