package main_test

import (
	"embed"
	"fmt"
	"io/fs"
	"testing"

	"github.com/mgnsk/balafon/lint"
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

		script, err := examples.ReadFile(path)
		if err != nil {
			return err
		}

		if err := lint.Lint(script); err != nil {
			return fmt.Errorf("error in file '%s': %w", path, err)
		}

		return nil
	}); err != nil {
		t.Fatal(err)
	}
}
