package scanner

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"sort"
	"sync"

	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/constants"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	"gitlab.com/gomidi/midi/v2"
)

// Message is a tempo or a MIDI message.
type Message struct {
	Tempo uint32
	Tick  uint64
	Msg   midi.Message
}

// Scanner scans messages from raw text input.
type Scanner struct {
	scanner *bufio.Scanner
	parser  *parser.Parser
	err     error

	mu              sync.Mutex
	messages        []Message
	notes           map[rune]uint8
	bars            map[string][]ast.Track
	barBuffer       []ast.Track
	currentBar      string
	currentTick     uint64
	currentChannel  uint8
	currentVelocity uint8
}

// Err returns the first non-EOF error that was encountered by the Scanner.
func (s *Scanner) Err() error {
	if s.err != nil {
		return s.err
	}
	return s.scanner.Err()
}

// Messages returns the currently accumulated messages.
func (s *Scanner) Messages() []Message {
	return s.messages
}

// Suggest returns suggestions for the next input.
func (s *Scanner) Suggest() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Suggest assigned notes at any time.
	// TODO command list
	var sug []string
	// for note := range s.notes {
	// 	sug = append(sug, note)
	// }
	// if s.currentBar != "" {
	// 	// Suggest ending the current bar if we're in the middle of a bar.
	// 	sug = append(sug, "end")
	// } else {
	// 	// Suggest commands.
	// 	// TODO get all commands
	// 	sug = append(sug, "bar", "tempo", "channel", "velocity", "program", "control")
	// 	// Suggest playing a bar.
	// 	for name := range s.bars {
	// 		sug = append(sug, "play "+name)
	// 	}
	// }
	return sug
}

// Scan the next batch of messages.
func (s *Scanner) Scan() bool {
	s.messages = nil
	s.err = nil

	for s.scanner.Scan() {
		if len(bytes.TrimSpace(s.scanner.Bytes())) == 0 {
			continue
		}

		s.parser.Reset()

		input := append(s.scanner.Bytes(), '\n')
		res, err := s.parser.Parse(lexer.NewLexer(input))
		if err != nil {
			s.err = err
			return false
		}

		if res == nil {
			// Skip comments.
			continue
		}

		fmt.Println(res)

		success, wantMore := s.parseResult(res)
		if wantMore {
			continue
		}

		return success
	}

	return false
}

func (s *Scanner) parseResult(res interface{}) (success, wantMore bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch r := res.(type) {
	case ast.NoteAssignment:
		if s.currentBar != "" {
			s.err = fmt.Errorf("cannot assign note: bar '%s' is not ended", s.currentBar)
			return false, false
		}
		s.notes[r.Note] = r.Key
		return true, true

	case ast.Track:
		if s.currentBar != "" {
			s.barBuffer = append(s.barBuffer, r)
			return true, true
		}
		messages, err := s.parseBar(r)
		if err != nil {
			s.err = err
			return false, false
		}
		s.messages = messages
		return true, false

	case ast.Command:
		switch r.Name {
		case "bar": // Begin a bar.
			if s.currentBar != "" {
				s.err = fmt.Errorf("cannot begin bar '%s': bar '%s' is not ended", r.Args[0], s.currentBar)
				return false, false
			}
			if _, ok := s.bars[r.Args[0]]; ok {
				s.err = fmt.Errorf("bar '%s' already defined", r.Args[0])
				return false, false
			}
			s.currentBar = r.Args[0]
			return true, true

		case "end": // End the current bar.
			if s.currentBar == "" {
				s.err = errors.New("cannot end a bar: no active bar")
				return false, false
			}
			s.bars[s.currentBar] = s.barBuffer
			s.currentBar = ""
			s.barBuffer = nil
			return true, true

		case "play": // Play a bar.
			if s.currentBar != "" {
				s.err = fmt.Errorf("cannot play bar '%s': bar '%s' is not ended", r.Args[0], s.currentBar)
				return false, false
			}
			bar, ok := s.bars[r.Args[0]]
			if !ok {
				s.err = fmt.Errorf("cannot play nonexistent bar '%s'", r.Args[0])
				return false, false
			}
			messages, err := s.parseBar(bar...)
			if err != nil {
				s.err = err
				return false, false
			}
			s.messages = messages
			return true, false

		case "tempo": // Set the current tempo.
			if s.currentBar != "" {
				s.err = fmt.Errorf("cannot change tempo: bar '%s' is not ended", s.currentBar)
				return false, false
			}
			s.messages = []Message{{
				Tempo: r.Uint32Args()[0],
			}}
			return true, false

		case "channel": // Set the current channel.
			if s.currentBar != "" {
				s.err = fmt.Errorf("cannot change channel: bar '%s' is not ended", s.currentBar)
				return false, false
			}
			s.currentChannel = r.Uint8Args()[0]
			return true, true

		case "velocity": // Set the current velocity.
			if s.currentBar != "" {
				s.err = fmt.Errorf("cannot change velocity: bar '%s' is not ended", s.currentBar)
				return false, false
			}
			s.currentVelocity = r.Uint8Args()[0]
			return true, true

		case "program": // Program change.
			if s.currentBar != "" {
				s.err = fmt.Errorf("cannot change program: bar '%s' is not ended", s.currentBar)
				return false, false
			}
			s.messages = []Message{{
				Msg: midi.NewMessage(midi.Channel(s.currentChannel).ProgramChange(r.Uint8Args()[0])),
			}}
			return true, false

		case "control": // Control change.
			if s.currentBar != "" {
				s.err = fmt.Errorf("cannot change control: bar '%s' is not ended", s.currentBar)
				return false, false
			}
			args := r.Uint8Args()
			s.messages = []Message{{
				Msg: midi.NewMessage(midi.Channel(s.currentChannel).ControlChange(args[0], args[1])),
			}}
			return true, false

		default:
			panic(fmt.Sprintf("invalid command '%s'", r.Name))
		}

	default:
		panic(fmt.Sprintf("invalid token %#v", r))
	}
}

func (s *Scanner) parseBar(tracks ...ast.Track) ([]Message, error) {
	var messages []Message

	for _, track := range tracks {
		var tick uint64
		for _, note := range track {
			length := noteLength(note)

			if note.Name != '-' {
				key, ok := s.notes[note.Name]
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

				messages = append(messages, Message{
					Tick: s.currentTick + tick,
					Msg:  midi.NewMessage(midi.Channel(s.currentChannel).NoteOn(key, s.currentVelocity)),
				})

				messages = append(messages, Message{
					Tick: s.currentTick + tick + length,
					Msg:  midi.NewMessage(midi.Channel(s.currentChannel).NoteOff(key)),
				})
			}

			tick += length
		}
	}

	// Sort the messages so that every note is off before on.
	sort.Stable(byMessageTypeOrKey(messages))

	s.currentTick = messages[len(messages)-1].Tick

	return messages, nil
}

// New creates a new Scanner instance.
func New(r io.Reader) *Scanner {
	return &Scanner{
		scanner:         bufio.NewScanner(r),
		parser:          parser.NewParser(),
		notes:           make(map[rune]uint8),
		bars:            make(map[string][]ast.Track),
		currentVelocity: 127,
	}
}

func noteLength(note ast.Note) uint64 {
	value := note.Value()
	length := 4 * constants.TicksPerQuarter / uint64(value)
	for i := uint(0); i < note.Dots(); i++ {
		length += (length / 2)
	}
	if division := note.Tuplet(); division > 0 {
		length = uint64(float64(length) * 2.0 / float64(division))
	}
	return length
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
