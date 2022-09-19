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
	"gitlab.com/gomidi/midi/v2/smf"
)

// Message is a MIDI message.
type Message struct {
	midi.Message
	Tick       uint32
	NoteLength uint32
}

type midiKey struct {
	channel uint8
	note    rune
}

// Interpreter evaluates messages from raw line input.
type Interpreter struct {
	parser       *parser.Parser
	keymap       map[midiKey]uint8
	ringing      map[midiKey]struct{}
	bars         map[string][]Message
	barBuffer    []Message
	curBar       string
	curTick      uint32
	curBarLength uint32
	curTempo     uint16
	curChannel   uint8
	curVelocity  uint8
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

// Suggest returns suggestions for the next input.
// It is not safe to call Suggest concurrently
// with Eval.
func (it *Interpreter) Suggest() []string {
	var sug []string

	// Suggest assigned notes at any time.
	for note := range it.keymap {
		sug = append(sug, string(note.note))
	}

	if it.curBar != "" {
		sug = append(sug, sugInsideBar...)
	} else {
		sug = append(sug, sugOutsideBar...)
		// Suggest playing a bar.
		for name := range it.bars {
			sug = append(sug, fmt.Sprintf(`play "%s"`, name))
		}
	}

	return sug
}

// Tempo returns the current tempo.
func (it *Interpreter) Tempo() uint16 {
	return it.curTempo
}

// Parse a single input line into an AST node.
func (it *Interpreter) Parse(input string) (interface{}, error) {
	if len(strings.TrimSpace(input)) == 0 {
		return nil, nil
	}

	it.parser.Reset()

	return it.parser.Parse(lexer.NewLexer([]byte(input + "\n")))
}

// NoteOn creates a real time note on event on zero tick with an optional preceding NoteOff if the note was ringing.
// All notes are left ringing.
func (it *Interpreter) NoteOn(note rune) ([]Message, error) {
	key, ok := it.getKey(note)
	if !ok {
		return nil, fmt.Errorf("note '%c' undefined", note)
	}

	velocity := it.curVelocity
	messages := make([]Message, 0, 2)

	if it.isRinging(note) {
		it.setRingingOff(note)
		messages = append(messages, Message{
			Message: midi.NoteOff(it.curChannel, key),
		})
	}

	messages = append(messages, Message{
		Message: midi.NoteOn(it.curChannel, key, velocity),
	})

	it.setRingingOn(note)

	return messages, nil
}

// Eval evaluates a single input line.
// If both return values are nil, more input is needed.
func (it *Interpreter) Eval(input string) ([]Message, error) {
	res, err := it.Parse(input)
	if err != nil {
		return nil, err
	}

	if res == nil {
		// Skip comments.
		return nil, nil
	}

	return it.evalResult(res)
}

// EvalAll evaluates all messages from r.
func (it *Interpreter) EvalAll(r io.Reader) ([]Message, error) {
	s := bufio.NewScanner(r)

	var format strings.Builder

	var messages []Message

	line := 0
	for s.Scan() {
		line++
		ms, err := it.Eval(s.Text())
		if err != nil {
			format.WriteString(fmt.Sprintf("[%d]: %s\n", line, err))
		} else if ms != nil {
			messages = append(messages, ms...)
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

func (it *Interpreter) errBarNotEnded(want string) error {
	return fmt.Errorf("cannot %s: bar '%s' is not ended", want, it.curBar)
}

func (it *Interpreter) play(messages []Message) []Message {
	// Offset relative messages to absolute tick.
	ms := make([]Message, len(messages))
	for i, m := range messages {
		m.Tick = it.curTick + m.Tick
		ms[i] = m
	}
	lastMsg := messages[len(messages)-1]
	if lastMsg.Is(midi.NoteOnMsg) {
		it.curTick += lastMsg.Tick + lastMsg.NoteLength
	} else {
		it.curTick += lastMsg.Tick
	}
	return ms
}

func (it *Interpreter) evalResult(res interface{}) ([]Message, error) {
	switch r := res.(type) {
	case ast.NoteList:
		if it.curBar != "" {
			if it.curBarLength > 0 {
				barLength := uint32(0)
				for _, note := range r {
					barLength += note.Length()
				}
				if barLength != it.curBarLength {
					return nil, fmt.Errorf("invalid bar length")
				}
			}
			ms, err := it.parseNoteList(r)
			if err != nil {
				return nil, err
			}
			it.barBuffer = append(it.barBuffer, ms...)
			return nil, nil
		}

		messages, err := it.parseNoteList(r)
		if err != nil {
			return nil, err
		}

		return it.play(messages), nil

	case ast.CmdAssign:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("assign note")
		}
		if err := it.assign(r.Note, r.Key); err != nil {
			return nil, err
		}
		return nil, nil

	case ast.CmdTempo:
		msg := Message{
			Tick:    it.curTick,
			Message: midi.Message(smf.MetaTempo(float64(r))),
		}
		if it.curBar != "" {
			if containsNotes(it.barBuffer) {
				return nil, fmt.Errorf("tempo command must be at the beginning of bar")
			}
			it.barBuffer = append(it.barBuffer, msg)
			it.curTempo = uint16(r)
			return nil, nil
		}
		it.curTempo = uint16(r)
		return []Message{msg}, nil

	case ast.CmdTimeSig:
		if it.curBar == "" {
			return nil, fmt.Errorf("timesig can only be set inside a bar")
		}
		if containsNotes(it.barBuffer) {
			return nil, fmt.Errorf("timesig command must be at the beginning of bar")
		}
		it.curBarLength = uint32(r.Num) * (4 * constants.TicksPerQuarter / uint32(r.Denom))
		it.barBuffer = append(it.barBuffer, Message{
			Tick:    it.curTick,
			Message: midi.Message(smf.MetaMeter(r.Num, r.Denom)),
		})
		return nil, nil

	case ast.CmdChannel:
		it.curChannel = uint8(r)
		return nil, nil

	case ast.CmdVelocity:
		it.curVelocity = uint8(r)
		return nil, nil

	case ast.CmdProgram:
		msg := Message{
			Tick:    it.curTick,
			Message: midi.ProgramChange(it.curChannel, uint8(r)),
		}
		if it.curBar != "" {
			it.barBuffer = append(it.barBuffer, msg)
			return nil, nil
		}
		return []Message{msg}, nil

	case ast.CmdControl:
		msg := Message{
			Tick:    it.curTick,
			Message: midi.ControlChange(it.curChannel, r.Control, r.Parameter),
		}
		if it.curBar != "" {
			it.barBuffer = append(it.barBuffer, msg)
			return nil, nil
		}
		return []Message{msg}, nil

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
		sort.Sort(byMessageTypeOrKey(it.barBuffer))
		it.curBarLength = 0
		it.bars[it.curBar] = it.barBuffer
		it.curBar = ""
		it.barBuffer = nil
		return nil, nil

	case ast.CmdPlay:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("play bar")
		}
		messages, ok := it.bars[string(r)]
		if !ok {
			return nil, fmt.Errorf("bar '%s' does not exist", string(r))
		}
		if len(messages) == 0 {
			return nil, fmt.Errorf("invalid bar '%s'", string(r))
		}
		return it.play(messages), nil

	case ast.CmdStart:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("start")
		}
		return []Message{{
			Tick:    it.curTick,
			Message: midi.Start(),
		}}, nil

	case ast.CmdStop:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("stop")
		}
		return []Message{{
			Tick:    it.curTick,
			Message: midi.Stop(),
		}}, nil

	default:
		panic(fmt.Sprintf("invalid expression: %#v", r))
	}
}

