package balafon

import (
	"fmt"
	"strings"

	"github.com/mgnsk/balafon/internal/constants"
	"gitlab.com/gomidi/midi/v2/smf"
	"golang.org/x/exp/slices"
)

// SMF is an SMF song.
type SMF []TrackEvent

func (song SMF) String() string {
	var s strings.Builder

	for _, ev := range song {
		s.WriteString(ev.String())
		s.WriteString("\n")
	}

	return s.String()
}

// TrackEvent is an SMF track event.
type TrackEvent struct {
	Message        smf.Message
	AbsTicks       uint32
	AbsNanoseconds int64
}

func (s *TrackEvent) String() string {
	return fmt.Sprintf("pos: %d ns: %d message: %s", s.AbsTicks, s.AbsNanoseconds, s.Message.String())
}

// Sequencer is a MIDI sequencer.
type Sequencer struct {
	song           SMF
	pos            uint32
	tempo          float64
	absNanoseconds int64
}

// AddBars adds bars to te sequence.
func (s *Sequencer) AddBars(bars ...*Bar) {
	for _, bar := range bars {
		for _, ev := range bar.Events {
			te := TrackEvent{
				Message:        ev.Message,
				AbsTicks:       s.pos + ev.Pos,
				AbsNanoseconds: s.absNanoseconds + constants.TicksPerQuarter.Duration(s.tempo, ev.Pos).Nanoseconds(),
			}

			s.song = append(s.song, te)

			var newTempo float64
			if ev.Message.GetMetaTempo(&newTempo) {
				s.tempo = newTempo
			}
		}

		ticks := bar.Cap()
		s.pos += ticks
		s.absNanoseconds += constants.TicksPerQuarter.Duration(s.tempo, ticks).Nanoseconds()
	}

	slices.SortStableFunc(s.song, func(a, b TrackEvent) bool {
		return a.AbsTicks < b.AbsTicks
	})
}

// Flush emits the accumulated SMF tracks.
func (s *Sequencer) Flush() SMF {
	song := make(SMF, len(s.song))
	copy(song, s.song)
	s.song = s.song[:0]
	return song
}

// NewSequencer creates an SMF sequencer.
func NewSequencer() *Sequencer {
	return &Sequencer{
		tempo: constants.DefaultTempo,
	}
}
