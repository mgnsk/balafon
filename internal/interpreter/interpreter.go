package interpreter

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/davecgh/go-spew/spew"
	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/constants"
	parseError "github.com/mgnsk/gong/internal/parser/errors"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	"github.com/mgnsk/gong/internal/parser/token"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/sequencer"
	"gitlab.com/gomidi/midi/v2/smf"
)

type midiKey struct {
	channel uint8
	note    rune
}

// Interpreter evaluates MIDI messages from text input.
type Interpreter struct {
	parser *parser.Parser
	keymap map[midiKey]uint8
	// ringing     map[midiKey]struct{}
	bars        map[string]sequencer.Events
	curChannel  uint8
	curVelocity uint8
}

func (it *Interpreter) clone() *Interpreter {
	newIt := New()

	// Use the same maps, assign and bar are not allowed in bars.
	newIt.keymap = it.keymap
	newIt.bars = it.bars
	newIt.curChannel = it.curChannel
	newIt.curVelocity = it.curVelocity

	return newIt
}

var sugInsideBar = []string{
	"tempo",
	"timesig",
	"channel",
	"velocity",
	"program",
	"control",
	"end",
}

// TODO: generate those in init()
// by parsing a parse error?
var sugOutsideBar = []string{
	"assign",
	"tempo",
	"channel",
	"velocity",
	"program",
	"control",
	"bar",
	"start",
	"stop",
}

// TODO:

// TODO: parseOption: fillBarSilence - fills bars with

func (it *Interpreter) Parse(input string) (ast.DeclList, error) {
	res, err := it.parser.Parse(lexer.NewLexer([]byte(input)))
	if err != nil {
		return nil, err
	}

	declList, ok := res.(ast.DeclList)
	if !ok {
		return nil, fmt.Errorf("invalid input, expected ast.DeclList")
	}

	return declList, nil
}

func (it *Interpreter) Eval(input string) (*sequencer.Song, error) {
	declList, err := it.Parse(input)
	if err != nil {
		return nil, err
	}

	return it.EvalAST(declList)
}

func (it *Interpreter) EvalAST(declList ast.DeclList) (*sequencer.Song, error) {
	var (
		song   = sequencer.New()
		buffer sequencer.Events
	)

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.Bar:
			if _, ok := it.bars[decl.Name]; ok {
				return nil, fmt.Errorf("bar '%s' already defined", decl.Name)
			}

			events, err := it.clone().parse(decl.DeclList)
			if err != nil {
				return nil, err
			}

			it.bars[decl.Name] = events

		case ast.CmdPlay:
			events, ok := it.bars[string(decl)]
			if !ok {
				return nil, fmt.Errorf("unknown bar '%s'", decl.String())
			}

			buffer = append(buffer, events...)

		default:
			events, err := it.parse(ast.DeclList{decl})
			if err != nil {
				return nil, err
			}

			buffer = append(buffer, events...)
		}

		if getDuration(buffer) > 0 {
			song.AddBar(createBar(buffer))
			buffer = buffer[:0]
		}
	}

	if len(buffer) > 0 {
		song.AddBar(createBar(buffer))
	}

	return song, nil
}

func createBar(events sequencer.Events) sequencer.Bar {
	bar := sequencer.Bar{
		Events: make(sequencer.Events, 0, len(events)),
	}

	for _, ev := range events {
		if num, denom, ok := getMeter(ev); ok {
			bar.TimeSig = [2]uint8{num, denom}
		} else {
			bar.Events = append(bar.Events, ev)
		}
	}

	return bar
}

func (it *Interpreter) parse(declList ast.DeclList) (sequencer.Events, error) {
	var events sequencer.Events

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.CmdAssign:
			if err := it.assign(it.curChannel, decl.Note, decl.Key); err != nil {
				return nil, err
			}

		case ast.CmdTempo:
			events = append(events, &sequencer.Event{
				Message: smf.MetaTempo(float64(decl)),
			})

		case ast.CmdTimeSig:
			events = append(events, &sequencer.Event{
				Message: smf.MetaMeter(decl.Num, decl.Denom),
			})

		case ast.CmdChannel:
			it.curChannel = uint8(decl)

		case ast.CmdVelocity:
			it.curVelocity = uint8(decl)

		case ast.CmdProgram:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.ProgramChange(it.curChannel, uint8(decl))),
			})

		case ast.CmdControl:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.ControlChange(it.curChannel, decl.Control, decl.Parameter)),
			})

		case ast.CmdStart:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.Start()),
			})

		case ast.CmdStop:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.Stop()),
			})

		case ast.NoteList:
			noteEvents, err := it.parseNoteList(decl)
			if err != nil {
				return nil, err
			}

			events = append(events, noteEvents...)

		case ast.LineComment:
		case ast.BlockComment:
		default:
			panic(fmt.Sprintf("parseBar: invalid token %T", decl))
		}
	}

	return events, nil
}

