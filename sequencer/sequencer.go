package sequencer

import (
	"time"

	"github.com/mgnsk/gong/constants"
	"github.com/mgnsk/gong/interpreter"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
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

			var ch, key, vel uint8
			if ev.Message.GetNoteOn(&ch, &key, &vel) {
				s.events = append(s.events, TrackEvent{
					Message:        smf.Message(midi.NoteOff(ch, key)),
					AbsTicks:       te.AbsTicks + ev.Duration,
					AbsNanoseconds: te.AbsNanoseconds + constants.TicksPerQuarter.Duration(s.tempo, ev.Duration).Nanoseconds(),
				})
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

// Play the accumulated sequence into a MIDI out port.
func (s *Sequencer) Play(out drivers.Out) error {
	var last int64
	for _, ev := range s.events {
		if ev.AbsNanoseconds > last {
			time.Sleep(time.Duration(ev.AbsNanoseconds - last))
			last = ev.AbsNanoseconds
		}
		if ev.Message.IsPlayable() {
			if err := out.Send(ev.Message); err != nil {
				return err
			}
		}
	}
	return nil
}

// ToSMF1 emits the accumulated SMF tracks.
func (s *Sequencer) ToSMF1() []TrackEvent {
	return s.events
}

// New creates an SMF sequencer.
func New() *Sequencer {
	return &Sequencer{
		tempo: constants.DefaultTempo,
	}
}
