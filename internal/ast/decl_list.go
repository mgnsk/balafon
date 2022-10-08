package ast

// DeclList is a list of declarations.
type DeclList []interface{}

// NewDeclList creates a new list of declarations.
func NewDeclList(decl, inner interface{}) (list DeclList) {
	if innerList, ok := inner.(DeclList); ok {
		list = make(DeclList, len(innerList)+1)
		list[0] = decl
		copy(list[1:], innerList)
	} else {
		list = DeclList{decl}
	}

	return list
}
