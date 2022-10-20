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
	"golang.org/x/exp/slices"
)

// Message is a MIDI message.
// type Message struct {
// 	midi.Message
// 	Tick       uint32
// 	NoteLength uint32
// }

type midiKey struct {
	channel uint8
	note    rune
}

// Interpreter evaluates messages from raw line input.
type Interpreter struct {
	parser      *parser.Parser
	keymap      map[midiKey]uint8
	ringing     map[midiKey]struct{}
	bars        map[string]sequencer.Events
	curChannel  uint8
	curVelocity uint8
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

// TODO: parseOption: fillBarSilence - fills bars with

func (it *Interpreter) Eval(input string) (sequencer.Events, error) {
	res, err := it.parser.Parse(lexer.NewLexer([]byte(input)))
	if err != nil {
		return nil, err
	}

	declList, ok := res.(ast.DeclList)
	if !ok {
		return nil, fmt.Errorf("invalid input")
	}

	return it.ParseAST(declList)
}

func (it *Interpreter) ParseAST(declList ast.DeclList) (sequencer.Events, error) {
	var events sequencer.Events

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.Bar:
			if _, ok := it.bars[decl.Name]; ok {
				return nil, fmt.Errorf("bar '%s' already defined", decl.Name)
			}

			barEvents, err := it.ParseAST(decl.DeclList)
			if err != nil {
				return nil, err
			}

			it.bars[decl.Name] = barEvents

		case ast.CmdAssign:
			if err := it.assign(it.curChannel, decl.Note, decl.Key); err != nil {
				return nil, err
			}

		case ast.CmdTempo:
			events = append(events, &sequencer.Event{
				Message: smf.MetaTempo(float64(decl)),
			})

		case ast.CmdTimeSig:
			// it.curBarLength = smf.MetricTicks(decl.Num) * (constants.TicksPerWhole / smf.MetricTicks(decl.Denom))
			events = append(events, &sequencer.Event{
				Message: smf.MetaMeter(decl.Num, decl.Denom),
			})

		case ast.CmdVelocity:
			it.curVelocity = uint8(decl)

		case ast.CmdChannel:
			it.curChannel = uint8(decl)

		case ast.CmdProgram:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.ProgramChange(it.curChannel, uint8(decl))),
			})

		case ast.NoteList:
			noteEvents, err := it.parseNoteList(decl)
			if err != nil {
				return nil, err
			}

			events = append(events, noteEvents...)

		case ast.CmdControl:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.ControlChange(it.curChannel, decl.Control, decl.Parameter)),
			})

		case ast.CmdPlay:
			// TODO
			// bar, ok := it.bars[string(decl)]
			// if !ok {
			// 	return bar, nil
			// }
			// return nil, fmt.Errorf("unknown bar in command %s", decl.String())

		case ast.CmdStart:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.Start()),
			})

		case ast.CmdStop:
			events = append(events, &sequencer.Event{
				Message: smf.Message(midi.Stop()),
			})

		default:
			panic(fmt.Sprintf("invalid token %T", decl))
		}
	}

	return events, nil
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

// Eval evaluates a single input line.
// If both return values are nil, more input is needed.
// func (it *Interpreter) Eval(input string) (sequencer.Events, error) {
// 	res, err := it.Parse(input)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if res == nil {
// 		// Skip comments.
// 		return nil, nil
// 	}

// 	switch r := res.(type) {
// 	case ast.NoteList:
// 		return it.addNotes(r)

// 	case ast.CmdAssign:
// 		if err := it.assign(r.Note, r.Key); err != nil {
// 			return nil, err
// 		}
// 		return nil, nil

// 	default:
// 		panic(fmt.Sprintf("invalid expression: %#v", r))
// 	}
// }

func getTimeSig(event *sequencer.Event) [2]uint8 {
	var num, denom, c, ds uint8
	if event.Message.GetMetaTimeSig(&num, &denom, &c, &ds) {
		return [2]uint8{num, denom}
	}
	return [2]uint8{}
}

// EvalAll evaluates all messages from r.
// func (it *Interpreter) EvalAll(r io.Reader) (*sequencer.Song, error) {
// 	s := bufio.NewScanner(r)

// 	var errs strings.Builder

// 	var bar sequencer.Bar
// 	song := sequencer.New()

// 	line := 0
// 	for s.Scan() {
// 		line++

// 		if events, err := it.Eval(s.Text()); err != nil {
// 			errs.WriteString(fmt.Sprintf("[%d]: %s\n", line, err))
// 		} else if events != nil {
// 			hasNotes := false

// 			for _, event := range events {
// 				switch event.Message.Type() {
// 				case smf.MetaTimeSigMsg:
// 					bar.TimeSig = getTimeSig(event)
// 				case midi.NoteOnMsg:
// 					hasNotes = true
// 				}
// 				bar.Events = append(bar.Events, event)
// 			}

