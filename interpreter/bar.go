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
	Tempo   float64
}

// Len returns the bar's actual length in ticks.
func (b *Bar) Len() uint32 {
	durations := map[uint8]uint32{}
	for _, ev := range b.Events {
		durations[ev.Channel] += ev.Duration
	}
	var dur uint32
	for _, v := range durations {
		if v > dur {
			dur = v
		}
	}
	return dur
}

// Cap returns the bar's capacity in ticks.
func (b *Bar) Cap() uint32 {
	return uint32(b.TimeSig[0]) * (uint32(constants.TicksPerWhole) / uint32(b.TimeSig[1]))
}

// Duration returns the bar's duration.
func (b *Bar) Duration() time.Duration {
	return constants.TicksPerQuarter.Duration(b.Tempo, b.Len())
}

// func tickDuration(bpm float64) time.Duration {
// 	var d smf.MetricTicks
// 	d.Duration(bpm, 1)
// 	return time.Duration(float64(time.Minute) / bpm / float64(constants.TicksPerQuarter))
// }

// Export the bar to a sequencer.Bar.
// Note: this loses precision as ticks are converted to 32ths.
func (b *Bar) Export() sequencer.Bar {
	return sequencer.Bar{
		TimeSig: b.TimeSig,
		Events: func() sequencer.Events {
			events := make(sequencer.Events, len(b.Events))
			for i, ev := range b.Events {
				events[i] = &sequencer.Event{
					TrackNo:  int(ev.Channel),
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
