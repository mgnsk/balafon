package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/mgnsk/gong/interpreter"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/testdrv"
	// "gitlab.com/gomidi/midi/v2/drivers"
	// _ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

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

	root.PersistentFlags().String("port", "0", "MIDI output port")

	root.AddCommand(&cobra.Command{
		Use:           "shell",
		Short:         "Run a gong shell",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE: func(c *cobra.Command, _ []string) error {
			runDebugListener()

			out, err := openOut(c.Flag("port").Value.String())
			if err != nil {
				return err
			}

			it := interpreter.New()
			return runPrompt(out, it)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "load [file]",
		Short: "Load a file and continue in a gong shell",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			runDebugListener()

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

			return runPrompt(out, it)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "play [file]",
		Short: "Play a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			runDebugListener()

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

			sh := interpreter.NewShell(out)

			return sh.Execute(it.Flush()...)
		},
	})

	if err := root.Execute(); err != nil {
		panic(err)
	}
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
	sh := interpreter.NewShell(out)

	pt := newBufferedPrompt(
		func(in string) {
			if err := it.Eval(in); err != nil {
				fmt.Println(err)
			} else if err = sh.Execute(it.Flush()...); err != nil {
				fmt.Println(err)
			}
		},
		sh.Complete,
	)

	defer restoreTerminal()
	pt.Run()

	return nil

}

func runDebugListener() {
	in, err := midi.InPort(0)
	if err != nil {
		panic(err)
	}

	midi.ListenTo(in, func(msg midi.Message, timestampms int32) {
		fmt.Println(msg)
	})
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
