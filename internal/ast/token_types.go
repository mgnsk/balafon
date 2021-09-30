package ast

import "github.com/mgnsk/gong/internal/parser/token"

var (
	sharpType     = token.TokMap.Type("sharp")
	flatType      = token.TokMap.Type("flat")
	uintType      = token.TokMap.Type("uint")
	dotType       = token.TokMap.Type("dot")
	tupletType    = token.TokMap.Type("tuplet")
	stringLitType = token.TokMap.Type("stringLit")
)
