package interpreter

import (
	"time"

	"github.com/mgnsk/balafon/internal/constants"
)

// Bar is a single bar of events.
type Bar struct {
	Events  []Event
	TimeSig [2]uint8
}

// IsZeroDuration returns whether the bar consists of only zero duration events.
func (b *Bar) IsZeroDuration() bool {
	if len(b.Events) == 0 {
		// A bar that consists of rests only.
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
	if b.IsZeroDuration() {
		return 0
	}

	return constants.TicksPerQuarter.Duration(tempo, b.Cap())
}
