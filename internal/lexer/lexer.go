// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"io/ioutil"
	"unicode/utf8"

	"github.com/mgnsk/gong/internal/token"
)

const (
	NoState    = -1
	NumStates  = 61
	NumSymbols = 64
)

type Lexer struct {
	src     []byte
	pos     int
	line    int
	column  int
	Context token.Context
}

func NewLexer(src []byte) *Lexer {
	lexer := &Lexer{
		src:     src,
		pos:     0,
		line:    1,
		column:  1,
		Context: nil,
	}
	return lexer
}

// SourceContext is a simple instance of a token.Context which
// contains the name of the source file.
type SourceContext struct {
	Filepath string
}

func (s *SourceContext) Source() string {
	return s.Filepath
}

func NewLexerFile(fpath string) (*Lexer, error) {
	src, err := ioutil.ReadFile(fpath)
	if err != nil {
		return nil, err
	}
	lexer := NewLexer(src)
	lexer.Context = &SourceContext{Filepath: fpath}
	return lexer, nil
}

func (l *Lexer) Scan() (tok *token.Token) {
	tok = &token.Token{}
	if l.pos >= len(l.src) {
		tok.Type = token.EOF
		tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = l.pos, l.line, l.column
		tok.Pos.Context = l.Context
		return
	}
	start, startLine, startColumn, end := l.pos, l.line, l.column, 0
	tok.Type = token.INVALID
	state, rune1, size := 0, rune(-1), 0
	for state != -1 {
		if l.pos >= len(l.src) {
			rune1 = -1
		} else {
			rune1, size = utf8.DecodeRune(l.src[l.pos:])
			l.pos += size
		}

		nextState := -1
		if rune1 != -1 {
			nextState = TransTab[state](rune1)
		}
		state = nextState

		if state != -1 {

			switch rune1 {
			case '\n':
				l.line++
				l.column = 1
			case '\r':
				l.column = 1
			case '\t':
				l.column += 4
			default:
				l.column++
			}

			switch {
			case ActTab[state].Accept != -1:
				tok.Type = ActTab[state].Accept
				end = l.pos
			case ActTab[state].Ignore != "":
				start, startLine, startColumn = l.pos, l.line, l.column
				state = 0
				if start >= len(l.src) {
					tok.Type = token.EOF
				}

			}
		} else {
			if tok.Type == token.INVALID {
				end = l.pos
			}
		}
	}
	if end > start {
		l.pos = end
		tok.Lit = l.src[start:end]
	} else {
		tok.Lit = []byte{}
	}
	tok.Pos.Offset, tok.Pos.Line, tok.Pos.Column = start, startLine, startColumn
	tok.Pos.Context = l.Context

	return
}

func (l *Lexer) Reset() {
	l.pos = 0
}

/*
Lexer symbols:
0: '#'
1: '.'
2: '/'
3: '='
4: 'b'
5: 'a'
6: 'r'
7: 'e'
8: 'n'
9: 'd'
10: 'p'
11: 'l'
12: 'a'
13: 'y'
14: 't'
15: 'e'
16: 'm'
17: 'p'
18: 'o'
19: 'c'
20: 'h'
21: 'a'
22: 'n'
23: 'n'
24: 'e'
25: 'l'
26: 'v'
27: 'e'
28: 'l'
29: 'o'
30: 'c'
31: 'i'
32: 't'
33: 'y'
34: 'p'
35: 'r'
36: 'o'
37: 'g'
38: 'r'
39: 'a'
40: 'm'
41: 'c'
42: 'o'
43: 'n'
44: 't'
45: 'r'
46: 'o'
47: 'l'
48: '-'
49: '0'
50: '"'
51: '"'
52: '/'
53: '/'
54: '\n'
55: ' '
56: '\t'
57: '\n'
58: '\r'
59: '0'-'9'
60: 'a'-'z'
61: 'A'-'Z'
62: '1'-'9'
63: .
*/
