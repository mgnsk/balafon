package interpreter

import (
	"fmt"

	"github.com/mgnsk/gong/ast"
	"github.com/mgnsk/gong/constants"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Parser parses AST into events.
type Parser struct {
	velocity uint8
	channel  uint8

	pos     uint32
	timesig [2]uint8

	keymap *KeyMap
	bars   map[string]*Bar
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
		bars:     map[string]*Bar{},
	}
}

// Parse AST.
func (p *Parser) Parse(declList ast.NodeList) ([]*Bar, error) {
	var bars []*Bar

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
			bars = append(bars, savedBar)

		default:
			bar, err := p.parseBar(ast.NodeList{decl})
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

func (p *Parser) parseBar(declList ast.NodeList) (*Bar, error) {
	bar := &Bar{
		TimeSig: p.timesig,
	}

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.CmdTempo:
			bar.PrependMetaMessage(smf.MetaTempo(float64(decl)))

		case ast.CmdTimeSig:
			p.timesig = [2]uint8{decl.Num, decl.Denom}
			bar.TimeSig = p.timesig

		case ast.CmdVelocity:
			p.velocity = uint8(decl)

		case ast.CmdChannel:
			p.channel = uint8(decl)

		case ast.CmdProgram:
			bar.PrependMetaMessage(midi.ProgramChange(p.channel, uint8(decl)))

		case ast.CmdControl:
			bar.PrependMetaMessage(midi.ControlChange(p.channel, decl.Control, decl.Parameter))

		case ast.CmdStart:
			bar.PrependMetaMessage(midi.Start())

		case ast.CmdStop:
			bar.PrependMetaMessage(midi.Stop())

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

	// bar.SortEvents()

	return bar, nil
}

// parseNoteList parses a note list into messages with relative ticks.
func (p *Parser) parseNoteList(bar *Bar, noteList ast.NoteList) error {
	p.pos = 0

	for _, note := range noteList {
		noteLen := note.Len()

		if note.IsPause() {
			p.pos += noteLen
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

		bar.Events = append(bar.Events, Event{
			Channel:  p.channel,
			Pos:      p.pos,
			Duration: noteLen,
			Message:  smf.Message(midi.NoteOn(p.channel, key, velocity)),
		})

		p.pos += noteLen
	}

	if p.pos > bar.Cap() {
		return fmt.Errorf("bar too long, timesig is %d/%d", p.timesig[0], p.timesig[1])
	}

	return nil
}