// parseNoteList parses a note list into messages with relative ticks.
func (it *Interpreter) parseNoteList(noteList ast.NoteList) ([]Message, error) {
	var (
		messages []Message
		tick     uint32
	)

	for _, note := range noteList {
		length := note.Length()

		if note.IsPause() {
			tick += length
			continue
		}

		key, ok := it.getKey(note.Name)
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
			it.setRingingOff(note.Name)
			messages = append(messages,
				Message{
					Tick:    tick,
					Message: midi.NoteOff(it.curChannel, key),
				},
			)
		}

		messages = append(messages, Message{
			Tick:       tick,
			NoteLength: length,
			Message:    midi.NoteOn(it.curChannel, key, velocity),
		})

		if note.IsLetRing() {
			it.setRingingOn(note.Name)
		} else {
			messages = append(messages, Message{
				Tick:    tick + length,
				Message: midi.NoteOff(it.curChannel, key),
			})
		}

		tick += length
	}

	sort.Sort(byMessageTypeOrKey(messages))

	return messages, nil
}

func (it *Interpreter) getKey(note rune) (uint8, bool) {
	key, ok := it.keymap[midiKey{it.curChannel, note}]
	return key, ok
}

func (it *Interpreter) assign(note rune, key uint8) error {
	if k, ok := it.getKey(note); ok {
		return fmt.Errorf("note '%c' already assigned to '%d' on channel '%d'", note, k, it.curChannel)
	}
	it.keymap[midiKey{it.curChannel, note}] = key
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

// New creates an interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser:      parser.NewParser(),
		keymap:      map[midiKey]uint8{},
		ringing:     map[midiKey]struct{}{},
		bars:        map[string][]Message{},
		curVelocity: constants.MaxValue,
		curTempo:    constants.DefaultTempo,
	}
}

type byMessageTypeOrKey []Message

func (s byMessageTypeOrKey) Len() int      { return len(s) }
func (s byMessageTypeOrKey) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byMessageTypeOrKey) Less(i, j int) bool {
	if s[i].Tick == s[j].Tick {
		a := s[i].Message
		b := s[j].Message

		if a.Is(midi.NoteOffMsg) && !b.Is(midi.NoteOffMsg) {
			return true
		}

		return false
	}

	return s[i].Tick < s[j].Tick
}

func containsNotes(messages []Message) bool {
	for _, msg := range messages {
		if msg.Is(midi.NoteOnMsg) {
			return true
		}
	}
	return false
}
