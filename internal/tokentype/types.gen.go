// Code generated by tokentypes-gen. DO NOT EDIT.

package tokentype

import (
	"github.com/mgnsk/balafon/internal/parser/token"
)

// Language tokens.
var (
	BlockComment = token.TokMap.Type("blockComment")
	BracketBegin = token.TokMap.Type("bracketBegin")
	BracketEnd   = token.TokMap.Type("bracketEnd")
	CmdAssign    = token.TokMap.Type("cmdAssign")
	CmdBar       = token.TokMap.Type("cmdBar")
	CmdChannel   = token.TokMap.Type("cmdChannel")
	CmdControl   = token.TokMap.Type("cmdControl")
	CmdEnd       = token.TokMap.Type("cmdEnd")
	CmdPlay      = token.TokMap.Type("cmdPlay")
	CmdProgram   = token.TokMap.Type("cmdProgram")
	CmdStart     = token.TokMap.Type("cmdStart")
	CmdStop      = token.TokMap.Type("cmdStop")
	CmdTempo     = token.TokMap.Type("cmdTempo")
	CmdTimesig   = token.TokMap.Type("cmdTimesig")
	CmdVelocity  = token.TokMap.Type("cmdVelocity")
	Empty        = token.TokMap.Type("empty")
	LineComment  = token.TokMap.Type("lineComment")
	PropAccent   = token.TokMap.Type("propAccent")
	PropDot      = token.TokMap.Type("propDot")
	PropFlat     = token.TokMap.Type("propFlat")
	PropGhost    = token.TokMap.Type("propGhost")
	PropLetRing  = token.TokMap.Type("propLetRing")
	PropMarcato  = token.TokMap.Type("propMarcato")
	PropSharp    = token.TokMap.Type("propSharp")
	PropStaccato = token.TokMap.Type("propStaccato")
	PropTuplet   = token.TokMap.Type("propTuplet")
	Rest         = token.TokMap.Type("rest")
	Symbol       = token.TokMap.Type("symbol")
	Terminator   = token.TokMap.Type("terminator")
	Uint         = token.TokMap.Type("uint")
)
