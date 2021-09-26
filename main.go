//go:generate gocc -o internal gong.bnf

package main

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/gong/internal/player"
	"github.com/mgnsk/gong/internal/scanner"
	"github.com/spf13/cobra"

	// replace with e.g. "gitlab.com/gomidi/rtmididrv" for real midi connections
	// driver "gitlab.com/gomidi/midi/testdrv"
	// driver "gitlab.com/gomidi/midicatdrv"
	// driver "gitlab.com/gomidi/portmididrv"

	driver "gitlab.com/gomidi/rtmididrv"
)

func main() {
	drv, err := driver.New()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer drv.Close()

	outs, err := drv.Outs()
	if err != nil {
		fmt.Println(err)
		return
	}

	root := &cobra.Command{
		Short: "gong shell",
		RunE: func(c *cobra.Command, _ []string) error {
			fmt.Println("Welcome to gong shell!")

			port, _ := strconv.Atoi(c.Flag("port").Value.String())
			out := outs[port]

			if err := out.Open(); err != nil {
				return err
			}

			r, w := io.Pipe()
			s := scanner.New(r)

			go prompt.New(
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

			return play(out, s)
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

			return playStrict(out, s)
		},
	})

	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}

// play messages into w.
// Scanner errors are logged.
func play(w io.Writer, s *scanner.Scanner) error {
	p := player.New(w)
	for {
		switch s.Scan() {
		case true:
			for _, msg := range s.Messages() {
				if err := p.Play(msg); err != nil {
					return err
				}
			}
		default:
			fmt.Println(s.Err())
		}
	}
}

// playStrict plays messages into w.
// It returns the first encountered error.
func playStrict(w io.Writer, s *scanner.Scanner) error {
	p := player.New(w)
	for s.Scan() {
		for _, msg := range s.Messages() {
			if err := p.Play(msg); err != nil {
				return err
			}
		}
	}
	return s.Err()
}
