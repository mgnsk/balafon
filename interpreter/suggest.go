package interpreter

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/balafon/constants"
	parseErrors "github.com/mgnsk/balafon/internal/parser/errors"
	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/token"
	"github.com/mgnsk/balafon/internal/tokentype"
	"github.com/mgnsk/evcache/v3/ringlist"
)

// scannerWithInvalid inserts an invalid token before the EOF token.
type scannerWithInvalid struct {
	lex    *lexer.Lexer
	tokEOF *token.Token
}

type tokenList = ringlist.ElementList[*token.Token]

func (s *scannerWithInvalid) PreScan() *tokenList {
	tokens := new(tokenList)
	for {
		tok := s.lex.Scan()
		if tok.Type == token.EOF {
			break
		}
		tokens.PushBack(ringlist.NewElement(tok))
	}
	s.lex.Reset()
	return tokens
}

func (s *scannerWithInvalid) Scan() *token.Token {
	if s.tokEOF != nil {
		return s.tokEOF
	}
	tok := s.lex.Scan()
	if tok.Type == token.EOF {
		s.tokEOF = tok
		return &token.Token{Type: token.INVALID}
	}
	return tok
}

func newScannerWithInvalid(lex *lexer.Lexer) *scannerWithInvalid {
	return &scannerWithInvalid{
		lex: lex,
	}
}

