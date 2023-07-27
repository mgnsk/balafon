package balafon_test

import (
	"fmt"
	"testing"

	"github.com/mgnsk/balafon"
	. "github.com/onsi/gomega"
)

func TestFmtNewlines(t *testing.T) {
	g := NewWithT(t)

	input := `
:assign c 60
:assign d 62
	`

	res, err := balafon.Format([]byte(input))
	g.Expect(err).NotTo(HaveOccurred())

	fmt.Println(string(res))
}
