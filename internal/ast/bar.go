package ast

// Bar is a bar.
type Bar struct {
	Name string
	Song Song
}

// NewBar creates a new bar.
func NewBar(name string, song interface{}) Bar {
	if s, ok := song.(Song); ok {
		return Bar{Name: name, Song: s}
	}
	return Bar{Name: name}
}
