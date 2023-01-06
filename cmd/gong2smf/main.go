package main

import (
	"io"

	"github.com/mgnsk/gong/interpreter"
	"github.com/mgnsk/gong/util"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2/sequencer"
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

			bars := it.Flush()

			song := sequencer.New()
			for _, bar := range bars {
				song.AddBar(bar)
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
