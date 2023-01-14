package interpreter

import (
	"fmt"
	"sort"

	"github.com/mgnsk/gong/ast"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
)

// Interpreter evaluates text input and emits MIDI events.
type Interpreter struct {
	parser    *parser.Parser
	astParser *Parser
	bars      []*Bar
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
		buf          []Event
		playableBars []*Bar
	)

	// Defer virtual bars and concatenate them forward.
	for _, bar := range it.bars {
		timesig = bar.TimeSig

		if bar.IsVirtual() {
			buf = append(buf, bar.Events...)
			continue
		}

		var barEvs []Event
		barEvs = append(barEvs, buf...)
		barEvs = append(barEvs, bar.Events...)
		bar.Events = barEvs

		buf = buf[:0]
		playableBars = append(playableBars, bar)
	}

	if len(buf) > 0 {
		// Append the remaining meta events to a new bar.
		playableBars = append(playableBars, &Bar{
			TimeSig: timesig,
			Events:  buf,
		})
	}

	it.bars = it.bars[:0]

	for _, bar := range playableBars {
		sort.SliceStable(bar.Events, func(i, j int) bool {
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
	}
}
