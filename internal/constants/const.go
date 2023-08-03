package constants

import "gitlab.com/gomidi/midi/v2/smf"

// Constant definitions.
const (
	TicksPerQuarter smf.MetricTicks = 960
	TicksPerWhole                   = 4 * TicksPerQuarter
	DefaultTempo                    = 120
	DefaultVelocity                 = 100
	MaxValue                        = 127
	MaxBeatsPerBar                  = 128
	MinHumanChannel                 = 1
	MaxHumanChannel                 = 16
	MinVoice                        = 1
	MaxVoice                        = 4
)
