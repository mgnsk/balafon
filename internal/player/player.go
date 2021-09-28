package player

import (
	"context"
	"sync"
	"time"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/scanner"
	"gitlab.com/gomidi/midi/v2"
)

// Player plays back scanner messages into a MIDI output port.
type Player struct {
	out          midi.Sender
	timer        *time.Timer
	tickDuration time.Duration
	once         sync.Once
	currentTick  uint64
}

// Play the message.
func (p *Player) Play(ctx context.Context, msg scanner.Message) error {
	if msg.Tempo > 0 {
		p.setTempo(msg.Tempo)
		return nil
	}

	p.once.Do(func() {
		p.currentTick = msg.Tick
	})

	if msg.Tick > p.currentTick {
		d := time.Duration(msg.Tick-p.currentTick) * p.tickDuration
		p.timer.Reset(d)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-p.timer.C:
		}
		p.currentTick = msg.Tick
	}

	if err := p.out.Send(msg.Msg.Data); err != nil {
		return err
	}

	return nil
}

func (p *Player) setTempo(bpm uint32) {
	p.tickDuration = time.Duration(float64(time.Minute) / float64(bpm) / float64(constants.TicksPerQuarter))
}

// New creates a new Player instance.
func New(out midi.Sender) *Player {
	timer := time.NewTimer(0)
	<-timer.C
	p := &Player{
		out:   out,
		timer: timer,
	}
	p.setTempo(120)
	return p
}
