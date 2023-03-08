package player

import (
	"time"

	"github.com/mgnsk/balafon/sequencer"
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
	if len(events) == 0 {
		return nil
	}

	// Play the first event without sleep.
	last := events[0].AbsNanoseconds

	for _, ev := range events {
		if delta := ev.AbsNanoseconds - last; delta > 0 {
			time.Sleep(time.Duration(delta))
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
