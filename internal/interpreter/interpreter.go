package interpreter

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	"gitlab.com/gomidi/midi/v2"
)

// Message is a tempo or a MIDI message.
type Message struct {
	Msg   midi.Message
	Tick  uint64
	Tempo uint16
}

// Interpreter evaluates messages from raw line input.
type Interpreter struct {
	parser       *parser.Parser
	notes        map[rune]uint8
	ringing      map[uint16]struct{}
	bars         map[string][]ast.NoteList
	barBuffer    []ast.NoteList
	curBar       string
	curTick      uint64
	curBarLength uint64
	curChannel   uint8
	curVelocity  uint8
}

// Suggest returns suggestions for the next input.
// It is not safe to call Suggest concurrently
// with Eval.
func (it *Interpreter) Suggest() []string {
	var sug []string

	// Suggest assigned notes at any time.
	for note := range it.notes {
		sug = append(sug, string(note))
	}

	if it.curBar != "" {
		// Suggest ending the current bar if we're in the middle of a bar.
		sug = append(sug, "end")
	} else {
		// Suggest commands.
		sug = append(sug, "assign", "tempo", "channel", "velocity", "program", "control", "bar")
		// Suggest playing a bar.
		for name := range it.bars {
			sug = append(sug, "play "+name)
		}
	}

	return sug
}

// Eval evaluates a single input line.
// If both return values are nil, more input is needed.
func (it *Interpreter) Eval(input string) ([]Message, error) {
	if len(strings.TrimSpace(input)) == 0 {
		return nil, nil
	}

	it.parser.Reset()

	res, err := it.parser.Parse(lexer.NewLexer([]byte(input + "\n")))
	if err != nil {
		return nil, err
	}

	if res == nil {
		// Skip comments.
		return nil, nil
	}

	return it.evalResult(res)
}

func (it *Interpreter) errBarNotEnded(want string) error {
	return fmt.Errorf("cannot %s: bar '%s' is not ended", want, it.curBar)
}

func (it *Interpreter) evalResult(res interface{}) ([]Message, error) {
	switch r := res.(type) {
	case ast.NoteList:
		if it.curBar != "" {
			if it.curBarLength > 0 {
				barLength := uint64(0)
				for _, note := range r {
					barLength += note.Length()
				}
				if barLength != it.curBarLength {
					return nil, fmt.Errorf("invalid bar length")
				}
			}
			it.barBuffer = append(it.barBuffer, r)
			return nil, nil
		}
		return it.parseBar(r)

	case ast.CmdAssign:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("assign note")
		}
		it.notes[r.Note] = r.Key
		return nil, nil

	case ast.CmdTempo:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("change tempo")
		}
		return []Message{{
			Tempo: uint16(r),
		}}, nil

	case ast.CmdTimeSig:
		if it.curBar == "" {
			return nil, fmt.Errorf("timesig can only be set inside a bar")
		}
		it.curBarLength = uint64(r.Beats) * (4 * constants.TicksPerQuarter / uint64(r.Value))
		return nil, nil

	case ast.CmdChannel:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("change channel")
		}
		it.curChannel = uint8(r)
		return nil, nil

	case ast.CmdVelocity:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("change velocity")
		}
		it.curVelocity = uint8(r)
		return nil, nil

	case ast.CmdProgram:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("change program")
		}
		return []Message{{
			Msg: midi.NewMessage(midi.Channel(it.curChannel).ProgramChange(uint8(r))),
		}}, nil

	case ast.CmdControl:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("change control")
		}
		return []Message{{
			Msg: midi.NewMessage(midi.Channel(it.curChannel).ControlChange(r.Control, r.Value)),
		}}, nil

	case ast.CmdBar:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("begin bar")
		}
		bar := string(r)
		if _, ok := it.bars[bar]; ok {
			return nil, fmt.Errorf("bar '%s' already defined", bar)
		}
		it.curBar = bar
		return nil, nil

	case ast.CmdEnd:
		if it.curBar == "" {
			return nil, errors.New("cannot end bar: no active bar")
		}
		it.curBarLength = 0
		it.bars[it.curBar] = it.barBuffer
		it.curBar = ""
		it.barBuffer = nil
		return nil, nil

	case ast.CmdPlay:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("play bar")
		}
		bar := string(r)
		noteList, ok := it.bars[bar]
		if !ok {
			return nil, fmt.Errorf("bar '%s' does not exist", bar)
		}
		return it.parseBar(noteList...)

	case ast.CmdStart:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("start")
		}
		return []Message{{
			Msg: midi.NewMessage(midi.Start()),
		}}, nil

	case ast.CmdStop:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("stop")
		}
		return []Message{{
			Msg: midi.NewMessage(midi.Stop()),
		}}, nil

	default:
		panic(fmt.Sprintf("invalid expression: %#v", r))
	}
}

