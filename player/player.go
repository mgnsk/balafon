package player

import (
	"time"

	"github.com/mgnsk/gong/sequencer"
	"gitlab.com/gomidi/midi/v2/drivers"
)

// Player is a MIDI player.
type Player struct {
	out drivers.Out
}

// New creates a new player.
func New(out drivers.Out) *Player {
	return &Player{
		out: out,
	}
}

// Play the events into the out port.
func (p *Player) Play(events ...sequencer.TrackEvent) error {
	var last int64
	for _, ev := range events {
		if ev.AbsNanoseconds > last {
			time.Sleep(time.Duration(ev.AbsNanoseconds - last))
			last = ev.AbsNanoseconds
		}
		if ev.Message.IsPlayable() {
			if err := p.out.Send(ev.Message); err != nil {
				return err
			}
		}
	}
	return nil
}
