package interpreter

import (
	"fmt"
	"sort"

	"github.com/mgnsk/gong/ast"
	"github.com/mgnsk/gong/constants"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Interpreter evaluates text input and emits MIDI events.
type Interpreter struct {
	parser    *parser.Parser
	astParser *Parser
	bars      []*Bar
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
func (it *Interpreter) Flush() []*Bar {
	var (
		timesig      [2]uint8
		metaBuffer   []Event
		playableBars []*Bar
	)

	// Defer bars consisting of only meta events and concatenate them forward.
	for _, bar := range it.bars {
		timesig = bar.TimeSig

		if bar.IsMeta() {
			// Bar that consists of meta events only.
			metaBuffer = append(metaBuffer, bar.Events...)
			continue
		}

		var barEvs []Event
		barEvs = append(barEvs, metaBuffer...)
		barEvs = append(barEvs, bar.Events...)
		bar.Events = barEvs

		metaBuffer = metaBuffer[:0]
		playableBars = append(playableBars, bar)
	}

	if len(metaBuffer) > 0 {
		// Append the remaining meta events to a new bar.
		playableBars = append(playableBars, &Bar{
			TimeSig: timesig,
			Events:  metaBuffer,
		})
	}

	it.bars = it.bars[:0]

	for _, bar := range playableBars {
		var newTempo float64
		hasTempo := false
		for _, ev := range bar.Events {
			// Get the last tempo event.
			if ev.Message.GetMetaTempo(&newTempo) {
				hasTempo = true
			}
		}
		if hasTempo {
			it.tempo = newTempo
		} else {
			bar.Events = append([]Event{{
				Message: smf.MetaTempo(it.tempo),
			}}, bar.Events...)
		}
		bar.Tempo = it.tempo
		sort.Slice(bar.Events, func(i, j int) bool {
			return bar.Events[i].Pos < bar.Events[j].Pos
		})
	}

	return playableBars
}

// New creates an interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser:    parser.NewParser(),
		astParser: NewParser(),
		tempo:     constants.DefaultTempo,
	}
}
