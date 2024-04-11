package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mgnsk/balafon"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"golang.org/x/term"
)

const (
	// Keycode for Ctrl+D.
	eot = 4

	defaultReso = 16

	gridBG    = "🟦"
	beatBG    = "⭕"
	currentBG = "🔴"
)

func addPortFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("port", "p", "0", "MIDI output port")
}

func main() {
	defer midi.CloseDriver()

	root := &cobra.Command{
		Short: "balafon is a MIDI control language and interpreter.",
		RunE: func(c *cobra.Command, args []string) error {
			outs, err := drivers.Outs()
			if err != nil {
				return err
			}

			fmt.Println("Available MIDI ports:")
			for _, out := range outs {
				fmt.Printf("%d: %s\n", out.Number(), out.String())
			}

			return nil
		},
	}

	root.AddCommand(createCmdLive())
	root.AddCommand(createCmdPlay())
	root.AddCommand(createCmdLint())
	root.AddCommand(createCmdFmt())
	root.AddCommand(createCmdSMF())

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createCmdLive() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "live [file]",
		Short: "Load a file and continue in a live shell",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			out, err := openOut(c.Flag("port").Value.String())
			if err != nil {
				return err
			}

			it := balafon.New()
			if err := it.EvalFile(args[0]); err != nil {
				return err
			}

			it.Flush()

			s := balafon.NewLiveShell(os.Stdin, it, out)

			oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
			if err != nil {
				return err
			}
			defer term.Restore(int(os.Stdin.Fd()), oldState)

			err = s.Run()
			if err != nil && errors.Is(err, io.EOF) {
				return nil
			}

			return err
		},
	}
	addPortFlag(cmd)
	return cmd
}

func createCmdPlay() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "play [file]",
		Short: "Play a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			out, err := openOut(c.Flag("port").Value.String())
			if err != nil {
				return err
			}

			it := balafon.New()
			if err := it.EvalFile(args[0]); err != nil {
				return err
			}

			s := balafon.NewSequencer()
			s.AddBars(it.Flush()...)

			events := s.Flush()

			p := balafon.NewPlayer(out)

			return p.Play(events...)
		},
	}
	addPortFlag(cmd)
	return cmd
}

func createCmdLint() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lint [file]",
		Short: "Lint a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			it := balafon.New()

			if err := it.EvalFile(args[0]); err != nil {
				if _, e := io.WriteString(os.Stderr, err.Error()); e != nil {
					return e
				}
				os.Exit(1)
			}

			return nil
		},
	}
	return cmd
}

func createCmdFmt() *cobra.Command {
	var write bool

	cmd := &cobra.Command{
		Use:   "fmt [file]",
		Short: "Format a file",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			if write {
				if err := cobra.ExactArgs(1)(c, args); err != nil {
					return err
				}

				return balafon.FormatFile(args[0])
			}

			b, err := io.ReadAll(os.Stdin)
			if err != nil {
				return err
			}

			result, err := balafon.Format(b)
			if err != nil {
				return err
			}

			if _, err := os.Stdout.Write(result); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().BoolVarP(&write, "write", "w", false, "write result to (source) file instead of stdout")

	return cmd
}

func createCmdSMF() *cobra.Command {
	var (
		isText     bool
		outputFile string
	)

	cmd := &cobra.Command{
		Use:   "smf [file]",
		Short: "Convert a file to SMF2",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if outputFile == "" {
				outputFile = strings.TrimSuffix(args[0], ".bal") + ".mid"
			}

			b, err := os.ReadFile(args[0])
			if err != nil {
				return err
			}

			s, err := balafon.ToSMF(b)
			if err != nil {
				return err
			}

			if isText {
				return os.WriteFile(outputFile, []byte(s.String()), 0644)
			}

			return s.WriteFile(outputFile)
		},
	}

	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "output file")
	cmd.PersistentFlags().BoolVarP(&isText, "text", "t", false, "write SMF as text")

	return cmd
}

func openOut(name string) (out drivers.Out, err error) {
	if portNum, perr := strconv.Atoi(name); perr == nil {
		out, err = midi.OutPort(portNum)
		if err != nil {
			return nil, err
		}
	} else {
		lcPort := strings.ToLower(name)
		for _, p := range midi.GetOutPorts() {
			if strings.Contains(strings.ToLower(p.String()), lcPort) {
				out = p
				break
			}
		}
		if out == nil {
			return nil, fmt.Errorf("can't find MIDI output port %v", name)
		}
	}

	if perr := out.Open(); perr != nil {
		return nil, perr
	}

	return out, nil
}
