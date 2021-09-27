//go:generate gocc -o internal gong.bnf

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/gong/internal/player"
	"github.com/mgnsk/gong/internal/scanner"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	outs, err := midi.Outs()
	if err != nil {
		fmt.Println(err)
		return
	}

	root := &cobra.Command{
		Short: "gong shell",
		RunE: func(c *cobra.Command, _ []string) error {
			defer func() {
				// Fix Ctrl+C not working after exit (https://github.com/c-bata/go-prompt/issues/228)
				rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
				rawModeOff.Stdin = os.Stdin
				_ = rawModeOff.Run()
				rawModeOff.Wait()
			}()

			fmt.Println("Welcome to the gong shell!")

			port, _ := strconv.Atoi(c.Flag("port").Value.String())
			out := outs[port]

			if err := out.Open(); err != nil {
				return err
			}

			r, w := io.Pipe()
			s := scanner.New(r)
			p := player.New(out)

			ctx, cancel := context.WithCancel(context.Background())

			go func() {
				defer cancel()
				defer w.Close()
				prompt.New(
					func(input string) {
						io.WriteString(w, input+"\n")
					},
					func(in prompt.Document) []prompt.Suggest {
						var sug []prompt.Suggest
						for _, text := range s.Suggest() {
							sug = append(sug, prompt.Suggest{Text: text})
						}
						return prompt.FilterHasPrefix(sug, in.GetWordBeforeCursor(), true)
					},
					prompt.OptionPrefixTextColor(prompt.Yellow),
					prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
					prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
					prompt.OptionSuggestionBGColor(prompt.DarkGray),
				).Run()
			}()

			for {
				for s.Scan() {
					for _, msg := range s.Messages() {
						if err := p.Play(ctx, msg); err != nil {
							if errors.Is(err, context.Canceled) {
								return nil
							}
							return err
						}
					}
				}
				if s.Err() == nil {
					return nil
				}
				fmt.Println(s.Err())
			}
		},
	}

	root.PersistentFlags().String("port", "0", "MIDI output port")

	root.AddCommand(&cobra.Command{
		Use:   "list-ports",
		Short: "List available MIDI output ports",
		Run: func(c *cobra.Command, _ []string) {
			for i, p := range outs {
				fmt.Printf("%d: %s\n", i, p)
			}
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "play [file]",
		Short: "Play a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			port, _ := strconv.Atoi(c.Flag("port").Value.String())
			out := outs[port]

			if err := out.Open(); err != nil {
				return err
			}

			s := scanner.New(io.TeeReader(f, os.Stdout))
			p := player.New(out)

			for s.Scan() {
				for _, msg := range s.Messages() {
					if err := p.Play(context.Background(), msg); err != nil {
						return err
					}
				}
			}

			return s.Err()
		},
	})

	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}