func (it *Interpreter) parseBar(tracks ...ast.NoteList) ([]Message, error) {
	var (
		messages     []Message
		furthestTick uint64
	)

	for _, track := range tracks {
		var tick uint64

		for _, note := range track {
			length := note.Length()

			if note.Name == '-' {
				tick += length
				continue
			}

			key, ok := it.notes[note.Name]
			if !ok {
				return nil, fmt.Errorf("key '%c' undefined", note.Name)
			}

			if note.IsSharp() {
				if key == 127 {
					return nil, fmt.Errorf("sharp note '%s' out of MIDI range", note)
				}
				key++
			} else if note.IsFlat() {
				if key == 0 {
					return nil, fmt.Errorf("flat note '%s' out of MIDI range", note)
				}
				key--
			}

			velocity := it.curVelocity
			if note.IsAccent() {
				velocity *= 2
				if velocity > 127 {
					velocity = 127
				}
			} else if note.IsGhost() {
				velocity /= 2
			}

			r := uint16(it.curChannel)<<8 | uint16(key)
			if _, ok := it.ringing[r]; ok {
				delete(it.ringing, r)
				messages = append(messages,
					Message{
						Tick: it.curTick + tick,
						Msg:  midi.NewMessage(midi.Channel(it.curChannel).NoteOff(key)),
					},
					Message{
						Tick: it.curTick + tick,
						Msg:  midi.NewMessage(midi.Channel(it.curChannel).NoteOn(key, velocity)),
					},
				)
			} else {
				messages = append(messages, Message{
					Tick: it.curTick + tick,
					Msg:  midi.NewMessage(midi.Channel(it.curChannel).NoteOn(key, velocity)),
				})
			}

			if note.IsLetRing() {
				it.ringing[r] = struct{}{}
			} else {
				messages = append(messages, Message{
					Tick: it.curTick + tick + length,
					Msg:  midi.NewMessage(midi.Channel(it.curChannel).NoteOff(key)),
				})
			}

			tick += length

			if tick > furthestTick {
				furthestTick = tick
			}
		}

		// check if tick  equals bar length

	}

	it.curTick += furthestTick

	// Sort the messages so that every note is off before on.
	sort.Sort(byMessageTypeOrKey(messages))

	return messages, nil
}

// New creates an interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser:      parser.NewParser(),
		notes:       map[rune]uint8{},
		ringing:     map[uint16]struct{}{},
		bars:        map[string][]ast.NoteList{},
		curVelocity: 127,
	}
}

// LoadAll loads all messages from r.
func LoadAll(r io.Reader) ([][]Message, error) {
	it := New()
	s := bufio.NewScanner(r)

	var format strings.Builder

	var messages [][]Message

	line := 0
	for s.Scan() {
		line++
		ms, err := it.Eval(s.Text())
		if err != nil {
			format.WriteString(lineError{line, err}.Error())
			format.WriteString("\n")
		} else if ms != nil {
			messages = append(messages, ms)
		}
	}

	if err := s.Err(); err != nil {
		return nil, err
	}

	if format.Len() > 0 {
		return nil, errors.New(format.String())
	}

	return messages, nil
}

type byMessageTypeOrKey []Message

func (s byMessageTypeOrKey) Len() int      { return len(s) }
func (s byMessageTypeOrKey) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byMessageTypeOrKey) Less(i, j int) bool {
	if s[i].Tick == s[j].Tick {
		a := s[i].Msg
		b := s[j].Msg

		if a.IsNoteEnd() {
			if b.IsNoteEnd() {
				// When both are NoteOff, sort by key.
				return a.Key() < b.Key()
			}
			// NoteOff before any other messages on the same tick.
			return true
		}
	}
	return s[i].Tick < s[j].Tick
}
