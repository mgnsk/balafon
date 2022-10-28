package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func main() {
	defer midi.CloseDriver()

	root := &cobra.Command{
		Short: "gong is a MIDI control language and interpreter.",
		RunE: func(c *cobra.Command, args []string) error {
			fmt.Println("Available MIDI ports:")
			for _, out := range midi.GetOutPorts() {
				fmt.Printf("%d: %s\n", out.Number(), out.String())
			}
			return nil
		},
	}

	root.PersistentFlags().String("port", "0", "MIDI output port")

	root.AddCommand(&cobra.Command{
		Use:           "shell",
		Short:         "Run a gong shell",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          createRunShellCommand(nil),
	})

	root.AddCommand(&cobra.Command{
		Use:   "load [file]",
		Short: "Load a file and continue in a gong shell",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			file, err := ioutil.ReadFile(args[0])
			if err != nil {
				return err
			}
			return createRunShellCommand(io.TeeReader(bytes.NewReader(file), os.Stdout))(c, args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "play [file]",
		Short: "Play a file",
		Args:  cobra.ExactArgs(1),
		RunE:  playFile,
	})

	if err := root.Execute(); err != nil {
		panic(err)
	}
}

type result struct {
	input    string
	messages []interpreter.Message
}

func createRunShellCommand(input io.Reader) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, _ []string) error {
		if strings.Contains(runtime.GOOS, "linux") {
			// TODO: eventually remove this when the bugs get fixed.
			defer func() {
				// Fix Ctrl+C not working after exit (https://github.com/c-bata/go-prompt/issues/228)
				rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
				rawModeOff.Stdin = os.Stdin
				_ = rawModeOff.Run()
				rawModeOff.Wait()
			}()
		}

		out, err := getPort(c.Flag("port").Value.String())
		if err != nil {
			return err
		}

		if err := out.Open(); err != nil {
			return err
		}

		it := interpreter.New()

		if input != nil {
			if _, err := it.EvalAll(input); err != nil {
				return err
			}
		}

		resultC := make(chan result)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := runPlayer(ctx, out, resultC, it.Tempo()); err != nil && !errors.Is(err, context.Canceled) {
				panic(err)
			}
		}()

		fmt.Printf("Welcome to the gong shell on MIDI port '%d: %s'!\n", out.Number(), out.String())

		newShell(resultC, it, prompt.NewStandardInputParser()).Run()

		cancel()
		wg.Wait()

		return nil
	}
}

func getPort(port string) (drivers.Out, error) {
	portNum, err := strconv.Atoi(port)
	if err == nil {
		return midi.OutPort(portNum)
	}
	return midi.FindOutPort(port)
}