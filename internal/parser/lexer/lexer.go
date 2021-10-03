// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"io/ioutil"
	"unicode/utf8"

	"github.com/mgnsk/gong/internal/parser/token"
)

const (
	NoState    = -1
	NumStates  = 74
	NumSymbols = 83
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
0: '"'
1: '"'
2: '#'
3: '$'
4: '^'
5: ')'
6: '.'
7: '/'
8: '['
9: ']'
10: 'a'
11: 's'
12: 's'
13: 'i'
14: 'g'
15: 'n'
16: 't'
17: 'e'
18: 'm'
19: 'p'
20: 'o'
21: 'c'
22: 'h'
23: 'a'
24: 'n'
25: 'n'
26: 'e'
27: 'l'
28: 'v'
29: 'e'
30: 'l'
31: 'o'
32: 'c'
33: 'i'
34: 't'
35: 'y'
36: 'p'
37: 'r'
38: 'o'
39: 'g'
40: 'r'
41: 'a'
42: 'm'
43: 'c'
44: 'o'
45: 'n'
46: 't'
47: 'r'
48: 'o'
49: 'l'
50: 'b'
51: 'a'
52: 'r'
53: 'e'
54: 'n'
55: 'd'
56: 'p'
57: 'l'
58: 'a'
59: 'y'
60: 's'
61: 't'
62: 'a'
63: 'r'
64: 't'
65: 's'
66: 't'
67: 'o'
68: 'p'
69: '-'
70: '0'
71: '/'
72: '/'
73: '\n'
74: ' '
75: '\t'
76: '\n'
77: '\r'
78: 'a'-'z'
79: 'A'-'Z'
80: '1'-'9'
81: '0'-'9'
82: .
*/
