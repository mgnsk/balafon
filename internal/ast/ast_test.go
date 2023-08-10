package ast_test

import (
	"testing"

	"github.com/mgnsk/balafon/internal/ast"
	"github.com/mgnsk/balafon/internal/parser/lexer"
	"github.com/mgnsk/balafon/internal/parser/parser"
	. "github.com/onsi/gomega"
)

func parse(input string) (ast.NodeList, error) {
	lex := lexer.NewLexer([]byte(input))
	p := parser.NewParser()

	v, err := p.Parse(lex)
	if err != nil {
		return nil, err
	}

	return v.(ast.NodeList), nil
}

func TestSyntaxNotAmbigous(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign t 0
:assign e 1
:assign m 2
:assign p 3
:assign o 4
:bar bar
	:time 5 4
	tempo
:end
:play bar
`

	_, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestBarNotAmbigous(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign b 0
:assign a 1
:assign r 2
:bar bar
	bar
:end
:play bar
`

	_, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestPlayNotAmbigous(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign p 0
:assign l 1
:assign a 2
:assign y 3
:bar play
	play
:end
:play play
`

	_, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestUniqueProperties(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign c 60
[c#4/5*]#8/3*
`

	nodeList, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())

	var notes []*ast.Note
	ast.WalkNotes(nodeList, nil, func(note *ast.Note) error {
		notes = append(notes, note)
		return nil
	})

	g.Expect(notes).To(HaveLen(1))
	n := notes[0]

	g.Expect(n.Props).To(HaveLen(4))
	g.Expect(n.Props.IsSharp()).To(BeTrue())
	g.Expect(n.Props.Value()).To(Equal(uint8(8)))
	g.Expect(n.Props.Tuplet()).To(Equal(3))
	g.Expect(n.Props.IsLetRing()).To(BeTrue())
}

func TestAdditiveProperties(t *testing.T) {
	g := NewWithT(t)

	input := ":assign c 60; [c`>^).]`>^)."

	nodeList, err := parse(input)
	g.Expect(err).NotTo(HaveOccurred())

	var notes []*ast.Note
	ast.WalkNotes(nodeList, nil, func(note *ast.Note) error {
		notes = append(notes, note)
		return nil
	})

	g.Expect(notes).To(HaveLen(1))
	n := notes[0]

	g.Expect(n.Props).To(HaveLen(10))
	g.Expect(n.Props.NumStaccato()).To(Equal(2))
	g.Expect(n.Props.NumAccent()).To(Equal(2))
	g.Expect(n.Props.NumMarcato()).To(Equal(2))
	g.Expect(n.Props.NumGhost()).To(Equal(2))
	g.Expect(n.Props.NumDot()).To(Equal(2))
}

func BenchmarkParser(b *testing.B) {
	lex := lexer.NewLexer([]byte(
		"[[[[[k*]/3].]$].8]))",
	))
	p := parser.NewParser()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lex.Reset()
		p.Reset()
		_, err := p.Parse(lex)
		if err != nil {
			b.Fatal(err)
		}
	}
}
