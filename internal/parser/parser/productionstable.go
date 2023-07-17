// Code generated by gocc; DO NOT EDIT.

package parser

import (
	"github.com/mgnsk/balafon/internal/ast"
	"github.com/mgnsk/balafon/internal/parser/token"
)

type (
	ProdTab      [numProductions]ProdTabEntry
	ProdTabEntry struct {
		String     string
		Id         string
		NTType     int
		Index      int
		NumSymbols int
		ReduceFunc func([]Attrib, interface{}) (Attrib, error)
	}
	Attrib interface {
	}
)

var productionsTable = ProdTab{
	ProdTabEntry{
		String:     `S' : SourceFile	<<  >>`,
		Id:         "S'",
		NTType:     0,
		Index:      0,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `SourceFile : RepeatTerminator TopLevelDeclList	<< X[1], nil >>`,
		Id:         "SourceFile",
		NTType:     1,
		Index:      1,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[1], nil
		},
	},
	ProdTabEntry{
		String:     `RepeatTerminator : empty	<<  >>`,
		Id:         "RepeatTerminator",
		NTType:     2,
		Index:      2,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String:     `RepeatTerminator : terminator RepeatTerminator	<<  >>`,
		Id:         "RepeatTerminator",
		NTType:     2,
		Index:      3,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `TopLevelDeclList : TopLevelDecl terminator RepeatTerminator TopLevelDeclList	<< ast.NewNodeList(X[0].(ast.Node), X[3].(ast.NodeList)), nil >>`,
		Id:         "TopLevelDeclList",
		NTType:     3,
		Index:      4,
		NumSymbols: 4,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNodeList(X[0].(ast.Node), X[3].(ast.NodeList)), nil
		},
	},
	ProdTabEntry{
		String:     `TopLevelDeclList : TopLevelDecl RepeatTerminator	<< ast.NewNodeList(X[0].(ast.Node), nil), nil >>`,
		Id:         "TopLevelDeclList",
		NTType:     3,
		Index:      5,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNodeList(X[0].(ast.Node), nil), nil
		},
	},
	ProdTabEntry{
		String:     `BarBodyDeclList : BarBodyDecl terminator RepeatTerminator BarBodyDeclList	<< ast.NewNodeList(X[0].(ast.Node), X[3].(ast.NodeList)), nil >>`,
		Id:         "BarBodyDeclList",
		NTType:     4,
		Index:      6,
		NumSymbols: 4,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNodeList(X[0].(ast.Node), X[3].(ast.NodeList)), nil
		},
	},
	ProdTabEntry{
		String:     `BarBodyDeclList : BarBodyDecl RepeatTerminator	<< ast.NewNodeList(X[0].(ast.Node), nil), nil >>`,
		Id:         "BarBodyDeclList",
		NTType:     4,
		Index:      7,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNodeList(X[0].(ast.Node), nil), nil
		},
	},
	ProdTabEntry{
		String:     `TopLevelDecl : Bar	<<  >>`,
		Id:         "TopLevelDecl",
		NTType:     5,
		Index:      8,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `TopLevelDecl : TopLevelCommand	<<  >>`,
		Id:         "TopLevelDecl",
		NTType:     5,
		Index:      9,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `TopLevelDecl : NoteList	<<  >>`,
		Id:         "TopLevelDecl",
		NTType:     5,
		Index:      10,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `BarBodyDecl : BarBodyCommand	<<  >>`,
		Id:         "BarBodyDecl",
		NTType:     6,
		Index:      11,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `BarBodyDecl : NoteList	<<  >>`,
		Id:         "BarBodyDecl",
		NTType:     6,
		Index:      12,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Bar : cmdBar RepeatTerminator BarBodyDeclList cmdEnd	<< ast.NewBar(string(X[0].(*token.Token).Lit[len(":bar "):]), X[2].(ast.NodeList)), nil >>`,
		Id:         "Bar",
		NTType:     7,
		Index:      13,
		NumSymbols: 4,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewBar(string(X[0].(*token.Token).Lit[len(":bar "):]), X[2].(ast.NodeList)), nil
		},
	},
	ProdTabEntry{
		String:     `NoteList : NoteObject	<< ast.NewNoteList(X[0].(ast.Node), nil), nil >>`,
		Id:         "NoteList",
		NTType:     8,
		Index:      14,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNoteList(X[0].(ast.Node), nil), nil
		},
	},
	ProdTabEntry{
		String:     `NoteList : NoteObject NoteList	<< ast.NewNoteList(X[0].(ast.Node), X[1].(ast.NoteList)), nil >>`,
		Id:         "NoteList",
		NTType:     8,
		Index:      15,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNoteList(X[0].(ast.Node), X[1].(ast.NoteList)), nil
		},
	},
	ProdTabEntry{
		String:     `NoteObject : NoteSymbol PropertyList	<< ast.NewNote([]rune(string(X[0].(*token.Token).Lit))[0], X[1].(ast.PropertyList)), nil >>`,
		Id:         "NoteObject",
		NTType:     9,
		Index:      16,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNote([]rune(string(X[0].(*token.Token).Lit))[0], X[1].(ast.PropertyList)), nil
		},
	},
	ProdTabEntry{
		String:     `NoteObject : bracketBegin NoteList bracketEnd PropertyList	<< ast.NewNoteListFromGroup(X[1].(ast.NoteList), X[3].(ast.PropertyList)) >>`,
		Id:         "NoteObject",
		NTType:     9,
		Index:      17,
		NumSymbols: 4,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNoteListFromGroup(X[1].(ast.NoteList), X[3].(ast.PropertyList))
		},
	},
	ProdTabEntry{
		String:     `NoteSymbol : symbol	<<  >>`,
		Id:         "NoteSymbol",
		NTType:     10,
		Index:      18,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `NoteSymbol : rest	<<  >>`,
		Id:         "NoteSymbol",
		NTType:     10,
		Index:      19,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `PropertyList : empty	<< ast.PropertyList(nil), nil >>`,
		Id:         "PropertyList",
		NTType:     11,
		Index:      20,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.PropertyList(nil), nil
		},
	},
	ProdTabEntry{
		String:     `PropertyList : Property PropertyList	<< ast.NewPropertyList(X[0].(*token.Token), X[1]) >>`,
		Id:         "PropertyList",
		NTType:     11,
		Index:      21,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewPropertyList(X[0].(*token.Token), X[1])
		},
	},
	ProdTabEntry{
		String:     `Property : propSharp	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      22,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propFlat	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      23,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propStaccato	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      24,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propAccent	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      25,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propMarcato	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      26,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propGhost	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      27,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : uint	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      28,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propDot	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      29,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propTuplet	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      30,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `Property : propLetRing	<<  >>`,
		Id:         "Property",
		NTType:     12,
		Index:      31,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `TopLevelCommand : cmdAssign symbol uint	<< ast.NewCmdAssign([]rune(string(X[1].(*token.Token).Lit))[0], ast.Must(X[2].(*token.Token).Int64Value())) >>`,
		Id:         "TopLevelCommand",
		NTType:     13,
		Index:      32,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdAssign([]rune(string(X[1].(*token.Token).Lit))[0], ast.Must(X[2].(*token.Token).Int64Value()))
		},
	},
	ProdTabEntry{
		String:     `TopLevelCommand : cmdPlay	<< ast.NewCmdPlay(string(X[0].(*token.Token).Lit[len(":play "):])) >>`,
		Id:         "TopLevelCommand",
		NTType:     13,
		Index:      33,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdPlay(string(X[0].(*token.Token).Lit[len(":play "):]))
		},
	},
	ProdTabEntry{
		String:     `TopLevelCommand : BarBodyCommand	<<  >>`,
		Id:         "TopLevelCommand",
		NTType:     13,
		Index:      34,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdTempo uint	<< ast.NewCmdTempo(ast.Must(X[1].(*token.Token).Int64Value())) >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      35,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdTempo(ast.Must(X[1].(*token.Token).Int64Value()))
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdTimesig uint uint	<< ast.NewCmdTimeSig(ast.Must(X[1].(*token.Token).Int64Value()), ast.Must(X[2].(*token.Token).Int64Value())) >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      36,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdTimeSig(ast.Must(X[1].(*token.Token).Int64Value()), ast.Must(X[2].(*token.Token).Int64Value()))
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdVelocity uint	<< ast.NewCmdVelocity(ast.Must(X[1].(*token.Token).Int64Value())) >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      37,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdVelocity(ast.Must(X[1].(*token.Token).Int64Value()))
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdChannel uint	<< ast.NewCmdChannel(ast.Must(X[1].(*token.Token).Int64Value())) >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      38,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdChannel(ast.Must(X[1].(*token.Token).Int64Value()))
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdProgram uint	<< ast.NewCmdProgram(ast.Must(X[1].(*token.Token).Int64Value())) >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      39,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdProgram(ast.Must(X[1].(*token.Token).Int64Value()))
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdControl uint uint	<< ast.NewCmdControl(ast.Must(X[1].(*token.Token).Int64Value()), ast.Must(X[2].(*token.Token).Int64Value())) >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      40,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCmdControl(ast.Must(X[1].(*token.Token).Int64Value()), ast.Must(X[2].(*token.Token).Int64Value()))
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdStart	<< ast.CmdStart{}, nil >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      41,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.CmdStart{}, nil
		},
	},
	ProdTabEntry{
		String:     `BarBodyCommand : cmdStop	<< ast.CmdStop{}, nil >>`,
		Id:         "BarBodyCommand",
		NTType:     14,
		Index:      42,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.CmdStop{}, nil
		},
	},
}