// Suggest returns suggestions for the next input.
// It is not safe to call Suggest concurrently
// with Eval.
func (it *Interpreter) Suggest(in prompt.Document) []prompt.Suggest {
	var (
		sug            []prompt.Suggest
		expectedTokens []string
	)

	lex := newScannerWithInvalid(lexer.NewLexer([]byte(in.Text)))
	tokens := lex.PreScan()

	if _, err := it.parser.Parse(lex); err != nil {
		var perr *parseErrors.Error
		if errors.As(err, &perr) {
			expectedTokens = perr.ExpectedTokens
		}
	} else {
		panic("expected a parse error")
	}

	for _, text := range expectedTokens {
		switch text {
		case tokentype.Terminator.ID:
			if lastTok := tokens.Back(); lastTok != nil {
				switch lastTok.Value.Type {
				case tokentype.Uint.Type:
					for i := 0; i <= constants.MaxValue; i++ {
						sug = append(sug, prompt.Suggest{
							Text:        strconv.Itoa(i),
							Description: "value",
						})
					}
				}
			}
		case tokentype.CmdBar.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":bar",
				Description: "command",
			})
		case tokentype.CmdEnd.ID:
		case tokentype.BracketBegin.ID, tokentype.BracketEnd.ID:
			sug = append(sug, prompt.Suggest{
				Text:        text,
				Description: "note group",
			})
		case tokentype.Symbol.ID:
			lastTok := tokens.Back()

			if lastTok != nil && lastTok.Value.Type == tokentype.CmdAssign.Type {
				// Suggest unassigned keys on the current channel.
				for note := 'a'; note < 'z'; note++ {
					if _, ok := it.keymap.Get(it.channel, note); !ok {
						sug = append(sug, prompt.Suggest{
							Text:        string(note) + " ",
							Description: "unassigned note",
						})
					}
				}
				for note := 'A'; note < 'Z'; note++ {
					if _, ok := it.keymap.Get(it.channel, note); !ok {
						sug = append(sug, prompt.Suggest{
							Text:        string(note) + " ",
							Description: "unassigned note",
						})
					}
				}
			} else {
				// Suggest assigned keys on the current channel.
				it.keymap.Range(func(channel uint8, note rune, key int) {
					if channel == it.channel {
						sug = append(sug, prompt.Suggest{
							Text:        string(note),
							Description: fmt.Sprintf("note (%d:%d)", channel, key),
						})
					}
				})
			}
		case tokentype.Rest.ID:
			sug = append(sug, prompt.Suggest{
				Text:        "-",
				Description: "rest",
			})
		case tokentype.PropSharp.ID:
			sug = append(sug, prompt.Suggest{
				Text:        "#",
				Description: "sharp property",
			})
		case tokentype.PropFlat.ID:
			sug = append(sug, prompt.Suggest{
				Text:        "$",
				Description: "flat property",
			})
		case tokentype.PropAccent.ID:
			sug = append(sug, prompt.Suggest{
				Text:        "^",
				Description: "accent property",
			})
		case tokentype.PropGhost.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ")",
				Description: "ghost property",
			})
		case tokentype.Uint.ID:
			var last2Tok *token.Token

			back := tokens.Back()
			if back != nil {
				last2Tok = back.Prev().Value
			}

			if last2Tok != nil {
				switch last2Tok.Type {
				case
					tokentype.CmdAssign.Type,
					tokentype.CmdTimesig.Type,
					tokentype.CmdControl.Type:
					for i := 0; i <= constants.MaxValue; i++ {
						sug = append(sug, prompt.Suggest{
							Text:        strconv.Itoa(i),
							Description: "value",
						})
					}
				}
			}

			if lastTok := tokens.Back(); lastTok != nil {
				switch lastTok.Value.Type {
				case
					tokentype.CmdTempo.Type,
					tokentype.CmdTimesig.Type,
					tokentype.CmdChannel.Type,
					tokentype.CmdVelocity.Type,
					tokentype.CmdProgram.Type,
					tokentype.CmdControl.Type:
					for i := 0; i <= constants.MaxValue; i++ {
						sug = append(sug, prompt.Suggest{
							Text:        strconv.Itoa(i),
							Description: "value",
						})
					}
				case tokentype.Symbol.Type:
					// Suggest note value properties.
					for _, value := range []string{"1", "2", "4", "8", "16", "32", "64"} {
						sug = append(sug, prompt.Suggest{
							Text:        value,
							Description: "note value",
						})
					}
				}
			}
		case tokentype.PropDot.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ".",
				Description: "dot property",
			})
		case tokentype.PropTuplet.ID:
			sug = append(sug, prompt.Suggest{
				Text:        "/3",
				Description: "tuplet property",
			})
			sug = append(sug, prompt.Suggest{
				Text:        "/5",
				Description: "tuplet property",
			})
		case tokentype.PropLetRing.ID:
			sug = append(sug, prompt.Suggest{
				Text:        "*",
				Description: "let ring property",
			})
		case tokentype.CmdAssign.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":assign",
				Description: "command",
			})
		case tokentype.CmdPlay.ID:
			for barName := range it.bars {
				sug = append(sug, prompt.Suggest{
					Text:        fmt.Sprintf(`:play "%s" `, barName),
					Description: "command",
				})
			}
		case tokentype.CmdTempo.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":tempo",
				Description: "command",
			})
		case tokentype.CmdTimesig.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":timesig",
				Description: "command",
			})
		case tokentype.CmdVelocity.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":velocity",
				Description: "command",
			})
		case tokentype.CmdChannel.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":channel",
				Description: "command",
			})
		case tokentype.CmdProgram.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":program",
				Description: "command",
			})
		case tokentype.CmdControl.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":control",
				Description: "command",
			})
		case tokentype.CmdStart.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":start",
				Description: "command",
			})
		case tokentype.CmdStop.ID:
			sug = append(sug, prompt.Suggest{
				Text:        ":stop",
				Description: "command",
			})
		}
	}

	// Don't filter by prefix when suggesting in note lists.
	if lastTok := tokens.Back(); lastTok != nil {
		switch lastTok.Value.Type {
		case
			tokentype.BracketBegin.Type,
			tokentype.BracketEnd.Type,
			tokentype.Symbol.Type,
			tokentype.Rest.Type,
			tokentype.PropSharp.Type,
			tokentype.PropFlat.Type,
			tokentype.PropAccent.Type,
			tokentype.PropGhost.Type,
			// tokentype.Uint.Type
			tokentype.PropDot.Type,
			tokentype.PropTuplet.Type,
			tokentype.PropLetRing.Type:
			return sug
		}
	}

	return prompt.FilterHasPrefix(sug, in.GetWordBeforeCursor(), true)
}
