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

// Message is a MIDI message.
type Message struct {
	Msg        midi.Message
	Tick       uint32
	NoteLength uint32
}

// Interpreter evaluates messages from raw line input.
type Interpreter struct {
	parser        *parser.Parser
	channelKeymap map[uint8]map[rune]uint8
	ringing       map[uint16]struct{}
	bars          map[string][]Message
	barBuffer     []Message
	curBar        string
	curTick       uint32
	curBarLength  uint32
	curChannel    uint8
	curVelocity   uint8
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
	for note := range it.channelKeymap {
		sug = append(sug, string(note))
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
			format.WriteString(lineError{line, err}.Error())
			format.WriteString("\n")
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
	if lastMsg.Msg.IsNoteStart() {
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
		if key, ok := it.channelKeymap[it.curChannel][r.Note]; ok {
			return nil, fmt.Errorf("note '%c' already assigned to '%d'", r.Note, key)
		}
		it.channelKeymap[it.curChannel][r.Note] = r.Key
		return nil, nil

	case ast.CmdTempo:
		msg := Message{
			Tick: it.curTick,
			Msg:  midi.NewMessage(midi.MetaTempo(float64(r))),
		}
		if it.curBar != "" {
			if containsNotes(it.barBuffer) {
				return nil, fmt.Errorf("tempo command must be at the beginning of bar")
			}
			it.barBuffer = append(it.barBuffer, msg)
			return nil, nil
		}
		return []Message{msg}, nil

	case ast.CmdTimeSig:
		if it.curBar == "" {
			return nil, fmt.Errorf("timesig can only be set inside a bar")
		}
		if containsNotes(it.barBuffer) {
			return nil, fmt.Errorf("timesig command must be at the beginning of bar")
		}
		it.curBarLength = uint32(r.Beats) * (4 * constants.TicksPerQuarter / uint32(r.Value))
		it.barBuffer = append(it.barBuffer, Message{
			Tick: it.curTick,
			Msg:  midi.NewMessage(midi.MetaMeter(r.Beats, r.Value)),
		})
		return nil, nil

	case ast.CmdChannel:
		it.curChannel = uint8(r)
		if _, ok := it.channelKeymap[it.curChannel]; !ok {
			it.channelKeymap[it.curChannel] = map[rune]uint8{}
		}
		return nil, nil

	case ast.CmdVelocity:
		it.curVelocity = uint8(r)
		return nil, nil

	case ast.CmdProgram:
		msg := Message{
			Tick: it.curTick,
			Msg:  midi.NewMessage(midi.Channel(it.curChannel).ProgramChange(uint8(r))),
		}
		if it.curBar != "" {
			it.barBuffer = append(it.barBuffer, msg)
			return nil, nil
		}
		return []Message{msg}, nil

	case ast.CmdControl:
		msg := Message{
			Tick: it.curTick,
			Msg:  midi.NewMessage(midi.Channel(it.curChannel).ControlChange(r.Control, r.Parameter)),
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
			Tick: it.curTick,
			Msg:  midi.NewMessage(midi.Start()),
		}}, nil

	case ast.CmdStop:
		if it.curBar != "" {
			return nil, it.errBarNotEnded("stop")
		}
		return []Message{{
			Tick: it.curTick,
			Msg:  midi.NewMessage(midi.Stop()),
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

		key, ok := it.channelKeymap[it.curChannel][note.Name]
		if !ok {
			return nil, fmt.Errorf("note '%c' undefined", note.Name)
		}

		if note.IsSharp() {
			if key == constants.MaxKey {
				return nil, fmt.Errorf("sharp note '%s' out of MIDI range", note)
			}
			key++
		} else if note.IsFlat() {
			if key == constants.MinKey {
				return nil, fmt.Errorf("flat note '%s' out of MIDI range", note)
			}
			key--
		}

		velocity := it.curVelocity
		if note.IsAccent() {
			velocity *= 2
			if velocity > constants.MaxVelocity {
				velocity = constants.MaxVelocity
			}
		} else if note.IsGhost() {
			velocity /= 2
		}

		r := uint16(it.curChannel)<<8 | uint16(key)
		if _, ok := it.ringing[r]; ok {
			delete(it.ringing, r)
			messages = append(messages,
				Message{
					Tick: tick,
					Msg:  midi.NewMessage(midi.Channel(it.curChannel).NoteOff(key)),
				},
			)
		}

		messages = append(messages, Message{
			Tick:       tick,
			NoteLength: length,
			Msg:        midi.NewMessage(midi.Channel(it.curChannel).NoteOn(key, velocity)),
		})

		if note.IsLetRing() {
			it.ringing[r] = struct{}{}
		} else {
			messages = append(messages, Message{
				Tick: tick + length,
				Msg:  midi.NewMessage(midi.Channel(it.curChannel).NoteOff(key)),
			})
		}

		tick += length
	}

	sort.Sort(byMessageTypeOrKey(messages))

	return messages, nil
}

// New creates an interpreter.
func New() *Interpreter {
	return &Interpreter{
		parser:        parser.NewParser(),
		channelKeymap: map[uint8]map[rune]uint8{0: {}},
		ringing:       map[uint16]struct{}{},
		bars:          map[string][]Message{},
		curVelocity:   constants.MaxVelocity,
	}
}

type byMessageTypeOrKey []Message

func (s byMessageTypeOrKey) Len() int      { return len(s) }
func (s byMessageTypeOrKey) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s byMessageTypeOrKey) Less(i, j int) bool {
	if s[i].Tick == s[j].Tick {
		a := s[i].Msg
		b := s[j].Msg

		if a.IsNoteEnd() && !b.IsNoteEnd() {
			return true
		}

		if a.MsgType == b.MsgType {
			return a.Key() < b.Key()
		}

		return false
	}

	return s[i].Tick < s[j].Tick
}

func containsNotes(messages []Message) bool {
	for _, msg := range messages {
		if msg.Msg.IsNoteStart() || msg.Msg.IsNoteEnd() {
			return true
		}
	}
	return false
}
