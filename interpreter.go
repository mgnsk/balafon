package balafon

import (
	"fmt"
	"math"

	"github.com/mgnsk/balafon/internal/ast"
	"github.com/mgnsk/balafon/internal/constants"
	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/parser"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
	"golang.org/x/exp/slices"
)

// Interpreter evaluates text input and emits MIDI events.
type Interpreter struct {
	parser    *parser.Parser
	barBuffer []*Bar

	velocity int
	channel  uint8

	pos     uint32
	timesig [2]uint8

	keymap *keyMap
	bars   map[string]*Bar
}

// EvalString evaluates the string input.
func (it *Interpreter) EvalString(input string) error {
	return it.Eval([]byte(input))
}

// Eval evaluates the input.
func (it *Interpreter) Eval(input []byte) error {
	scanner := lexer.NewLexer(input)

	res, err := it.parser.Parse(scanner)
	if err != nil {
		return err
	}

	declList, ok := res.(ast.NodeList)
	if !ok {
		return fmt.Errorf("invalid input, expected ast.NodeList")
	}

	bars, err := it.parse(declList)
	if err != nil {
		return err
	}

	it.barBuffer = append(it.barBuffer, bars...)

	return nil
}

// Flush the parsed bar queue.
func (it *Interpreter) Flush() []*Bar {
	var (
		timesig      [2]uint8
		buf          []Event
		playableBars = make([]*Bar, 0, len(it.barBuffer))
	)

	// Defer virtual bars and concatenate them forward.
	for _, bar := range it.barBuffer {
		timesig = bar.TimeSig

		if bar.IsZeroDuration() {
			buf = append(buf, bar.Events...)
			continue
		}

		barEvs := make([]Event, 0, len(buf)+len(bar.Events))
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

	it.barBuffer = it.barBuffer[:0]

	for _, bar := range playableBars {
		slices.SortStableFunc(bar.Events, func(a, b Event) bool {
			return a.Pos < b.Pos
		})
	}

	return playableBars
}

func (it *Interpreter) beginBar() *Interpreter {
	return &Interpreter{
		velocity: it.velocity,
		channel:  it.channel,

		pos:     it.pos,
		timesig: it.timesig,

		keymap: it.keymap,
		bars:   it.bars,
	}
}

func (it *Interpreter) parse(declList ast.NodeList) ([]*Bar, error) {
	var bars []*Bar

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.CmdAssign:
			if !it.keymap.Set(it.channel, decl.Note, decl.Key) {
				old, _ := it.keymap.Get(it.channel, decl.Note)
				return nil, fmt.Errorf("note '%c' already assigned to key '%d' on channel '%d'", decl.Note, old, it.channel)
			}

		case ast.Bar:
			if _, ok := it.bars[decl.Name]; ok {
				return nil, fmt.Errorf("bar '%s' already defined", decl.Name)
			}
			barParser := it.beginBar()
			newBar, err := barParser.parseBar(decl.DeclList)
			if err != nil {
				return nil, err
			}
			if newBar == nil {
				return nil, fmt.Errorf("invalid empty bar '%s'", decl.Name)
			}
			it.bars[decl.Name] = newBar

		case ast.CmdPlay:
			savedBar, ok := it.bars[decl.BarName]
			if !ok {
				return nil, fmt.Errorf("unknown bar '%s'", decl.BarName)
			}
			bars = append(bars, savedBar)

		default:
			bar, err := it.parseBar(ast.NodeList{decl})
			if err != nil {
				return nil, err
			}
			if bar != nil {
				bars = append(bars, bar)
			}
		}
	}

	return bars, nil
}

func (it *Interpreter) parseBar(declList ast.NodeList) (*Bar, error) {
	bar := &Bar{
		TimeSig: it.timesig,
	}

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.CmdTempo:
			bar.Events = append(bar.Events, Event{
				Message: smf.MetaTempo(decl.Value()),
			})

		case ast.CmdTimeSig:
			it.timesig = decl
			bar.TimeSig = decl

		case ast.CmdVelocity:
			it.velocity = decl.Value()

		case ast.CmdChannel:
			it.channel = decl.Value()

		case ast.CmdProgram:
			bar.Events = append(bar.Events, Event{
				Message: smf.Message(midi.ProgramChange(it.channel, decl.Value())),
			})

		case ast.CmdControl:
			bar.Events = append(bar.Events, Event{
				Message: smf.Message(midi.ControlChange(it.channel, decl.Control, decl.Parameter)),
			})

		case ast.CmdStart:
			bar.Events = append(bar.Events, Event{
				Message: smf.Message(midi.Start()),
			})

		case ast.CmdStop:
			bar.Events = append(bar.Events, Event{
				Message: smf.Message(midi.Stop()),
			})

		case ast.NoteList:
			if err := it.parseNoteList(bar, decl); err != nil {
				return nil, err
			}

		default:
			panic(fmt.Sprintf("parse: invalid node %T", decl))
		}
	}

	if it.pos == 0 && len(bar.Events) == 0 {
		// Bar that consists of only timesig, velocity or channel commands and no events.
		return nil, nil
	}

	return bar, nil
}

// parseNoteList parses a note list into messages with relative ticks.
func (it *Interpreter) parseNoteList(bar *Bar, noteList ast.NoteList) error {
	it.pos = 0

	for _, note := range noteList {
		noteLen := note.Len()

		if note.IsPause() {
			it.pos += noteLen
			continue
		}

		key, ok := it.keymap.Get(it.channel, note.Name)
		if !ok {
			return fmt.Errorf("note '%c' undefined", note.Name)
		}

		key += note.NumSharp()
		key -= note.NumFlat()
		if key < 0 || key > constants.MaxValue {
			return fmt.Errorf("note key must be in range [%d, %d], got: %d", 0, constants.MaxValue, key)
		}

		v := it.velocity
		v += note.NumAccent() * 5
		v += note.NumMarcato() * 10
		v -= note.NumGhosts() * 5
		if v < 0 {
			v = 0
		} else if v > constants.MaxValue {
			v = math.MaxUint8
		}

		actualNoteLen := noteLen
		if n := uint32(note.NumStaccato()); n > 0 {
			actualNoteLen = actualNoteLen / (2 * n)
		}

		bar.Events = append(bar.Events, Event{
			Pos:      it.pos,
			Duration: actualNoteLen,
			Message:  smf.Message(midi.NoteOn(it.channel, uint8(key), uint8(v))),
		})

		if !note.IsLetRing() {
			bar.Events = append(bar.Events, Event{
				Pos:      it.pos + actualNoteLen,
				Duration: 0,
				Message:  smf.Message(midi.NoteOff(it.channel, uint8(key))),
			})
		}

		it.pos += noteLen
	}

	if it.pos > bar.Cap() {
		return fmt.Errorf("bar too long, timesig is %d/%d", it.timesig[0], it.timesig[1])
	}

	return nil
}

// New creates a balafon interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser:   parser.NewParser(),
		velocity: constants.DefaultVelocity,
		channel:  0,
		pos:      0,
		timesig:  [2]uint8{4, 4},
		keymap:   newKeyMap(),
		bars:     map[string]*Bar{},
	}
}
