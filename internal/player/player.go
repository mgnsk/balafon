package player

import (
	"io"
	"sync"
	"time"

	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/scanner"
	"gitlab.com/gomidi/midi/writer"
)

const tempo = 120

// Player plays back MIDI messages into a MIDI output port.
type Player struct {
	wr           *writer.Writer
	tickDuration time.Duration
	once         sync.Once
	currentTick  uint64
}

// Play the message.
func (p *Player) Play(msg scanner.Message) error {
	if msg.Tempo > 0 {
		p.tickDuration = time.Duration(float64(time.Minute) / float64(msg.Tempo) / float64(constants.TicksPerQuarter))
		return nil
	}

	p.once.Do(func() {
		p.currentTick = msg.Tick
	})

	if msg.Tick > p.currentTick {
		d := time.Duration(msg.Tick-p.currentTick) * p.tickDuration
		time.Sleep(d)
	}

	if err := p.wr.Write(msg.Msg); err != nil {
		return err
	}

	p.currentTick = msg.Tick

	return nil
}

// New creates a new Player instance.
func New(w io.Writer) *Player {
	d := float64(time.Minute) / float64(tempo) / float64(constants.TicksPerQuarter)
	return &Player{
		wr:           writer.New(w),
		tickDuration: time.Duration(d),
	}
}
