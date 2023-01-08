package interpreter

import (
	"fmt"

	"github.com/mgnsk/gong/ast"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	"gitlab.com/gomidi/midi/v2/sequencer"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Interpreter evaluates text input and emits MIDI events.
type Interpreter struct {
	parser    *parser.Parser
	astParser *Parser
	bars      []sequencer.Bar
	tempo     float64
}

// Eval the input.
func (it *Interpreter) Eval(input string) error {
	res, err := it.parser.Parse(lexer.NewLexer([]byte(input)))
	if err != nil {
		return err
	}

	declList, ok := res.(ast.NodeList)
	if !ok {
		return fmt.Errorf("invalid input, expected ast.NodeList")
	}

	bars, err := it.astParser.Parse(declList)
	if err != nil {
		return err
	}

	it.bars = append(it.bars, bars...)

	return nil
}

// Flush the parsed bar queue.
func (it *Interpreter) Flush() []sequencer.Bar {
	var (
		timesig      = [2]uint8{4, 4}
		metaBuffer   sequencer.Events
		playableBars []sequencer.Bar
	)

	// Concatenate bars that consist only of meta events.
	for _, bar := range it.bars {
		var newTempo float64
		hasTempoChange := false

		for _, ev := range bar.Events {
			if ev.Message.GetMetaTempo(&newTempo) {
				hasTempoChange = true
			}
		}

		switch isPlayable(bar.Events) {
		case true:
			if hasTempoChange && newTempo != it.tempo {
				// Restore the old tempo in next bar if not already pending.
				containsTempo := false
				for _, ev := range metaBuffer {
					var pendingTempo float64
					if ev.Message.GetMetaTempo(&pendingTempo) {
						containsTempo = true
						if pendingTempo != it.tempo {
							panic("interpreter: pending tempo invariant failure")
						}
					}
				}
				if !containsTempo {
					metaBuffer = append(metaBuffer, &sequencer.Event{
						TrackNo: 0,
						Message: smf.MetaTempo(it.tempo),
					})
				}
			}

			// Filter the meta buffer in place to skip overridden tempo.
			n := 0
			for _, ev := range metaBuffer {
				if hasTempoChange && ev.Message.Is(smf.MetaTempoMsg) {
					// Bar already has tempo change.
					// Keep in meta buffer for now.
					metaBuffer[n] = ev
					n++
				} else {
					// Prepend to bar.
					bar.Events = append(sequencer.Events{ev}, bar.Events...)
				}
			}

			barContainsTempo := false
			for _, ev := range bar.Events {
				if ev.Message.Is(smf.MetaTempoMsg) {
					barContainsTempo = true
					break
				}
			}

			if !barContainsTempo {
				// Always prepend tempo to each bar.
				// This is used by live shell where tempo
				bar.Events = append(sequencer.Events{&sequencer.Event{
					TrackNo: 0,
					Message: smf.MetaTempo(it.tempo),
				}}, bar.Events...)
			}

			// if not has tempo change

			metaBuffer = metaBuffer[:n]
			playableBars = append(playableBars, bar)
			timesig = bar.TimeSig

		case false:
			for _, ev := range bar.Events {
				ev.Message.GetMetaTempo(&it.tempo)
			}

			// Bar that consists of meta events only.
			metaBuffer = append(bar.Events, metaBuffer...)
			// TODO: ordering sometimes flips
			if !hasTempoChange {
				metaBuffer = append(metaBuffer, &sequencer.Event{
					TrackNo: 0,
					Message: smf.MetaTempo(it.tempo),
				})
			}
		}
	}

	// Append the last meta bar with the remaining meta events.
	if len(metaBuffer) > 0 {
		playableBars = append(playableBars, sequencer.Bar{
			TimeSig: timesig,
			Events:  metaBuffer,
		})
	}

	it.bars = it.bars[:0]

	return playableBars
}

// New creates an interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser:    parser.NewParser(),
		astParser: NewParser(),
		tempo:     120,
	}
}

func isPlayable(events sequencer.Events) bool {
	if len(events) == 0 {
		// A bar that consists of rests only.
		return true
	}
	for _, ev := range events {
		if ev.Duration > 0 {
			return true
		}
	}
	return false
}