// 			if hasNotes {
// 				song.AddBar(bar)
// 				bar = sequencer.Bar{}
// 			}
// 		}
// 	}

// 	if err := s.Err(); err != nil {
// 		return nil, err
// 	}

// 	if errs.Len() > 0 {
// 		return nil, errors.New(errs.String())
// 	}

// 	if len(bar.Events) > 0 {
// 		song.AddBar(bar)
// 	}

// 	song.SetBarAbsTicks()

// 	return song, nil
// }

// func (it *Interpreter) errBarNotEnded(want string) error {
// 	return fmt.Errorf("cannot %s: bar '%s' is not ended", want, it.curBar)
// }

// func (it *Interpreter) play(events sequencer.Events) {
// 	// Offset relative messages to absolute tick.
// 	ms := make([]Message, len(events))
// 	for i, m := range events {
// 		m.Tick = it.curTick + m.Tick
// 		ms[i] = m
// 	}
// 	lastMsg := events[len(events)-1]
// 	if lastMsg.Is(midi.NoteOnMsg) {
// 		it.curTick += lastMsg.Tick + lastMsg.NoteLength
// 	} else {
// 		it.curTick += lastMsg.Tick
// 	}
// 	return ms
// }

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

		if it.isRinging(note.Name) {
			// TODO: midi dur 0?
			it.setRingingOff(note.Name)
		}

		// TrackNo  int
		// Pos      uint8       // in 32th
		// Duration uint8       // in 32th for noteOn messages, it is the length of the note, for all other messages, it is 0
		// Message  smf.Message // may only be channel messages or sysex messages. no noteon velocity 0, or noteoff messages, this is expressed via Duration
		// absTicks int64       // just for smf import

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

		if note.IsLetRing() {
			it.setRingingOn(note.Name)
		}

		tick += length
	}

	sort.Sort(events)
	// sort.Sort(byMessageTypeOrKey(events))

	return events, nil
}

func (it *Interpreter) getKey(channel uint8, note rune) (uint8, bool) {
	key, ok := it.keymap[midiKey{channel, note}]
	return key, ok
}

func (it *Interpreter) assign(channel uint8, note rune, key uint8) error {
	if key, ok := it.getKey(channel, note); ok {
		return fmt.Errorf("note '%c' already assigned to key '%d' on channel '%d'", note, key, it.curChannel)
	}
	it.keymap[midiKey{channel, note}] = key
	return nil
}

func (it *Interpreter) isRinging(note rune) bool {
	_, ok := it.ringing[midiKey{it.curChannel, note}]
	return ok
}

func (it *Interpreter) setRingingOn(note rune) {
	it.ringing[midiKey{it.curChannel, note}] = struct{}{}
}

func (it *Interpreter) setRingingOff(note rune) {
	delete(it.ringing, midiKey{it.curChannel, note})
}

// func (it *Interpreter) addNotes(list ast.NoteList) (sequencer.Events, error) {
// 	if it.curBar != "" {
// 		if it.curBarLength > 0 {
// 			var barLength smf.MetricTicks
// 			for _, note := range list {
// 				barLength += note.Ticks()
// 			}
// 			if barLength != it.curBarLength {
// 				return nil, fmt.Errorf("invalid bar length")
// 			}
// 		}
// 		events, err := it.parseNoteList(list)
// 		if err != nil {
// 			return nil, err
// 		}
// 		it.barBuffer = append(it.barBuffer, events...)
// 		return nil, nil
// 	}

// 	events, err := it.parseNoteList(list)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return events, nil
// }

// case ast.CmdControl:
// 	event := &sequencer.Event{
// 		// Tick:    it.curTick,
// 		Message: smf.Message(midi.ControlChange(it.curChannel, r.Control, r.Parameter)),
// 	}
// 	if it.barName != "" {
// 		it.barBuffer = append(it.barBuffer, event)
// 		return nil, nil
// 	}
// 	return sequencer.Events{event}, nil
// New creates an interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser:      parser.NewParser(),
		keymap:      map[midiKey]uint8{},
		ringing:     map[midiKey]struct{}{},
		bars:        map[string]sequencer.Events{},
		curVelocity: constants.MaxValue,
		// curTempo:    constants.DefaultTempo,
	}
}

// type byMessageTypeOrKey []Message

// func (s byMessageTypeOrKey) Len() int      { return len(s) }
// func (s byMessageTypeOrKey) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
// func (s byMessageTypeOrKey) Less(i, j int) bool {
// 	if s[i].Tick == s[j].Tick {
// 		a := s[i].Message
// 		b := s[j].Message

// 		if a.Is(midi.NoteOffMsg) && !b.Is(midi.NoteOffMsg) {
// 			return true
// 		}

// 		return false
// 	}

// 	return s[i].Tick < s[j].Tick
// }

func find(events sequencer.Events, t midi.Type) int {
	return slices.IndexFunc(events, func(e *sequencer.Event) bool {
		return e.Message.Is(t)
	})
}
