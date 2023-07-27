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
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("Available MIDI ports:")
			for _, out := range midi.GetOutPorts() {
				fmt.Printf("%d: %s\n", out.Number(), out.String())
			}
		},
	}

	root.AddCommand(createCmdLive())
	root.AddCommand(createCmdPlay())
	root.AddCommand(createCmdLint())
	root.AddCommand(createCmdFmt())

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
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			stat, err := f.Stat()
			if err != nil {
				return err
			}

			b, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			if err := f.Close(); err != nil {
				return err
			}

			result, err := balafon.Format(b)
			if err != nil {
				if _, e := io.WriteString(os.Stderr, err.Error()); e != nil {
					return e
				}
				os.Exit(1)
			}

			if write {
				return os.WriteFile(args[0], result, stat.Mode())
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
