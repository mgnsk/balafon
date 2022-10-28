package main

import (
	"fmt"
	"io"

	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/mgnsk/gong/internal/util"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use:   "gonglint [file]",
		Short: "Lint a gong file.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			f, err := util.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			input, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			if _, err := interpreter.New().Eval(string(input)); err != nil {
				fmt.Println(err)
				return nil
			}

			return nil
		},
	}

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
