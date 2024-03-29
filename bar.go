package balafon

import (
	"fmt"
	"strings"
	"time"

	"github.com/mgnsk/balafon/internal/constants"
)

// Bar is a single bar of events.
type Bar struct {
	Events  []Event
	timeSig [2]uint8
}

// SetTimeSig sets the timesig for testing.
func (b *Bar) SetTimeSig(num, denom uint8) {
	b.timeSig = [2]uint8{num, denom}
}

func (b *Bar) String() string {
	var s strings.Builder

	s.WriteString(fmt.Sprintf("time: %d/%d", b.timeSig[0], b.timeSig[1]))

	if len(b.Events) > 0 {
		s.WriteString("\nevents:\n")
		for _, ev := range b.Events {
			s.WriteString(ev.String())
			s.WriteString("\n")
		}
	}

	return s.String()
}

// IsZeroDuration returns whether the bar consists of only zero duration events.
func (b *Bar) IsZeroDuration() bool {
	for _, ev := range b.Events {
		if ev.Duration > 0 {
			return false
		}
	}

	return true
}

// Cap returns the bar's capacity in ticks.
func (b *Bar) Cap() uint32 {
	return uint32(b.timeSig[0]) * (uint32(constants.TicksPerWhole) / uint32(b.timeSig[1]))
}

// Duration returns the bar's duration.
func (b *Bar) Duration(tempo float64) time.Duration {
	if b.IsZeroDuration() {
		return 0
	}

	return constants.TicksPerQuarter.Duration(tempo, b.Cap())
}
