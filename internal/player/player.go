package player

import (
	"context"
	"sync"
	"time"

	"github.com/mgnsk/gong/internal/constants"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/sequencer"
)

// Player plays back interpreted messages into a MIDI output port.
type Player struct {
	out          drivers.Out
	timer        *time.Timer
	tickDuration time.Duration
	once         sync.Once
	currentPos   uint32
}

// Play the message.
func (p *Player) Play(ctx context.Context, bar *sequencer.Bar) error {
	return nil
	// var bpm float64

	// if smf.Message(event.Message).GetMetaTempo(&bpm) {
	// 	p.SetTempo(uint16(bpm))
	// 	return nil
	// }

	// p.once.Do(func() {
	// 	p.currentPos = uint32(bar.AbsTicks)
	// })

	// if event.Tick > p.currentPos {
	// 	p.timer.Reset(time.Duration(event.Tick-p.currentPos) * p.tickDuration)
	// 	select {
	// 	case <-ctx.Done():
	// 		return ctx.Err()
	// 	case <-p.timer.C:
	// 	}
	// 	p.currentPos = event.Tick
	// }

	// return p.out.Send(event.Message)
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
