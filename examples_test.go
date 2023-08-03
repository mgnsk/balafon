package balafon_test

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/mgnsk/balafon"
)

//go:embed examples/*
var examples embed.FS

func TestExamples(t *testing.T) {
	var buf bytes.Buffer

	if err := fs.WalkDir(examples, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		defer func() {
			if r := recover(); r != nil {
				panic(fmt.Sprintf("panic in file %s: %v", path, r))
			}
		}()

		song, err := balafon.ToSMF(b)
		if err != nil {
			return err
		}

		buf.Reset()
		if _, err := song.WriteTo(&buf); err != nil {
			return err
		}

		buf.Reset()
		return balafon.ToXML(&buf, b)
	}); err != nil {
		t.Fatal(err)
	}
}
