package player

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/scanner"
	"gitlab.com/gomidi/midi/writer"
)

// Player plays back MIDI messages into a MIDI output port.
type Player struct {
	wr           *writer.Writer
	tickDuration time.Duration
	timer        *time.Timer
	once         sync.Once
	currentTick  uint64
}

// Play the message.
func (p *Player) Play(ctx context.Context, msg scanner.Message) error {
	if msg.Tempo > 0 {
		p.tickDuration = time.Duration(float64(time.Minute) / float64(msg.Tempo) / float64(constants.TicksPerQuarter))
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

	if err := p.wr.Write(msg.Msg); err != nil {
		return err
	}

	return nil
}

// New creates a new Player instance.
func New(w io.Writer) *Player {
	tempo := 120
	d := float64(time.Minute) / float64(tempo) / float64(constants.TicksPerQuarter)
	timer := time.NewTimer(0)
	<-timer.C
	return &Player{
		wr:           writer.New(w),
		tickDuration: time.Duration(d),
		timer:        timer,
	}
}
