package lint_test

import (
	"testing"

	"github.com/mgnsk/balafon/lint"
	. "github.com/onsi/gomega"
)

func TestLintErrorFormat(t *testing.T) {
	for _, tc := range []struct {
		script         string
		expectedPrefix string
	}{
		{
			script:         "",
			expectedPrefix: `/myfile:1:1: error:`,
		},
	} {
		t.Run(tc.script, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := lint.Lint("/myfile", []byte(tc.script))
			g.Expect(err.Error()).To(HavePrefix(tc.expectedPrefix))
		})
	}
}
