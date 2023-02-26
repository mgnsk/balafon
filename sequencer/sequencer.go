package sequencer

import (
	"github.com/mgnsk/gong/constants"
	"github.com/mgnsk/gong/interpreter"
	"gitlab.com/gomidi/midi/v2/smf"
	"golang.org/x/exp/slices"
)

// TrackEvent is an SMF track event.
type TrackEvent struct {
	Message        smf.Message
	AbsTicks       uint32
	AbsNanoseconds int64
}

// Sequencer is a MIDI sequencer.
type Sequencer struct {
	events         []TrackEvent
	pos            uint32
	tempo          float64
	absNanoseconds int64
}

// AddBars adds bars to te sequence.
func (s *Sequencer) AddBars(bars ...*interpreter.Bar) {
	for _, bar := range bars {
		for _, ev := range bar.Events {
			te := TrackEvent{
				Message:        ev.Message,
				AbsTicks:       s.pos + ev.Pos,
				AbsNanoseconds: s.absNanoseconds + constants.TicksPerQuarter.Duration(s.tempo, ev.Pos).Nanoseconds(),
			}

			s.events = append(s.events, te)

			var newTempo float64
			if ev.Message.GetMetaTempo(&newTempo) {
				s.tempo = newTempo
				continue
			}
		}

		ticks := bar.Cap()
		s.pos += ticks
		s.absNanoseconds += constants.TicksPerQuarter.Duration(s.tempo, ticks).Nanoseconds()
	}

	slices.SortStableFunc(s.events, func(a, b TrackEvent) bool {
		return a.AbsTicks < b.AbsTicks
	})
}

// Flush emits the accumulated SMF tracks.
func (s *Sequencer) Flush() []TrackEvent {
	events := make([]TrackEvent, len(s.events))
	copy(events, s.events)
	s.events = s.events[:0]
	return events
}

// New creates an SMF sequencer.
func New() *Sequencer {
	return &Sequencer{
		tempo: constants.DefaultTempo,
	}
}
