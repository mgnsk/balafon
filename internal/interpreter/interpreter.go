package interpreter

import (
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/gong/internal/ast"
	"github.com/mgnsk/gong/internal/constants"
	parseError "github.com/mgnsk/gong/internal/parser/errors"
	"github.com/mgnsk/gong/internal/parser/lexer"
	"github.com/mgnsk/gong/internal/parser/parser"
	"github.com/mgnsk/gong/internal/parser/token"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/sequencer"
	"gitlab.com/gomidi/midi/v2/smf"
)

type midiKey struct {
	channel uint8
	note    rune
}

// Interpreter evaluates text input and emits MIDI events.
type Interpreter struct {
	parser *parser.Parser
	keymap map[midiKey]uint8
	// ringing     map[midiKey]struct{}
	song        *sequencer.Song
	buffer      sequencer.Events
	bars        map[string]sequencer.Bar
	curTimeSig  [2]uint8
	curChannel  uint8
	curVelocity uint8
}

// beginBar clones the interpreter but with empty song and buffer.
func (it *Interpreter) beginBar() *Interpreter {
	newIt := New()

	// Use the same maps, assign and bar commands are not allowed in bars.
	newIt.keymap = it.keymap
	newIt.bars = it.bars
	newIt.curTimeSig = it.curTimeSig
	newIt.curChannel = it.curChannel
	newIt.curVelocity = it.curVelocity

	return newIt
}

func (it *Interpreter) Eval(input string) error {
	res, err := it.parser.Parse(lexer.NewLexer([]byte(input)))
	if err != nil {
		return err
	}

	declList, ok := res.(ast.NodeList)
	if !ok {
		return fmt.Errorf("invalid input, expected ast.DeclList")
	}

	return it.EvalAST(declList)
}

func (it *Interpreter) Flush() (song *sequencer.Song) {
	song = it.song

	if len(it.buffer) > 0 {
		if song == nil {
			song = sequencer.New()
		}
		song.AddBar(it.createBar(it.buffer))
	}

	it.song = nil
	it.buffer = it.buffer[:0]

	return song
}

func (it *Interpreter) EvalAST(declList ast.NodeList) error {
	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.Bar:
			if _, ok := it.bars[decl.Name]; ok {
				return fmt.Errorf("bar '%s' already defined", decl.Name)
			}
			// TODO flush?
			// does bar remember previous tempo
			// when used again after another tempo change?

			// the situation: it has buffered events (not notes)
			// a bar is defined, does the buffer also get cloned?

			// TODO: clone but with empty buffer
			itBar := it.beginBar()

			events, err := itBar.parseBar(decl.DeclList)
			if err != nil {
				return err
			}

			it.bars[decl.Name] = sequencer.Bar{
				TimeSig: itBar.curTimeSig,
				Events:  events,
			}

		case ast.CmdPlay:
			bar, ok := it.bars[string(decl)]
			if !ok {
				return fmt.Errorf("unknown bar '%s'", string(decl))
			}

			it.curTimeSig = bar.TimeSig
			if it.song == nil {
				it.song = sequencer.New()
			}
			it.song.AddBar(bar)

		default:
			events, err := it.parseBar(ast.NodeList{decl})
			if err != nil {
				return err
			}

			it.buffer = append(it.buffer, events...)
		}

		if len32th := getLen32th(it.buffer); len32th > 0 {
			if len32th > uint32(getBarLength32th(it.curTimeSig)) {
				return fmt.Errorf("bar too long")
			}
			if it.song == nil {
				it.song = sequencer.New()
			}
			it.song.AddBar(it.createBar(it.buffer))
			it.buffer = it.buffer[:0]
		}
	}

	return nil
}

func getBarLength32th(timeSig [2]uint8) uint8 {
	return timeSig[0] * 32 / timeSig[1]
}

func (it *Interpreter) createBar(events sequencer.Events) sequencer.Bar {
	bar := sequencer.Bar{
		TimeSig: it.curTimeSig,
		Events:  make(sequencer.Events, 0, len(events)),
	}

	for _, ev := range events {
		bar.Events = append(bar.Events, ev)
	}

	return bar
}

// parseBar parses declList as if it were a whole bar.
func (it *Interpreter) parseBar(declList ast.NodeList) (sequencer.Events, error) {
	var events sequencer.Events

	for _, decl := range declList {
		switch decl := decl.(type) {
		case ast.CmdAssign:
			// CmdAssign never occurs in a bar.
			if err := it.assign(it.curChannel, decl.Note, decl.Key); err != nil {
				return nil, err
			}

		case ast.CmdTempo:
			events = append(events, &sequencer.Event{
				Message: smf.MetaTempo(float64(decl)),
			})

		case ast.CmdTimeSig:
			it.curTimeSig = [2]uint8{decl.Num, decl.Denom}

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
			panic(fmt.Sprintf("parse: invalid token %T", decl))
		}
	}

	return events, nil
}

