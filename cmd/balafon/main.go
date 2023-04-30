package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/c-bata/go-prompt"
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

	root.AddCommand(createCmdShell())
	root.AddCommand(createCmdLoad())
	root.AddCommand(createCmdLive())
	root.AddCommand(createCmdPlay())
	root.AddCommand(createCmdLint())

	if err := root.Execute(); err != nil {
		log.Fatal(err)
	}
}

func createCmdShell() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "shell",
		Short:         "Run a balafon shell",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, _ []string) error {
			out, err := openOut(c.Flag("port").Value.String())
			if err != nil {
				return err
			}

			it := interpreter.New()
			seq := sequencer.New()
			runPrompt(out, it, seq)

			return nil
		},
	}
	addPortFlag(cmd)
	return cmd
}

func createCmdLoad() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load [file]",
		Short: "Load a file and continue in a balafon shell",
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
			seq := sequencer.New()

			if err := it.Eval(file); err != nil {
				return err
			}

			fmt.Println(string(file))

			seq.AddBars(it.Flush()...)
			seq.Flush()

			runPrompt(out, it, seq)

			return nil
		},
	}
	addPortFlag(cmd)
	return cmd
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

			oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
			if err != nil {
				return err
			}
			defer term.Restore(int(os.Stdin.Fd()), oldState)

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

func runPrompt(out drivers.Out, it *interpreter.Interpreter, seq *sequencer.Sequencer) {
	p := player.New(out)

	pt := shell.NewBufferedPrompt(
		prompt.NewStandardInputParser(),
		prompt.NewStdoutWriter(),
		func(in string) {
			if err := it.EvalString(in); err != nil {
				fmt.Println(err)
				return
			}

			seq.AddBars(it.Flush()...)

			events := seq.Flush()

			if err := p.Play(events...); err != nil {
				fmt.Println(err)
			}
		},
		func(in prompt.Document) []prompt.Suggest {
			// TODO: fix tab key overwrites existing text
			return it.Suggest(in)
		},
	)

	defer shell.RestoreTerminal()
	pt.Run()
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

func openOut(port string) (out drivers.Out, err error) {
	if portNum, perr := strconv.Atoi(port); perr == nil {
		out, err = midi.OutPort(portNum)
	} else {
		out, err = midi.FindOutPort(port)
	}

	if err != nil {
		return nil, err
	}

	if perr := out.Open(); perr != nil {
		return nil, perr
	}

	return out, nil
}
