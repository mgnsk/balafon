// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"io/ioutil"
	"unicode/utf8"

	"github.com/mgnsk/gong/internal/parser/token"
)

const (
	NoState    = -1
	NumStates  = 86
	NumSymbols = 102
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
0: ';'
1: '\n'
2: ';'
3: '\n'
4: '"'
5: '"'
6: '-'
7: '#'
8: '$'
9: '^'
10: ')'
11: '.'
12: '/'
13: '3'
14: '/'
15: '5'
16: '*'
17: '/'
18: '/'
19: '\n'
20: '/'
21: '*'
22: '*'
23: '*'
24: '/'
25: 'b'
26: 'a'
27: 'r'
28: 'e'
29: 'n'
30: 'd'
31: '['
32: ']'
33: 'a'
34: 's'
35: 's'
36: 'i'
37: 'g'
38: 'n'
39: 't'
40: 'e'
41: 'm'
42: 'p'
43: 'o'
44: 't'
45: 'i'
46: 'm'
47: 'e'
48: 's'
49: 'i'
50: 'g'
51: 'v'
52: 'e'
53: 'l'
54: 'o'
55: 'c'
56: 'i'
57: 't'
58: 'y'
59: 'c'
60: 'h'
61: 'a'
62: 'n'
63: 'n'
64: 'e'
65: 'l'
66: 'p'
67: 'r'
68: 'o'
69: 'g'
70: 'r'
71: 'a'
72: 'm'
73: 'c'
74: 'o'
75: 'n'
76: 't'
77: 'r'
78: 'o'
79: 'l'
80: 'p'
81: 'l'
82: 'a'
83: 'y'
84: 's'
85: 't'
86: 'a'
87: 'r'
88: 't'
89: 's'
90: 't'
91: 'o'
92: 'p'
93: '0'
94: ' '
95: '\t'
96: '\r'
97: 'a'-'z'
98: 'A'-'Z'
99: '1'-'9'
100: '0'-'9'
101: .
*/
