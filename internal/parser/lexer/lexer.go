// Code generated by gocc; DO NOT EDIT.

package lexer

import (
	"os"
	"unicode/utf8"

	"github.com/mgnsk/balafon/internal/parser/token"
)

const (
	NoState    = -1
	NumStates  = 115
	NumSymbols = 158
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
	src, err := os.ReadFile(fpath)
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
2: '-'
3: 'b'
4: 'a'
5: 'r'
6: 'e'
7: 'n'
8: 'd'
9: 'p'
10: 'l'
11: 'a'
12: 'y'
13: 'a'
14: 's'
15: 's'
16: 'i'
17: 'g'
18: 'n'
19: 't'
20: 'e'
21: 'm'
22: 'p'
23: 'o'
24: 'k'
25: 'e'
26: 'y'
27: 't'
28: 'i'
29: 'm'
30: 'e'
31: 'v'
32: 'e'
33: 'l'
34: 'o'
35: 'c'
36: 'i'
37: 't'
38: 'y'
39: 'c'
40: 'h'
41: 'a'
42: 'n'
43: 'n'
44: 'e'
45: 'l'
46: 'v'
47: 'o'
48: 'i'
49: 'c'
50: 'e'
51: 'p'
52: 'r'
53: 'o'
54: 'g'
55: 'r'
56: 'a'
57: 'm'
58: 'c'
59: 'o'
60: 'n'
61: 't'
62: 'r'
63: 'o'
64: 'l'
65: 's'
66: 't'
67: 'a'
68: 'r'
69: 't'
70: 's'
71: 't'
72: 'o'
73: 'p'
74: '['
75: ']'
76: '#'
77: '$'
78: '`'
79: '>'
80: '^'
81: ')'
82: '.'
83: '/'
84: '3'
85: '/'
86: '5'
87: '*'
88: '/'
89: '*'
90: '*'
91: '*'
92: '/'
93: '0'
94: ' '
95: '\t'
96: ' '
97: '\t'
98: ':'
99: 'C'
100: 'G'
101: 'D'
102: 'A'
103: 'E'
104: 'B'
105: 'F'
106: '#'
107: 'F'
108: 'B'
109: 'b'
110: 'E'
111: 'b'
112: 'A'
113: 'b'
114: 'D'
115: 'b'
116: 'G'
117: 'b'
118: 'A'
119: 'm'
120: 'E'
121: 'm'
122: 'B'
123: 'm'
124: 'F'
125: '#'
126: 'm'
127: 'C'
128: '#'
129: 'm'
130: 'G'
131: '#'
132: 'm'
133: 'D'
134: '#'
135: 'm'
136: 'D'
137: 'm'
138: 'G'
139: 'm'
140: 'C'
141: 'm'
142: 'F'
143: 'm'
144: 'B'
145: 'b'
146: 'm'
147: 'E'
148: 'b'
149: 'm'
150: ' '
151: '\t'
152: '\r'
153: '1'-'9'
154: '0'-'9'
155: 'a'-'z'
156: 'A'-'Z'
157: .
*/
