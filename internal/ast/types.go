package ast

import (
	"fmt"

	"github.com/mgnsk/gong/internal/parser/token"
)

var (
	charType      = mustGetType("char")
	stringLitType = mustGetType("stringLit")
	sharpType     = mustGetType("sharp")
	flatType      = mustGetType("flat")
	accentType    = mustGetType("accent")
	ghostType     = mustGetType("ghost")
	uintType      = mustGetType("uint")
	dotType       = mustGetType("dot")
	tupletType    = mustGetType("tuplet")
	letRingType   = mustGetType("letRing")
)

func mustGetType(tok string) token.Type {
	t := token.TokMap.Type(tok)
	if t == token.INVALID {
		panic(fmt.Sprintf("invalid token %s", tok))
	}
	return t
}
