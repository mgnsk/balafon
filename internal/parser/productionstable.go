// Code generated by gocc; DO NOT EDIT.

package parser

import (
    "github.com/mgnsk/gong/internal/ast"
    "github.com/mgnsk/gong/internal/token"
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
		String: `S' : Expr	<<  >>`,
		Id:         "S'",
		NTType:     0,
		Index:      0,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Expr : NoteAssignment	<<  >>`,
		Id:         "Expr",
		NTType:     1,
		Index:      1,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Expr : Track	<<  >>`,
		Id:         "Expr",
		NTType:     1,
		Index:      2,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `Expr : Command	<<  >>`,
		Id:         "Expr",
		NTType:     1,
		Index:      3,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return X[0], nil
		},
	},
	ProdTabEntry{
		String: `NoteAssignment : multiNote "=" uint	<< ast.NewNoteAssignment(X[0].(*token.Token).IDValue(), X[2].(*token.Token).IDValue()) >>`,
		Id:         "NoteAssignment",
		NTType:     2,
		Index:      4,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNoteAssignment(X[0].(*token.Token).IDValue(), X[2].(*token.Token).IDValue())
		},
	},
	ProdTabEntry{
		String: `Track : empty	<<  >>`,
		Id:         "Track",
		NTType:     3,
		Index:      5,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `Track : NoteList Track	<< ast.NewTrack(X[0].(ast.NoteList), X[1]), nil >>`,
		Id:         "Track",
		NTType:     3,
		Index:      6,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewTrack(X[0].(ast.NoteList), X[1]), nil
		},
	},
	ProdTabEntry{
		String: `NoteList : multiNote NotePropertyList	<< ast.NewNoteList(X[0].(*token.Token).IDValue(), X[1]), nil >>`,
		Id:         "NoteList",
		NTType:     4,
		Index:      7,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNoteList(X[0].(*token.Token).IDValue(), X[1]), nil
		},
	},
	ProdTabEntry{
		String: `NotePropertyList : empty	<<  >>`,
		Id:         "NotePropertyList",
		NTType:     5,
		Index:      8,
		NumSymbols: 0,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return nil, nil
		},
	},
	ProdTabEntry{
		String: `NotePropertyList : uint NotePropertyList	<< ast.NewNotePropertyList(X[0].(*token.Token), X[1]), nil >>`,
		Id:         "NotePropertyList",
		NTType:     5,
		Index:      9,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNotePropertyList(X[0].(*token.Token), X[1]), nil
		},
	},
	ProdTabEntry{
		String: `NotePropertyList : dot NotePropertyList	<< ast.NewNotePropertyList(X[0].(*token.Token), X[1]), nil >>`,
		Id:         "NotePropertyList",
		NTType:     5,
		Index:      10,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNotePropertyList(X[0].(*token.Token), X[1]), nil
		},
	},
	ProdTabEntry{
		String: `NotePropertyList : tuplet NotePropertyList	<< ast.NewNotePropertyList(X[0].(*token.Token), X[1]), nil >>`,
		Id:         "NotePropertyList",
		NTType:     5,
		Index:      11,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewNotePropertyList(X[0].(*token.Token), X[1]), nil
		},
	},
	ProdTabEntry{
		String: `Command : "bar" barName	<< ast.NewCommand("bar", X[1].(*token.Token).IDValue()) >>`,
		Id:         "Command",
		NTType:     6,
		Index:      12,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("bar", X[1].(*token.Token).IDValue())
		},
	},
	ProdTabEntry{
		String: `Command : "end"	<< ast.NewCommand("end") >>`,
		Id:         "Command",
		NTType:     6,
		Index:      13,
		NumSymbols: 1,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("end")
		},
	},
	ProdTabEntry{
		String: `Command : "play" barName	<< ast.NewCommand("play", X[1].(*token.Token).IDValue()) >>`,
		Id:         "Command",
		NTType:     6,
		Index:      14,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("play", X[1].(*token.Token).IDValue())
		},
	},
	ProdTabEntry{
		String: `Command : "tempo" uint	<< ast.NewCommand("tempo", X[1].(*token.Token).IDValue()) >>`,
		Id:         "Command",
		NTType:     6,
		Index:      15,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("tempo", X[1].(*token.Token).IDValue())
		},
	},
	ProdTabEntry{
		String: `Command : "channel" uint	<< ast.NewCommand("channel", X[1].(*token.Token).IDValue()) >>`,
		Id:         "Command",
		NTType:     6,
		Index:      16,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("channel", X[1].(*token.Token).IDValue())
		},
	},
	ProdTabEntry{
		String: `Command : "velocity" uint	<< ast.NewCommand("velocity", X[1].(*token.Token).IDValue()) >>`,
		Id:         "Command",
		NTType:     6,
		Index:      17,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("velocity", X[1].(*token.Token).IDValue())
		},
	},
	ProdTabEntry{
		String: `Command : "program" uint	<< ast.NewCommand("program", X[1].(*token.Token).IDValue()) >>`,
		Id:         "Command",
		NTType:     6,
		Index:      18,
		NumSymbols: 2,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("program", X[1].(*token.Token).IDValue())
		},
	},
	ProdTabEntry{
		String: `Command : "control" uint uint	<< ast.NewCommand("control", X[1].(*token.Token).IDValue(), X[2].(*token.Token).IDValue()) >>`,
		Id:         "Command",
		NTType:     6,
		Index:      19,
		NumSymbols: 3,
		ReduceFunc: func(X []Attrib, C interface{}) (Attrib, error) {
			return ast.NewCommand("control", X[1].(*token.Token).IDValue(), X[2].(*token.Token).IDValue())
		},
	},
}
