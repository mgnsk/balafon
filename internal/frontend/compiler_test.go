package frontend_test

import (
	"bytes"
	"io/ioutil"
	"strings"
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

	g.Expect(string(b)).To(Equal(`channel 2
assign c 48
assign d 50
channel 10
assign k 36
assign s 38
channel 1
assign c 60
assign d 62

bar "sound A"
channel 2
program 1
channel 10
program 127
channel 1
program 1
end

bar "lead reverb on"
channel 1
control 100 100
end

bar "lead reverb off"
channel 1
control 100 0
end

bar "tempo 2"
tempo 200
end

bar "Verse"
timesig 4 4
channel 2
[cd]2
channel 10
ksks
channel 1
ccdd
[cd]2
end

bar "Fill"
timesig 3 8
channel 10
[ksk]8
end

play "sound A"

play "lead reverb on"

play "Verse"

play "lead reverb off"

play "tempo 2"

play "Fill"

play "Verse"
`))

	it := interpreter.New()
	_, err = it.EvalAll(bytes.NewReader(b))
	g.Expect(err).NotTo(HaveOccurred())
}

func TestCompileInvalidInput(t *testing.T) {
	input := []byte(`
instruments:
  lead:
    channel: 1
bars:
  - name: bar
    tracks:
      bass:
        - a
play:
  - bar
`)

	g := NewGomegaWithT(t)

	_, err := frontend.Compile(input)
	g.Expect(err).To(HaveOccurred())

	cleanMsg := strings.TrimSpace(stripansi.Strip(err.Error()))
	g.Expect(cleanMsg).To(Equal(`missing properties: 'assign':
   2 | instruments:
   3 |   lead:
>  4 |     channel: 1
                  ^
   5 | bars:
   6 |   - name: bar
   7 |     tracks:
instrument 'bass' not defined:
   6 |   - name: bar
   7 |     tracks:
   8 |       bass:
>  9 |         - a
               ^
  10 | play:
  11 |   - bar`))
}
