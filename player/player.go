package player

import (
	"fmt"
	"time"

	"github.com/mgnsk/gong/sequencer"
	"gitlab.com/gomidi/midi/v2/drivers"
)

// Player is a MIDI player.
type Player struct {
	out  drivers.Out
	last int64
}

// New creates a new player.
func New(out drivers.Out) *Player {
	return &Player{
		out: out,
	}
}

// Play the events into the out port.
func (p *Player) Play(events ...sequencer.TrackEvent) error {
	for _, ev := range events {
		if delta := ev.AbsNanoseconds - p.last; delta > 0 {
			fmt.Println(delta)
			time.Sleep(time.Duration(delta))
			p.last = ev.AbsNanoseconds
		}
		if ev.Message.IsPlayable() {
			if err := p.out.Send(ev.Message); err != nil {
				return err
			}
		}
	}
	return nil
}
