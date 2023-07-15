package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/mgnsk/balafon/interpreter"
	"github.com/mgnsk/balafon/lint"
	"github.com/mgnsk/balafon/player"
	"github.com/mgnsk/balafon/sequencer"
	"github.com/mgnsk/balafon/shell"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"gitlab.com/gomidi/midi/v2/smf"
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
			file, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}

			out, err := openOut(c.Flag("port").Value.String())
			if err != nil {
				return err
			}

			it := interpreter.New()
			if err := it.Eval(file); err != nil {
				return err
			}

			it.Flush()
			fmt.Println(string(file))

			s := shell.NewLiveShell(os.Stdin, it, func(msg smf.Message) error {
				if msg.Is(midi.NoteOnMsg) {
					if err := out.Send(msg); err != nil {
						return err
					}
				}
				return nil
			})

			return s.Run()
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
			file, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}

			out, err := openOut(c.Flag("port").Value.String())
			if err != nil {
				return err
			}

			it := interpreter.New()
			if err := it.Eval(file); err != nil {
				return err
			}

			s := sequencer.New()
			s.AddBars(it.Flush()...)

			events := s.Flush()

			p := player.New(out)

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
			f, err := openInputFile(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			input, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			// TODO: stdin filename argument for error formatting
			return lint.Lint(args[0], input)
		},
	}
	return cmd
}

func openInputFile(name string) (io.ReadCloser, error) {
	if name == "-" {
		return os.Stdin, nil
	} else if name == "" {
		return nil, fmt.Errorf("file argument or '-' for stdin required")
	}

	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}

	return f, nil
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
