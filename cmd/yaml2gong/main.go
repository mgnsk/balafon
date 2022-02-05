package main

import (
	"fmt"
	"io/ioutil"

	"github.com/mgnsk/gong/internal/frontend"
	"github.com/mgnsk/gong/internal/util"
	"github.com/spf13/cobra"
)

func main() {
	defer util.HandleExit()

	root := &cobra.Command{
		Use:   "yaml2gong [file]",
		Short: "Compile YAML to gong script.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			f, err := util.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			b, err := ioutil.ReadAll(f)
			if err != nil {
				return err
			}

			script, err := frontend.Compile(b)
			if err != nil {
				return err
			}

			fmt.Print(script)

			return nil
		},
	}

	if err := root.Execute(); err != nil {
		panic(err)
	}
}
