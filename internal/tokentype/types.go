package tokentype

import (
	"fmt"

	"github.com/mgnsk/balafon/internal/parser/token"
)

// Type is a language token type.
type Type struct {
	ID   string
	Type token.Type
}

// Token types.
var (
	_                = Type{"INVALID", token.INVALID}
	_                = typ("␚")
	_                = typ("empty")
	Terminator       = typ("terminator")
	CmdBar           = typ("cmdBar")
	End              = typ(":end")
	SquareBraceBegin = typ("[")
	SquareBraceEnd   = typ("]")
	Symbol           = typ("symbol")
	Rest             = typ("rest")
	Sharp            = typ("sharp")
	Flat             = typ("flat")
	Accent           = typ("accent")
	Ghost            = typ("ghost")
	Uint             = typ("uint")
	Dot              = typ("dot")
	Tuplet           = typ("tuplet")
	LetRing          = typ("letRing")
	Assign           = typ(":assign")
	CmdPlay          = typ("cmdPlay")
	Tempo            = typ(":tempo")
	Timesig          = typ(":timesig")
	Velocity         = typ(":velocity")
	Channel          = typ(":channel")
	Program          = typ(":program")
	Control          = typ(":control")
	Start            = typ(":start")
	Stop             = typ(":stop")
)

func typ(id string) Type {
	t := token.TokMap.Type(id)
	if t == token.INVALID {
		panic(fmt.Sprintf("invalid token %s", id))
	}
	return Type{
		Type: t,
		ID:   id,
	}
}
