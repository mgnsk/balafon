package player

import (
	"context"
	"sync"
	"time"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/interpreter"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Player plays back interpreted messages into a MIDI output port.
type Player struct {
	out          drivers.Out
	timer        *time.Timer
	tickDuration time.Duration
	once         sync.Once
	currentTick  uint32
}

// Play the message.
func (p *Player) Play(ctx context.Context, msg interpreter.Message) error {
	var bpm float64

	if smf.Message(msg.Message).GetMetaTempo(&bpm) {
		p.SetTempo(uint16(bpm))
		return nil
	}

	p.once.Do(func() {
		p.currentTick = msg.Tick
	})

	if msg.Tick > p.currentTick {
		p.timer.Reset(time.Duration(msg.Tick-p.currentTick) * p.tickDuration)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.timer.C:
		}
		p.currentTick = msg.Tick
	}

	return p.out.Send(msg.Message)
}

// SetTempo sets the current tempo.
func (p *Player) SetTempo(bpm uint16) {
	p.tickDuration = time.Duration(float64(time.Minute) / float64(bpm) / float64(constants.TicksPerQuarter))
}

// New creates a new Player instance.
func New(out drivers.Out) *Player {
	timer := time.NewTimer(0)
	<-timer.C
	p := &Player{
		out:   out,
		timer: timer,
	}
	p.SetTempo(constants.DefaultTempo)
	return p
}