func getDuration(events sequencer.Events) uint32 {
	var d uint32
	for _, ev := range events {
		d += uint32(ev.Duration)
	}
	return d
}

func negIndex(tokens []*token.Token, i int) *token.Token {
	if len(tokens) > 0 {
		return tokens[len(tokens)-1]
	}
	return nil
}

// Suggest returns suggestions for the next input.
// It is not safe to call Suggest concurrently
// with Eval.
func (it *Interpreter) Suggest(buffer string) []string {
	var (
		once          sync.Once
		currentTokens []token.Type
	)

	getLastNTokens := func(n int) []token.Type {
		once.Do(func() {
			lex := lexer.NewLexer([]byte(buffer))
			for {
				tok := lex.Scan()
				if tok.Type == token.EOF {
					break
				}
				currentTokens = append(currentTokens, tok.Type)
			}
		})
		if from := len(currentTokens) - n; from >= 0 {
			return currentTokens[from:]
		}
		return nil
	}

	var (
		sug            []string
		expectedTokens []string
	)

	if _, err := it.Parse(buffer); err != nil {
		var perr *parseError.Error
		if errors.As(err, &perr) {
			expectedTokens = perr.ExpectedTokens
		}
	} else {
		// Pass an 'at' symbol, not used in syntax and guaranteed to produce an error.
		_, err := it.Parse(fmt.Sprintf("%s @", buffer))
		if err == nil {
			panic("expected a parse error")
		}
		spew.Dump(err)
		var perr *parseError.Error
		if errors.As(err, &perr) {
			expectedTokens = perr.ExpectedTokens
		}
	}

	for _, text := range expectedTokens {
		switch text {
		// "INVALID":      0,
		// "$":            1,
		// "empty":        2,
		case "terminator":

		case "lineComment":
			sug = append(sug, "//")

		case "blockComment":
			sug = append(sug, "/*")

		case "bar":
			sug = append(sug, text)

		case "stringLit":
			sug = append(sug, `"`)

		case "{":
			sug = append(sug, text)

		case "}":
			sug = append(sug, text)

		case "[":
			sug = append(sug, text)

		case "]":
			sug = append(sug, text)

		case "char":
			if tokens := getLastNTokens(1); len(tokens) == 1 && tokens[0] == token.TokMap.Type("assign") {
				// Suggest unassigned keys on the current channel.
				for note := 'a'; note < 'z'; note++ {
					if _, ok := it.keymap[midiKey{it.curChannel, note}]; !ok {
						sug = append(sug, string(note))
					}
				}
				for note := 'A'; note < 'Z'; note++ {
					if _, ok := it.keymap[midiKey{it.curChannel, note}]; !ok {
						sug = append(sug, string(note))
					}
				}
			} else {
				// Suggest assigned keys on the current channel.
				for note := range it.keymap {
					if note.channel == it.curChannel {
						sug = append(sug, string(note.note))
					}
				}
			}

		case "rest":
			sug = append(sug, "-")

		case "sharp":
			sug = append(sug, "#")

		case "flat":
			sug = append(sug, "$")

		case "accent":
			sug = append(sug, "^")

		case "ghost":
			sug = append(sug, ")")

		case "uint":
			if tokens := getLastNTokens(2); len(tokens) == 2 {
				switch tokens[0] {
				case
					token.TokMap.Type("assign"),
					token.TokMap.Type("timesig"),
					token.TokMap.Type("control"):
					for i := 0; i <= 127; i++ {
						sug = append(sug, strconv.Itoa(i))
					}
				}
			}

			if tokens := getLastNTokens(1); len(tokens) == 1 {
				switch tokens[0] {
				case
					token.TokMap.Type("char"),
					token.TokMap.Type("tempo"),
					token.TokMap.Type("timesig"),
					token.TokMap.Type("channel"),
					token.TokMap.Type("velocity"),
					token.TokMap.Type("program"),
					token.TokMap.Type("control"):
					for i := 0; i <= 127; i++ {
						sug = append(sug, strconv.Itoa(i))
					}
				}
			}

		case "dot":
			sug = append(sug, ".")

		case "tuplet":
			// TODO: from 7th the midi precision gets lost
			sug = append(sug, "/3", "/5")

		case "letRing":
			sug = append(sug, "*")

		case "assign":
			sug = append(sug, text)

		case "tempo":
			sug = append(sug, text)

		case "timesig":
			sug = append(sug, text)

		case "channel":
			sug = append(sug, text)

		case "velocity":
			sug = append(sug, text)

		case "program":
			sug = append(sug, text)

		case "control":
			sug = append(sug, text)

		case "play":
			for name := range it.bars {
				sug = append(sug, fmt.Sprintf(`play "%s"`, name))
			}

		case "start":
			sug = append(sug, text)

		case "stop":
			sug = append(sug, text)
		}
	}

	return sug
}

