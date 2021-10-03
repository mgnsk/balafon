package interpreter

import (
	"errors"
	"fmt"
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
	Tempo uint32
}

// Interpreter evaluates messages from raw line input.
type Interpreter struct {
	parser          *parser.Parser
	notes           map[rune]uint8
	bars            map[string][]ast.NoteList
	barBuffer       []ast.NoteList
	currentBar      string
	currentTick     uint64
	currentChannel  uint8
	currentVelocity uint8
}

// Suggest returns suggestions for the next input.
// It is not safe to call Suggest concurrently
// with Eval or EvalString.
func (i *Interpreter) Suggest() []string {
	var sug []string

	// Suggest assigned notes at any time.
	for note := range i.notes {
		sug = append(sug, string(note))
	}

	if i.currentBar != "" {
		// Suggest ending the current bar if we're in the middle of a bar.
		sug = append(sug, "end")
	} else {
		// Suggest commands.
		sug = append(sug, "assign", "tempo", "channel", "velocity", "program", "control", "bar")
		// Suggest playing a bar.
		for name := range i.bars {
			sug = append(sug, "play "+name)
		}
	}

	return sug
}

// Eval evaluates a single input line.
// If both return values are nil, more input is needed.
func (i *Interpreter) Eval(input string) ([]Message, error) {
	if len(strings.TrimSpace(input)) == 0 {
		return nil, nil
	}

	i.parser.Reset()

	res, err := i.parser.Parse(lexer.NewLexer([]byte(input + "\n")))
	if err != nil {
		return nil, err
	}

	if res == nil {
		// Skip comments.
		return nil, nil
	}

	messages, err := i.evalResult(res)
	if err != nil {
		return nil, err
	}

	return messages, nil
}

func (i *Interpreter) evalResult(res interface{}) ([]Message, error) {
	switch r := res.(type) {
	case ast.NoteList:
		if i.currentBar != "" {
			i.barBuffer = append(i.barBuffer, r)
			return nil, nil
		}
		messages, err := i.parseBar(r)
		if err != nil {
			return nil, err
		}
		return messages, nil

	case ast.Command:
		switch r.Name {
		case "assign":
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot assign note: bar '%s' is not ended", i.currentBar)
			}
			v, err := r.Args[1].Int32Value()
			if err != nil {
				panic(err)
			}
			// Guaranteed to be one ASCII character.
			i.notes[[]rune(r.Args[0].IDValue())[0]] = uint8(v)
			return nil, nil
		case "bar": // Begin a bar.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot begin bar '%s': bar '%s' is not ended", r.Args[0].StringValue(), i.currentBar)
			}
			barName := r.Args[0].StringValue()
			if _, ok := i.bars[barName]; ok {
				return nil, fmt.Errorf("bar '%s' already defined", barName)
			}
			i.currentBar = barName
			return nil, nil

		case "end": // End the current bar.
			if i.currentBar == "" {
				return nil, errors.New("cannot end a bar: no active bar")
			}
			i.bars[i.currentBar] = i.barBuffer
			i.currentBar = ""
			i.barBuffer = nil
			return nil, nil

		case "play": // Play a bar.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot play bar '%s': bar '%s' is not ended", r.Args[0].StringValue(), i.currentBar)
			}
			bar, ok := i.bars[r.Args[0].StringValue()]
			if !ok {
				return nil, fmt.Errorf("cannot play nonexistent bar '%s'", r.Args[0].StringValue())
			}
			messages, err := i.parseBar(bar...)
			if err != nil {
				return nil, err
			}
			return messages, nil

		case "tempo": // Set the current tempo.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot change tempo: bar '%s' is not ended", i.currentBar)
			}
			return []Message{{
				Tempo: r.Uint32Args()[0],
			}}, nil

		case "channel": // Set the current channel.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot change channel: bar '%s' is not ended", i.currentBar)
			}
			i.currentChannel = r.Uint8Args()[0]
			return nil, nil

		case "velocity": // Set the current velocity.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot change velocity: bar '%s' is not ended", i.currentBar)
			}
			i.currentVelocity = r.Uint8Args()[0]
			return nil, nil

		case "program": // Program change.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot change program: bar '%s' is not ended", i.currentBar)
			}
			return []Message{{
				Msg: midi.NewMessage(midi.Channel(i.currentChannel).ProgramChange(r.Uint8Args()[0])),
			}}, nil

		case "control": // Control change.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot change control: bar '%s' is not ended", i.currentBar)
			}
			args := r.Uint8Args()
			return []Message{{
				Msg: midi.NewMessage(midi.Channel(i.currentChannel).ControlChange(args[0], args[1])),
			}}, nil

		case "start": // Start message.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot start: bar '%s' is not ended", i.currentBar)
			}
			return []Message{{
				Msg: midi.NewMessage(midi.Start()),
			}}, nil

		case "stop": // Stop message.
			if i.currentBar != "" {
				return nil, fmt.Errorf("cannot stop: bar '%s' is not ended", i.currentBar)
			}
			return []Message{{
				Msg: midi.NewMessage(midi.Stop()),
			}}, nil

		default:
			panic(fmt.Sprintf("invalid command '%s'", r.Name))
		}

	default:
		panic(fmt.Sprintf("invalid token %#v", r))
	}
}

func (i *Interpreter) parseBar(tracks ...ast.NoteList) ([]Message, error) {
	count := 0
	for _, track := range tracks {
		count += len(track) * 2
	}
	messages := make([]Message, count)
	n := 0

	for _, track := range tracks {
		var tick uint64
		for _, note := range track {
			length := noteLength(note)

			if note.Name != '-' {
				key, ok := i.notes[note.Name]
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

				velocity := i.currentVelocity
				if note.IsAccent() {
					if v := 2 * velocity; v <= 127 {
						velocity = v
					} else {
						velocity = 127
					}
				} else if note.IsGhost() {
					velocity /= 2
				}

				messages[n] = Message{
					Tick: i.currentTick + tick,
					Msg:  midi.NewMessage(midi.Channel(i.currentChannel).NoteOn(key, velocity)),
				}

				messages[n+1] = Message{
					Tick: i.currentTick + tick + length,
					Msg:  midi.NewMessage(midi.Channel(i.currentChannel).NoteOff(key)),
				}

				n += 2
			}

			tick += length
		}
	}

	// Sort the messages so that every note is off before on.
	sort.Sort(byMessageTypeOrKey(messages))

	i.currentTick = messages[len(messages)-1].Tick

	return messages, nil
}

// NewInterpreter creates an interpreter.
func NewInterpreter() *Interpreter {
	return &Interpreter{
		parser:          parser.NewParser(),
		notes:           make(map[rune]uint8),
		bars:            make(map[string][]ast.NoteList),
		currentVelocity: 127,
	}
}

func noteLength(note ast.Note) uint64 {
	value := note.Value()
	length := 4 * constants.TicksPerQuarter / uint64(value)
	newLength := length
	for i := uint(0); i < note.Dots(); i++ {
		length = (length / 2)
		newLength += length
	}
	if division := note.Tuplet(); division > 0 {
		newLength = uint64(float64(newLength) * 2.0 / float64(division))
	}
	return newLength
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
