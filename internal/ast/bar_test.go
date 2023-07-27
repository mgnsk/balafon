package ast_test

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestBarIdentifierAllowedNumeric(t *testing.T) {
	for _, input := range []string{
		`:bar bar c :end`,
		`:bar 1 c :end`,
		`:play 1`,
		`:bar 1a c :end`,
		`:play 1a`,
	} {
		t.Run(input, func(t *testing.T) {
			g := NewGomegaWithT(t)

			_, err := parse(input)
			g.Expect(err).NotTo(HaveOccurred())
		})
	}
}