// Tempo returns the current tempo.
// func (it *Interpreter) Tempo() uint16 {
// 	return it.curTempo
// }

// Parse a single input line into an AST node.
// func (it *Interpreter) Parse(input string) (interface{}, error) {
// 	if len(strings.TrimSpace(input)) == 0 {
// 		return nil, nil
// 	}

// 	it.parser.Reset()

// 	return it.parser.Parse(lexer.NewLexer([]byte(input)))
// }

// NoteOn creates a real time note on event on zero tick with an optional preceding NoteOff if the note was ringing.
// All notes are left ringing.
// func (it *Interpreter) NoteOn(note rune) ([]midi.Message, error) {
// 	key, ok := it.getKey(note)
// 	if !ok {
// 		return nil, fmt.Errorf("note '%c' undefined", note)
// 	}

// 	velocity := it.curVelocity
// 	ms := make([]midi.Message, 0, 2)

// 	if it.isRinging(note) {
// 		it.setRingingOff(note)
// 		ms = append(ms, midi.NoteOff(it.curChannel, key))
// 	}

// 	ms = append(ms, midi.NoteOn(it.curChannel, key, velocity))

// 	it.setRingingOn(note)

// 	return ms, nil
// }

func getMeter(event *sequencer.Event) (num uint8, denom uint8, ok bool) {
	ok = event.Message.GetMetaMeter(&num, &denom)
	return
}

// parseNoteList parses a note list into messages with relative ticks.
func (it *Interpreter) parseNoteList(noteList ast.NoteList) (sequencer.Events, error) {
	var (
		events sequencer.Events
		tick   smf.MetricTicks
	)

	for _, note := range noteList {
		length := note.Ticks()

		if note.IsPause() {
			tick += length
			continue
		}

		key, ok := it.getKey(it.curChannel, note.Name)
		if !ok {
			return nil, fmt.Errorf("note '%c' undefined", note.Name)
		}

		if note.IsSharp() {
			if key == constants.MaxValue {
				return nil, fmt.Errorf("sharp note '%s' out of MIDI range", note)
			}
			key++
		} else if note.IsFlat() {
			if key == constants.MinValue {
				return nil, fmt.Errorf("flat note '%s' out of MIDI range", note)
			}
			key--
		}

		velocity := it.curVelocity
		for i := uint(0); i < note.NumAccents(); i++ {
			if velocity > constants.MaxValue {
				velocity = constants.MaxValue
				break
			}
			// TODO: find the optimal value
			velocity += 10
		}

		for i := uint(0); i < note.NumGhosts(); i++ {
			// TODO: find the optimal value
			if velocity <= 10 {
				velocity = 1
				break
			}
			velocity -= 10
		}

		// TODO
		// if it.isRinging(note.Name) {
		// 	it.setRingingOff(note.Name)
		// }

		var pos uint8
		if tick > 0 {
			pos = uint8(tick.Ticks32th())
		}

		events = append(events, &sequencer.Event{
			TrackNo:  int(it.curChannel),
			Pos:      pos,
			Duration: uint8(length.Ticks32th()),
			Message:  smf.Message(midi.NoteOn(it.curChannel, key, velocity)),
		})

		// if note.IsLetRing() {
		// 	it.setRingingOn(note.Name)
		// }

		tick += length
	}

	sort.Sort(events)

	return events, nil
}

func (it *Interpreter) getKey(channel uint8, note rune) (uint8, bool) {
	key, ok := it.keymap[midiKey{channel, note}]
	return key, ok
}

func (it *Interpreter) assign(channel uint8, note rune, key uint8) error {
	if existingKey, ok := it.getKey(channel, note); ok {
		return fmt.Errorf("note '%c' already assigned to key '%d' on channel '%d'", note, existingKey, channel)
	}
	it.keymap[midiKey{channel, note}] = key
	return nil
}

// func (it *Interpreter) isRinging(note rune) bool {
// 	_, ok := it.ringing[midiKey{it.curChannel, note}]
// 	return ok
// }

// func (it *Interpreter) setRingingOn(note rune) {
// 	it.ringing[midiKey{it.curChannel, note}] = struct{}{}
// }

// func (it *Interpreter) setRingingOff(note rune) {
// 	delete(it.ringing, midiKey{it.curChannel, note})
// }

// New creates an interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser: parser.NewParser(),
		keymap: map[midiKey]uint8{},
		// ringing:     map[midiKey]struct{}{},
		bars:        map[string]sequencer.Events{},
		curVelocity: constants.DefaultVelocity,
	}
}