func getLen32th(events sequencer.Events) uint32 {
	var d uint32
	for _, ev := range events {
		d += uint32(ev.Duration)
	}
	return d
}

// func negIndex(tokens []*token.Token, i int) *token.Token {
// 	if len(tokens) > 0 {
// 		return tokens[len(tokens)-1]
// 	}
// 	return nil
// }

// TokenList is a list of tokens that implements parser.Scanner.
type TokenList struct {
	tokens []*token.Token
	cur    int
}

// Scan tokens from input.
func Scan(input string) *TokenList {
	lex := lexer.NewLexer([]byte(input))
	var tokens []*token.Token
	for {
		tok := lex.Scan()
		if tok.Type == token.EOF {
			break
		}
		tokens = append(tokens, tok)
	}
	return &TokenList{
		tokens: tokens,
	}
}

// Append a token to the list and return a new list.
func (l *TokenList) Append(tok *token.Token) *TokenList {
	return &TokenList{
		tokens: append(l.tokens, tok),
		cur:    l.cur,
	}
}

// Get the i-th element if i > 0, otherwise the i-th last element.
// For example -1 is the last element.
func (l *TokenList) Get(i int) *token.Token {
	if i >= 0 && i < len(l.tokens) {
		return l.tokens[i]
	} else if i < 0 && len(l.tokens)+i >= 0 {
		return l.tokens[len(l.tokens)+i]
	}
	return nil
}

// Reset the scanner.
func (l *TokenList) Reset() {
	l.cur = 0
}

// Scan the next token.
func (l *TokenList) Scan() *token.Token {
	if l.cur >= len(l.tokens) {
		return tokEOF
	}
	tok := l.tokens[l.cur]
	l.cur++
	return tok
}

var tokEOF = &token.Token{Type: token.EOF}

// type Scanner interface {
// 	Scan() (tok *token.Token)
// }

