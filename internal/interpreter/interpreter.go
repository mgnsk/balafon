package interpreter

import (
	"fmt"
	"sort"

	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
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

// Clone returns a copy of the interpreter.
func (it *Interpreter) Clone() *Interpreter {
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

func (it *Interpreter) Eval(input string) (*sequencer.Song, error) {
	res, err := it.parser.Parse(lexer.NewLexer([]byte(input)))
	if err != nil {

		return nil, err
	}

	declList, ok := res.(ast.DeclList)
	if !ok {
		return nil, fmt.Errorf("invalid input, expected ast.DeclList")
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

			events, err := it.Clone().parse(decl.DeclList)
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

// Suggest returns suggestions for the next input.
// It is not safe to call Suggest concurrently
// with Eval.
func (it *Interpreter) Suggest() []string {
	var sug []string

	// TODO
	// // Suggest assigned notes at any time.
	// for note := range it.keymap {
	// 	sug = append(sug, string(note.note))
	// }

	// if it.curBar != "" {
	// 	sug = append(sug, sugInsideBar...)
	// } else {
	// 	sug = append(sug, sugOutsideBar...)
	// 	// Suggest playing a bar.
	// 	for name := range it.bars {
	// 		sug = append(sug, fmt.Sprintf(`play "%s"`, name))
	// 	}
	// }

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
		if note.IsAccent() {
			velocity *= 2
			if velocity > constants.MaxValue {
				velocity = constants.MaxValue
			}
		} else if note.IsGhost() {
			velocity /= 2
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
