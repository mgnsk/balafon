package main

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"slices"

	"github.com/iancoleman/strcase"
	"github.com/mgnsk/balafon/internal/parser/token"
	. "github.com/moznion/gowrtr/generator"
	"github.com/spf13/cobra"
)

type word struct {
	varName string
	tok     string
}

func main() {
	root := &cobra.Command{
		Use: "gen",
		RunE: func(c *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			_ = cwd

			s := []Statement{
				NewComment(" Code generated by tokentypes-gen. DO NOT EDIT."),
				NewNewline(),
				NewPackage("tokentype"),
				NewImport(
					"github.com/mgnsk/balafon/internal/parser/token",
				),
			}

			var words []word

			typ := token.Type(0)
			for {
				id := token.TokMap.Id(typ)
				if id == "unknown" {
					break
				}

				if typ == token.EOF {
					id = "EOF"
				}

				words = append(words, word{
					varName: strcase.ToCamel(id),
					tok:     id,
				})

				typ++
			}

			// Filter out INVALID and EOF.
			n := 0
			for _, w := range words {
				if w.tok != "INVALID" && w.tok != "EOF" {
					words[n] = w
					n++
				}
			}
			words = words[:n]

			slices.SortFunc(words, func(a, b word) int {
				return cmp.Compare(a.varName, b.varName)
			})

			s = append(s, NewComment("Language tokens."))
			s = append(s, NewRawStatement("var ("))
			for _, w := range words {
				s = append(s, NewRawStatementf(
					"%s = token.TokMap.Type(%q)",
					w.varName,
					w.tok,
				))
			}
			s = append(s, NewRawStatement(")"))

			root, err := NewRoot(s...).Gofmt().Generate(0)
			if err != nil {
				return fmt.Errorf("error generating code: %w", err)
			}

			if err := os.Remove("types.gen.go"); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("error removing previous file: %w", err)
			}

			if err := os.WriteFile("types.gen.go", []byte(root), 0644); err != nil {
				return fmt.Errorf("error writing new file: %w", err)
			}

			return nil
		},
	}

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}
