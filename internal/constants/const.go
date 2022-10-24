package constants

import "gitlab.com/gomidi/midi/v2/smf"

// Constant definitions.
const (
	TicksPerQuarter smf.MetricTicks = 960
	TicksPerWhole   smf.MetricTicks = 4 * TicksPerQuarter
	DefaultTempo                    = 120
	DefaultVelocity                 = 100
	MinValue                        = 0
	MaxValue                        = 127
	MaxBeatsPerBar                  = 128
	MaxChannel                      = 15
)
