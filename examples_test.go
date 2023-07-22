package balafon_test

import (
	"embed"
	"io/fs"
	"testing"

	"github.com/mgnsk/balafon"
)

//go:embed examples/*
var examples embed.FS

func TestExamples(t *testing.T) {
	if err := fs.WalkDir(examples, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		it := balafon.New()

		return it.EvalFile(path)
	}); err != nil {
		t.Fatal(err)
	}
}
