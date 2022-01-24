package frontend_test

import (
	"bytes"
	"io/ioutil"
	"testing"

	"github.com/acarl005/stripansi"
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
channel 2
assign c 48
assign d 50
channel 10
assign k 36
assign s 38

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

bar "tempo 2"
tempo 200
end

bar "Verse"
timesig 4 4
channel 1
cccc
d1
channel 2
dddd
c1
channel 10
ksks
end

play "setup channels"

play "Verse"

play "cool sound preset"

play "tempo 2"

play "Verse"
`))

	it := interpreter.New()
	_, err = it.EvalAll(bytes.NewReader(b))
	g.Expect(err).NotTo(HaveOccurred())
}

func TestCompileInvalidInput(t *testing.T) {
	input := []byte(`
instruments:
  - channel: 1
bars:
  - name: bar
    tracks:
      - channel: 1
        voices:
          - a
play:
  - bar
`)

	g := NewGomegaWithT(t)

	_, err := frontend.Compile(input)
	g.Expect(err).To(HaveOccurred())

	cleanMsg := stripansi.Strip(err.Error())
	g.Expect(cleanMsg).To(Equal(`missing properties: 'assign':
   2 | instruments:
>  3 |   - channel: 1
                  ^
   4 | bars:
   5 |   - name: bar
   6 |     tracks:
note 'a' undefined:
   6 |     tracks:
   7 |       - channel: 1
   8 |         voices:
>  9 |           - a
                   ^
  10 | play:
  11 |   - bar
invalid bar 'bar':
   8 |         voices:
   9 |           - a
  10 | play:
> 11 |   - bar
           ^

`))
}
