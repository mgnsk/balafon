package player_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mgnsk/gong/interpreter"
	"github.com/mgnsk/gong/player"
	. "github.com/onsi/gomega"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/testdrv"
)

type testMessage struct {
	Timestamp time.Time
	Msg       midi.Message
}

func TestPlayerTiming(t *testing.T) {
	for _, tc := range []struct {
		timesig string
		tempo   uint
		input   string
	}{
		{
			timesig: "1 4",
			tempo:   60 * 10,
			input:   "c",
		},
		{
			timesig: "2 8",
			tempo:   60 * 10,
			input:   "c",
		},
		{
			timesig: "2 4",
			tempo:   120 * 10,
			input:   "c2",
		},
	} {
		t.Run(fmt.Sprintf("timesig %s tempo %d", tc.timesig, tc.tempo), func(t *testing.T) {
			defer midi.CloseDriver()
			out, _ := midi.OutPort(0)
			defer out.Close()
			in, _ := midi.InPort(0)
			defer in.Close()

			var msgs []testMessage

			midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
				msgs = append(msgs, testMessage{
					Timestamp: time.Now(),
					Msg:       msg,
				})
			})

			g := NewWithT(t)

			it := interpreter.New()

			g.Expect(it.Eval(fmt.Sprintf("timesig %s", tc.timesig))).To(Succeed())
			g.Expect(it.Eval(fmt.Sprintf("tempo %d", tc.tempo))).To(Succeed())
			g.Expect(it.Eval("assign c 60")).To(Succeed())
			g.Expect(it.Eval(tc.input)).To(Succeed())

			bars := it.Flush()
			g.Expect(bars).To(HaveLen(1))
			g.Expect(bars[0].Duration()).To(Equal(time.Second / 10))

			p := player.New(out)
			g.Expect(p.Play(context.TODO(), bars...)).To(Succeed())

			g.Expect(msgs).To(HaveLen(2))
			g.Expect(msgs[0].Msg.Type()).To(Equal(midi.NoteOnMsg))
			g.Expect(msgs[1].Msg.Type()).To(Equal(midi.NoteOffMsg))

			// Assert that bar duration is 1s (downscaled by 10x).
			g.Expect(msgs[len(msgs)-1].Timestamp).To(BeTemporally("~", msgs[0].Timestamp.Add(time.Second/10), 10*time.Millisecond))
		})
	}
}

func TestPlayerNoteOnNoteOffSorted(t *testing.T) {
	defer midi.CloseDriver()
	out, _ := midi.OutPort(0)
	defer out.Close()
	in, _ := midi.InPort(0)
	defer in.Close()

	var msgs []testMessage

	midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		msgs = append(msgs, testMessage{
			Timestamp: time.Now(),
			Msg:       msg,
		})
	})

	g := NewWithT(t)

	it := interpreter.New()

	input := `
channel 1
assign x 42
channel 2
assign k 36
tempo 600
timesig 4 4
bar "test"
	channel 1
	xxxx
	channel 2
	kkkk
end
play "test"
`

	g.Expect(it.Eval(input)).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(1))
	g.Expect(bars[0].Duration()).To(Equal(4 * time.Second / 10))

	p := player.New(out)
	g.Expect(p.Play(context.TODO(), bars...)).To(Succeed())

	g.Expect(msgs).To(HaveLen(16))

	// Assert that notes are played on simultaneously.
	g.Expect(msgs[0].Msg.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(msgs[1].Msg.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(msgs[0].Timestamp).To(BeTemporally("~", msgs[1].Timestamp))
	g.Expect(msgs[14].Msg.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(msgs[15].Msg.Type()).To(Equal(midi.NoteOffMsg))
	g.Expect(msgs[14].Timestamp).To(BeTemporally("~", msgs[15].Timestamp))
}

func TestPlayIncompleteBars(t *testing.T) {
	defer midi.CloseDriver()
	out, _ := midi.OutPort(0)
	defer out.Close()
	in, _ := midi.InPort(0)
	defer in.Close()

	var msgs []testMessage

	midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		msgs = append(msgs, testMessage{
			Timestamp: time.Now(),
			Msg:       msg,
		})
	})

	g := NewWithT(t)

	it := interpreter.New()

	input := `
timesig 4 4
assign c 60
tempo 600
c
c
`

	g.Expect(it.Eval(input)).To(Succeed())

	bars := it.Flush()
	g.Expect(bars).To(HaveLen(2))
	g.Expect(bars[0].Duration()).To(Equal(4 * time.Second / 10))
	g.Expect(bars[1].Duration()).To(Equal(4 * time.Second / 10))

	p := player.New(out)
	g.Expect(p.Play(context.TODO(), bars...)).To(Succeed())

	g.Expect(msgs).To(HaveLen(4))

	// Bar 1.
	g.Expect(msgs[0].Msg.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(msgs[1].Msg.Type()).To(Equal(midi.NoteOffMsg))
	// Bar 2.
	g.Expect(msgs[2].Msg.Type()).To(Equal(midi.NoteOnMsg))
	g.Expect(msgs[3].Msg.Type()).To(Equal(midi.NoteOffMsg))

	g.Expect(msgs[0].Timestamp.Add(time.Second / 10)).To(BeTemporally("~", msgs[1].Timestamp, 10*time.Millisecond))
	// 3s pause until second bar starts.
	g.Expect(msgs[1].Timestamp.Add(3 * time.Second / 10)).To(BeTemporally("~", msgs[2].Timestamp, 10*time.Millisecond))
	g.Expect(msgs[2].Timestamp.Add(time.Second / 10)).To(BeTemporally("~", msgs[3].Timestamp, 10*time.Millisecond))
}
