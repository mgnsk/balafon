package interpreter

import (
	"fmt"

	"github.com/mgnsk/gong/ast"
	"github.com/mgnsk/gong/constants"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/sequencer"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Parser parses AST into sequencer events.
type Parser struct {
	velocity uint8
	channel  uint8

	pos     uint8
	timesig [2]uint8

	keymap *KeyMap
	bars   map[string]*sequencer.Bar
}

func (p *Parser) beginBar() *Parser {
	return &Parser{
		velocity: p.velocity,
		channel:  p.channel,

		pos:     0,
		timesig: p.timesig,

		keymap: p.keymap,
		bars:   p.bars,
	}
}

// NewParser creates a parser.
func NewParser() *Parser {
	return &Parser{
		velocity: constants.DefaultVelocity,
		channel:  0,
		pos:      0,
		timesig:  [2]uint8{4, 4},
		keymap:   NewKeyMap(),
		bars:     map[string]*sequencer.Bar{},
	}
}

// Parse AST.
func (p *Parser) Parse(declList ast.NodeList) ([]sequencer.Bar, error) {
	var bars []sequencer.Bar

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.CmdAssign:
			if !p.keymap.Set(p.channel, decl.Note, decl.Key) {
				old, _ := p.keymap.Get(p.channel, decl.Note)
				return nil, fmt.Errorf("note '%c' already assigned to key '%d' on channel '%d'", decl.Note, old, p.channel)
			}

		case ast.Bar:
			if _, ok := p.bars[decl.Name]; ok {
				return nil, fmt.Errorf("bar '%s' already defined", decl.Name)
			}

			barParser := p.beginBar()
			newBar, err := barParser.parseBar(decl.DeclList)
			if err != nil {
				return nil, err
			}
			if newBar == nil {
				panic("TODO: nil bar")
			}
			p.bars[decl.Name] = newBar

		case ast.CmdPlay:
			savedBar, ok := p.bars[string(decl)]
			if !ok {
				return nil, fmt.Errorf("unknown bar '%s'", string(decl))
			}
			bars = append(bars, *savedBar)

		default:
			bar, err := p.parseBar(ast.NodeList{decl})
			if err != nil {
				return nil, err
			}
			if bar != nil {
				bars = append(bars, *bar)
			}
		}
	}

	return bars, nil
}

func (p *Parser) parseBar(declList ast.NodeList) (*sequencer.Bar, error) {
	bar := &sequencer.Bar{
		TimeSig: p.timesig,
	}

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.CmdTempo:
			bar.Events = append(bar.Events, &sequencer.Event{
				TrackNo: 0,
				Message: smf.MetaTempo(float64(decl)),
			})

		case ast.CmdTimeSig:
			p.timesig = [2]uint8{decl.Num, decl.Denom}
			bar.TimeSig = [2]uint8{decl.Num, decl.Denom}

		case ast.CmdVelocity:
			p.velocity = uint8(decl)

		case ast.CmdChannel:
			p.channel = uint8(decl)

		case ast.CmdProgram:
			bar.Events = append(bar.Events, &sequencer.Event{
				TrackNo: int(p.channel),
				Message: smf.Message(midi.ProgramChange(p.channel, uint8(decl))),
			})

		case ast.CmdControl:
			bar.Events = append(bar.Events, &sequencer.Event{
				TrackNo: int(p.channel),
				Message: smf.Message(midi.ControlChange(p.channel, decl.Control, decl.Parameter)),
			})

		case ast.CmdStart:
			bar.Events = append(bar.Events, &sequencer.Event{
				TrackNo: int(p.channel),
				Message: smf.Message(midi.Start()),
			})

		case ast.CmdStop:
			bar.Events = append(bar.Events, &sequencer.Event{
				TrackNo: int(p.channel),
				Message: smf.Message(midi.Stop()),
			})

		case ast.NoteList:
			if err := p.parseNoteList(bar, decl); err != nil {
				return nil, err
			}

		default:
			panic(fmt.Sprintf("parse: invalid node %T", decl))
		}
	}

	if p.pos == 0 && len(bar.Events) == 0 {
		return nil, nil
	}

	bar.SortEvents()

	return bar, nil
}

// parseNoteList parses a note list into messages with relative ticks.
func (p *Parser) parseNoteList(bar *sequencer.Bar, noteList ast.NoteList) error {
	p.pos = 0

	for _, note := range noteList {
		length32th := note.Len()

		if note.IsPause() {
			p.pos += length32th
			continue
		}

		key, ok := p.keymap.Get(p.channel, note.Name)
		if !ok {
			return fmt.Errorf("note '%c' undefined", note.Name)
		}

		if note.IsSharp() {
			if key == constants.MaxValue {
				return fmt.Errorf("sharp note '%c' out of MIDI range", note.Name)
			}
			key++
		} else if note.IsFlat() {
			if key == constants.MinValue {
				return fmt.Errorf("flat note '%c' out of MIDI range", note.Name)
			}
			key--
		}

		velocity := p.velocity
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

		bar.Events = append(bar.Events, &sequencer.Event{
			TrackNo:  int(p.channel),
			Pos:      p.pos,
			Duration: length32th,
			Message:  smf.Message(midi.NoteOn(p.channel, key, velocity)),
		})

		p.pos += length32th
	}

	if p.pos > bar.Len() {
		return fmt.Errorf("bar too long")
	}

	return nil
}
