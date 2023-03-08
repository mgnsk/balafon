package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"unicode/utf8"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/balafon/interpreter"
	"github.com/mgnsk/balafon/player"
	"github.com/mgnsk/balafon/sequencer"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"golang.org/x/term"
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

			if err := it.Eval(string(file)); err != nil {
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
			if err := it.Eval(string(file)); err != nil {
				return err
			}

			it.Flush()

			fmt.Println(string(file))

			return runLive(out, it)
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
			if err := it.Eval(string(file)); err != nil {
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

			it := interpreter.New()

			if err := it.Eval(string(input)); err != nil {
				// TODO: lint error format
				fmt.Println(err)
			}

			return nil
		},
	}
	return cmd
}

func runLive(out drivers.Out, it *interpreter.Interpreter) error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	input := make([]byte, 1)

	for {
		_, err := os.Stdin.Read(input)
		if err != nil {
			return err
		}

		r, _ := utf8.DecodeRune(input)
		if r == eot {
			return nil
		}

		if err := it.Eval(string(r)); err != nil {
			fmt.Println(err)
			continue
		}

		for _, bar := range it.Flush() {
			for _, ev := range bar.Events {
				if ev.Message.Is(midi.NoteOnMsg) {
					if err := out.Send(ev.Message); err != nil {
						return err
					}
				}
			}
		}
	}
}

func runPrompt(out drivers.Out, it *interpreter.Interpreter, seq *sequencer.Sequencer) {
	p := player.New(out)

	pt := newBufferedPrompt(
		func(in string) {
			if err := it.Eval(in); err != nil {
				fmt.Println(err)
				return
			}

			seq.AddBars(it.Flush()...)

			events := seq.Flush()

			if err := p.Play(events...); err != nil {
				fmt.Println(err)
				return
			}
		},
		func(in prompt.Document) []prompt.Suggest {
			return nil
		},
	)

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
