package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/gong/interpreter"
	"github.com/mgnsk/gong/sequencer"
	"github.com/mgnsk/gong/util"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func addPortFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP("port", "p", "0", "MIDI output port")
}

func main() {
	defer midi.CloseDriver()

	root := &cobra.Command{
		Short: "gong is a MIDI control language and interpreter.",
		Run: func(c *cobra.Command, args []string) {
			fmt.Println("Available MIDI ports:")
			for _, out := range midi.GetOutPorts() {
				fmt.Printf("%d: %s\n", out.Number(), out.String())
			}
		},
	}

	root.AddCommand(createCmdShell())
	root.AddCommand(createCmdLoad())
	root.AddCommand(createCmdPlay())
	root.AddCommand(createCmdLint())

	if err := root.Execute(); err != nil {
		panic(err)
	}
}

func createCmdShell() *cobra.Command {
	cmd := &cobra.Command{
		Use:           "shell",
		Short:         "Run a gong shell",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, _ []string) error {
			out, err := openOut(c.Flag("port").Value.String())
			if err != nil {
				return err
			}

			it := interpreter.New()
			return runPrompt(out, it)
		},
	}
	addPortFlag(cmd)
	return cmd
}

func createCmdLoad() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load [file]",
		Short: "Load a file and continue in a gong shell",
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

			// TODO: need to advance the sequencer
			// to keep latest tempo
			it.Flush()

			fmt.Println(string(file))

			return runPrompt(out, it)
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

			s := sequencer.NewSequencer()
			s.AddBars(it.Flush()...)

			return s.Play(out)
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
				// TODO: lint error format
				fmt.Println(err)
			}

			return nil
		},
	}
	return cmd
}

func restoreTerminal() {
	if strings.Contains(runtime.GOOS, "linux") {
		// TODO: eventually remove this when the bugs get fixed.
		// Fix Ctrl+C not working after exit (https://github.com/c-bata/go-prompt/issues/228)
		rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
		rawModeOff.Stdin = os.Stdin
		_ = rawModeOff.Run()
		rawModeOff.Wait()
	}
}

func runPrompt(out drivers.Out, it *interpreter.Interpreter) error {
	pt := newBufferedPrompt(
		func(in string) {
			if err := it.Eval(in); err != nil {
				fmt.Println(err)
				return
			}
			s := sequencer.NewSequencer()
			s.AddBars(it.Flush()...)
			if err := s.Play(out); err != nil {
				fmt.Println(err)
				return
			}
		},
		func(in prompt.Document) []prompt.Suggest {
			return nil
		},
	)

	defer restoreTerminal()
	pt.Run()

	return nil
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
