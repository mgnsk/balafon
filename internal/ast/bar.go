package ast

// Bar is a bar.
type Bar struct {
	Name     string
	DeclList DeclList
}

// NewBar creates a new bar.
func NewBar(name string, declList interface{}) Bar {
	if list, ok := declList.(DeclList); ok {
		return Bar{Name: name, DeclList: list}
	}
	return Bar{Name: name}
}
