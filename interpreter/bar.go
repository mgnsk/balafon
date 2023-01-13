package interpreter

import (
	"time"

	"github.com/mgnsk/gong/constants"
	"gitlab.com/gomidi/midi/v2/sequencer"
)

// Bar is a single bar of events.
type Bar struct {
	Events  []Event
	TimeSig [2]uint8
}

// IsMeta returns whether the bar consists of only meta events.
func (b *Bar) IsMeta() bool {
	if len(b.Events) == 0 {
		// A bar that consists of rests only.
		// TODO: nil bar is not valid and is not emitted by Parser.
		return false
	}

	for _, ev := range b.Events {
		if ev.Duration > 0 {
			return false
		}
	}

	return true
}

// Cap returns the bar's capacity in ticks.
func (b *Bar) Cap() uint32 {
	return uint32(b.TimeSig[0]) * (uint32(constants.TicksPerWhole) / uint32(b.TimeSig[1]))
}

// Duration returns the bar's duration.
func (b *Bar) Duration(tempo float64) time.Duration {
	return constants.TicksPerQuarter.Duration(tempo, b.Cap())
}

// Export the bar to a sequencer.Bar.
// Note: this loses precision as ticks are converted to 32ths.
func (b *Bar) Export() sequencer.Bar {
	return sequencer.Bar{
		TimeSig: b.TimeSig,
		Events: func() sequencer.Events {
			events := make(sequencer.Events, len(b.Events))
			for i, ev := range b.Events {
				events[i] = &sequencer.Event{
					TrackNo: int(ev.Channel),
					// TODO
					Pos:      ticksTo32th(ev.Pos),
					Duration: ticksTo32th(ev.Duration),
					Message:  ev.Message,
				}
			}
			return events
		}(),
	}
}

// PrependMetaMessage prepends a message that has zero channel, pos and duration.
func (b *Bar) PrependMetaMessage(msg []byte) {
	b.Events = append(b.Events, Event{
		Channel:  0,
		Pos:      0,
		Duration: 0,
		Message:  msg,
	})
}

func ticksTo32th(ticks uint32) uint8 {
	return uint8(ticks / (uint32(constants.TicksPerQuarter) / 8))
}
