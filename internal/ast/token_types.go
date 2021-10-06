package ast

import "github.com/mgnsk/gong/internal/parser/token"

var (
	charType      = token.TokMap.Type("char")
	sharpType     = token.TokMap.Type("sharp")
	flatType      = token.TokMap.Type("flat")
	accentType    = token.TokMap.Type("accent")
	ghostType     = token.TokMap.Type("ghost")
	uintType      = token.TokMap.Type("uint")
	dotType       = token.TokMap.Type("dot")
	tupletType    = token.TokMap.Type("tuplet")
	letRingType   = token.TokMap.Type("letRing")
	stringLitType = token.TokMap.Type("stringLit")
)