// Suggest returns suggestions for the next input.
// It is not safe to call Suggest concurrently
// with Eval.
func (it *Interpreter) Suggest(in prompt.Document) []prompt.Suggest {
	tokens := Scan(in.Text)

	var (
		sug            []prompt.Suggest
		expectedTokens []string
	)

	if _, err := it.parser.Parse(tokens); err != nil {
		var perr *parseError.Error
		if errors.As(err, &perr) {
			expectedTokens = perr.ExpectedTokens
		}
	} else {
		tokens.Reset()

		// Pass an INVALID token to produce an error.
		_, err := it.parser.Parse(tokens.Append(&token.Token{Type: token.INVALID}))
		if err == nil {
			panic("expected a parse error")
		}

		var perr *parseError.Error
		if errors.As(err, &perr) {
			expectedTokens = perr.ExpectedTokens
		}
	}

	// spew.Dump(expectedTokens)

	tokens.Reset()

	for _, text := range expectedTokens {
		switch text {
		// "INVALID":      0,
		// "$":            1,
		// "empty":        2,
		case "terminator":
			if lastTok := tokens.Get(-1); lastTok != nil {
				switch lastTok.Type {
				case token.TokMap.Type("uint"):
					for i := 0; i <= constants.MaxValue; i++ {
						sug = append(sug, prompt.Suggest{
							Text:        strconv.Itoa(i),
							Description: "value",
						})
					}
				}
			}

		case "lineComment":
			sug = append(sug, prompt.Suggest{
				Text:        "//",
				Description: "line comment",
			})

		case "blockComment":
			sug = append(sug, prompt.Suggest{
				Text:        "/*",
				Description: "block comment",
			})

		case "bar":
			sug = append(sug, prompt.Suggest{
				Text:        text,
				Description: "command", // TODO: rename bar to func?
			})

		case "stringLit":
			sug = append(sug, prompt.Suggest{
				Text:        `"`,
				Description: "string",
			})

		case "{", "}":
			sug = append(sug, prompt.Suggest{
				Text: text,
			})

		case "[":
			sug = append(sug, prompt.Suggest{
				Text: text,
			})

		case "]":
			sug = append(sug, prompt.Suggest{
				Text: text,
			})

		case "char":
			if lastTok := tokens.Get(-1); lastTok != nil && lastTok.Type == token.TokMap.Type("assign") {
				// Suggest unassigned keys on the current channel.
				for note := 'a'; note < 'z'; note++ {
					if _, ok := it.keymap[midiKey{it.curChannel, note}]; !ok {
						sug = append(sug, prompt.Suggest{
							Text:        string(note),
							Description: "note",
						})
					}
				}
				for note := 'A'; note < 'Z'; note++ {
					if _, ok := it.keymap[midiKey{it.curChannel, note}]; !ok {
						sug = append(sug, prompt.Suggest{
							Text:        string(note),
							Description: "note",
						})
					}
				}
			} else {
				// Suggest assigned keys on the current channel.
				for note := range it.keymap {
					if note.channel == it.curChannel {
						sug = append(sug, prompt.Suggest{
							Text:        string(note.note),
							Description: "note",
						})
					}
				}
			}

		case "rest":
			sug = append(sug, prompt.Suggest{
				Text:        "-",
				Description: text,
			})

		case "sharp":
			sug = append(sug, prompt.Suggest{
				Text:        "#",
				Description: text,
			})

		case "flat":
			sug = append(sug, prompt.Suggest{
				Text:        "$",
				Description: text,
			})

		case "accent":
			sug = append(sug, prompt.Suggest{
				Text:        "^",
				Description: text,
			})

		case "ghost":
			sug = append(sug, prompt.Suggest{
				Text:        ")",
				Description: text,
			})

		case "uint":
			if last2Tok := tokens.Get(-2); last2Tok != nil {
				switch last2Tok.Type {
				case
					token.TokMap.Type("assign"),
					token.TokMap.Type("timesig"),
					token.TokMap.Type("control"):
					for i := 0; i <= constants.MaxValue; i++ {
						sug = append(sug, prompt.Suggest{
							Text:        strconv.Itoa(i),
							Description: "value",
						})
					}
				}
			}

			if lastTok := tokens.Get(-1); lastTok != nil {
				switch lastTok.Type {
				case
					token.TokMap.Type("tempo"),
					token.TokMap.Type("timesig"),
					token.TokMap.Type("channel"),
					token.TokMap.Type("velocity"),
					token.TokMap.Type("program"),
					token.TokMap.Type("control"):
					for i := 0; i <= constants.MaxValue; i++ {
						sug = append(sug, prompt.Suggest{
							Text:        strconv.Itoa(i),
							Description: "value",
						})
					}
				case token.TokMap.Type("char"):
					// Suggest note value properties.
					for _, value := range []string{"1", "2", "4", "8", "16", "32", "64"} {
						sug = append(sug, prompt.Suggest{
							Text:        value,
							Description: "note value",
						})
					}
				}
			}

		case "dot":
			sug = append(sug, prompt.Suggest{
				Text:        ".",
				Description: text,
			})

		case "tuplet":
			// TODO: from 7th the midi precision gets lost
			sug = append(sug,
				prompt.Suggest{
					Text:        "/3",
					Description: text,
				},
				prompt.Suggest{
					Text:        "/5",
					Description: text,
				},
			)

		case "letRing":
			sug = append(sug, prompt.Suggest{
				Text:        "*",
				Description: "let ring",
			})

		case "assign", "tempo", "timesig", "channel", "velocity", "program", "control", "start", "stop":
			sug = append(sug, prompt.Suggest{
				Text:        text,
				Description: "command",
			})

		case "play":
			for name := range it.bars {
				sug = append(sug, prompt.Suggest{
					Text:        fmt.Sprintf(`play "%s"`, name),
					Description: "command",
				})
			}
		}
	}

	// Don't filter by prefix when suggesting in note lists.
	if lastTok := tokens.Get(-1); lastTok != nil {
		switch lastTok.Type {
		case
			token.TokMap.Type("["),
			token.TokMap.Type("]"),
			token.TokMap.Type("char"),
			token.TokMap.Type("rest"),
			token.TokMap.Type("sharp"),
			token.TokMap.Type("flat"),
			token.TokMap.Type("accent"),
			token.TokMap.Type("ghost"),
			// token.TokMap.Type("uint"),
			token.TokMap.Type("dot"),
			token.TokMap.Type("tuplet"),
			token.TokMap.Type("letRing"):
			return sug
		}
	}

	return prompt.FilterHasPrefix(sug, in.GetWordBeforeCursor(), true)
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
		events   sequencer.Events
		tick32th uint8
	)

	for _, note := range noteList {
		length32th := note.Len()

		if note.IsPause() {
			tick32th += length32th
			continue
		}

		key, ok := it.getKey(it.curChannel, note.Name)
		if !ok {
			return nil, fmt.Errorf("note '%c' undefined", note.Name)
		}

		if note.IsSharp() {
			if key == constants.MaxValue {
				return nil, fmt.Errorf("sharp note '%c' out of MIDI range", note.Name)
			}
			key++
		} else if note.IsFlat() {
			if key == constants.MinValue {
				return nil, fmt.Errorf("flat note '%c' out of MIDI range", note.Name)
			}
			key--
		}

		velocity := it.curVelocity
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

		// TODO
		// if it.isRinging(note.Name) {
		// 	it.setRingingOff(note.Name)
		// }

		// 1th - 32
		// 2th - 16
		// 4th - 8
		// 8th - 4
		// 16th - 2
		// 32th - 1

		events = append(events, &sequencer.Event{
			TrackNo:  int(it.curChannel),
			Pos:      tick32th,
			Duration: length32th,
			Message:  smf.Message(midi.NoteOn(it.curChannel, key, velocity)),
		})

		// if note.IsLetRing() {
		// 	it.setRingingOn(note.Name)
		// }

		// only advance tick if note length > 0

		tick32th += length32th
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
		bars:        map[string]sequencer.Bar{},
		curTimeSig:  [2]uint8{4, 4},
		curVelocity: constants.DefaultVelocity,
	}
}
