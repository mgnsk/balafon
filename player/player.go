package player

import (
	"context"
	"sort"
	"time"

	"github.com/mgnsk/gong/constants"
	"github.com/mgnsk/gong/interpreter"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Player plays back MIDI bars into an output port.
type Player struct {
	out        drivers.Out
	timer      *time.Timer
	events     []interpreter.Event
	currentPos uint32
}

// for _, ev := range bar.Events {
// }

// Play the bar.
func (p *Player) Play(ctx context.Context, bars ...*interpreter.Bar) error {
	for _, bar := range bars {
		p.events = p.events[:0]

		for _, ev := range bar.Events {
			p.events = append(p.events, ev)

			var ch, key, vel uint8
			if ev.Message.GetNoteOn(&ch, &key, &vel) {
				p.events = append(p.events, interpreter.Event{
					Message:  smf.Message(midi.NoteOff(ch, key)),
					Pos:      ev.Pos + ev.Duration,
					Duration: 0,
					Channel:  ch,
				})
			}
		}

		sort.Slice(p.events, func(i, j int) bool {
			return p.events[i].Pos < p.events[j].Pos
		})

		for _, ev := range p.events {
			if ev.Pos > p.currentPos {
				d := constants.TicksPerQuarter.Duration(bar.Tempo, ev.Pos-p.currentPos)
				p.timer.Reset(d)
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-p.timer.C:
				}
			}

			if ev.Message.IsPlayable() {
				if err := p.out.Send(ev.Message); err != nil {
					return err
				}
			}

			p.currentPos = ev.Pos
		}

		// Handle incomplete bars by sleeping until the end.
		if pause := bar.Cap() - p.currentPos; pause > 0 {
			d := constants.TicksPerQuarter.Duration(bar.Tempo, pause)
			p.timer.Reset(d)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-p.timer.C:
			}
		}
	}

	return nil
}

// New creates a new Player instance.
func New(out drivers.Out) *Player {
	timer := time.NewTimer(0)
	<-timer.C
	return &Player{
		out:   out,
		timer: timer,
	}
}
