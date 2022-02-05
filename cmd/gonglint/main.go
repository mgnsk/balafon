package main

import (
	"fmt"

	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/mgnsk/gong/internal/util"
	"github.com/spf13/cobra"
)

func main() {
	defer util.HandleExit()

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

			it := interpreter.New()
			if _, err := it.EvalAll(f); err != nil {
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
