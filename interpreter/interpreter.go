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
	bars      []*sequencer.Bar
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
func (it *Interpreter) Flush() []*sequencer.Bar {
	var (
		timesig      [2]uint8
		metaBuffer   sequencer.Events
		playableBars []*sequencer.Bar
	)

	// Defer bars consisting of only meta events and concatenate them forward.
	for _, bar := range it.bars {
		timesig = bar.TimeSig

		switch isPlayable(bar.Events) {
		case true:
			var barEvs sequencer.Events
			barEvs = append(barEvs, metaBuffer...)
			barEvs = append(barEvs, bar.Events...)
			bar.Events = barEvs

			metaBuffer = metaBuffer[:0]
			playableBars = append(playableBars, bar)

		case false:
			// Bar that consists of meta events only.
			metaBuffer = append(bar.Events, metaBuffer...)
		}
	}

	if len(metaBuffer) > 0 {
		// Append the remaining meta events to the end of last bar.
		if len(playableBars) > 0 {
			lastBar := playableBars[len(playableBars)-1]
			pos := lastBar.Len()
			for _, ev := range metaBuffer {
				ev.Pos = pos
			}
			lastBar.Events = append(lastBar.Events, metaBuffer...)
		} else {
			playableBars = append(playableBars, &sequencer.Bar{
				TimeSig: timesig,
				Events:  metaBuffer,
			})
		}
	}

	it.bars = it.bars[:0]

	for _, bar := range playableBars {
		var newTempo float64
		hasTempo := false
		for _, ev := range bar.Events {
			if ev.Message.GetMetaTempo(&newTempo) {
				hasTempo = true
				break
			}
		}
		if hasTempo {
			it.tempo = newTempo
		} else {
			bar.Events = append(sequencer.Events{&sequencer.Event{
				Message: smf.MetaTempo(it.tempo),
			}}, bar.Events...)
		}
	}

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
