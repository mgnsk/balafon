package balafon_test

import (
	"embed"
	"fmt"
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

		defer func() {
			if r := recover(); r != nil {
				panic(fmt.Sprintf("panic in file %s: %v", path, r))
			}
		}()

		return it.EvalFile(path)
	}); err != nil {
		t.Fatal(err)
	}
}
