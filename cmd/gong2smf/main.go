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
		Use:   "gong2smf [file]",
		Short: "Compile gong script to SMF.",
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

			it := interpreter.New()

			if err := it.Eval(string(input)); err != nil {
				return err
			}

			song := it.Flush()
			if song == nil {
				return fmt.Errorf("empty input")
			}

			s := song.ToSMF1()

			return s.WriteFile(c.Flag("output").Value.String())
		},
	}
	root.Flags().StringP("output", "o", "out.mid", "Output file")
	root.MarkFlagRequired("output")

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
