package frontend_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/mgnsk/gong/internal/frontend"
	"github.com/mgnsk/gong/internal/interpreter"
	. "github.com/onsi/gomega"
)

func TestCompiler(t *testing.T) {
	g := NewGomegaWithT(t)

	input, err := ioutil.ReadFile("../../examples/example.yml")
	g.Expect(err).NotTo(HaveOccurred())

	b, err := frontend.Compile(input)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(string(b)).To(Equal(`channel 1
assign c 60
assign d 62
assign e 64
assign f 65
assign g 67
assign a 69
assign b 71
channel 2
assign C 48
assign D 50
assign E 52
assign F 53
assign G 55
assign A 57
assign B 59
channel 10
assign k 36
assign s 38
assign x 42

bar "setup channels"
channel 1
program 1
control 1 1
channel 2
program 2
control 2 2
channel 10
program 127
end

bar "cool sound preset"
channel 1
program 10
control 10 10
channel 2
program 20
control 20 20
end

bar "Verse"
timesig 4 4
channel 1
cegc
[cg]2
channel 2
CEC2
channel 10
control 3 3
xxxx
ksks
end

play "setup channels"

play "Verse"

play "cool sound preset"

play "Verse"
`))

	it := interpreter.New()
	_, err = it.EvalAll(bytes.NewReader(b))
	g.Expect(err).NotTo(HaveOccurred())
}
