//go:generate gocc -o internal/parser gong.bnf

package main

import (
	"bytes"
	"context"
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
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

func handleExit() {
	if e := recover(); e != nil {
		if err, ok := e.(error); ok {
			fmt.Println(err)
			os.Exit(1)
		}
		panic(e)
	}
}

func main() {
	defer handleExit()
	defer midi.CloseDriver()

	root := &cobra.Command{
		Short: "gong is a MIDI control language and interpreter.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          createRunShellCommand(nil),
	}

	root.PersistentFlags().String("port", "0", "MIDI output port")

	root.AddCommand(&cobra.Command{
		Use:   "list",
		Short: "List available MIDI output ports",
		RunE: func(c *cobra.Command, _ []string) error {
			outs, err := midi.Outs()
			if err != nil {
				return err
			}
			for _, out := range outs {
				fmt.Printf("%d: %s\n", out.Number(), out.String())
			}
			return nil
		},
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
		Args:  cobra.MaximumNArgs(1),
		RunE:  playFile,
	})

	compileToSMF := &cobra.Command{
		Use:   "smf [file]",
		Short: "Compile a gong file to SMF",
		Args:  cobra.MaximumNArgs(1),
		RunE:  compileToSMF,
	}
	compileToSMF.Flags().StringP("output", "o", "out.mid", "Output file")
	root.AddCommand(compileToSMF)

	compileToGong := &cobra.Command{
		Use:   "compile [file]",
		Short: "Compile a YAML file to gong script",
		Args:  cobra.MaximumNArgs(1),
		RunE:  compileToGong,
	}
	root.AddCommand(compileToGong)

	root.AddCommand(&cobra.Command{
		Use:   "lint [file]",
		Short: "Lint a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			f, err := stdinOrFile(args)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := it.EvalAll(f); err != nil {
				fmt.Println(err)
				return nil
			}

			return nil
		},
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

		var tempo uint16
		if input != nil {
			messages, err := it.EvalAll(input)
			if err != nil {
				return err
			}
			for _, msg := range messages {
				if bpm := msg.Msg.BPM(); bpm > 0 {
					tempo = uint16(bpm)
				}
			}
		}

		fmt.Printf("Welcome to the gong shell on MIDI port '%d: %s'!\n", out.Number(), out.String())

		resultC := make(chan result)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			startPlayer(ctx, out, resultC, tempo)
		}()

		parser := prompt.NewStandardInputParser()

		sh := &shell{
			parser:  parser,
			writer:  prompt.NewStandardOutputWriter(),
			results: resultC,
		}

		prompt.New(
			func(input string) {
				if err := sh.handleInputLine(input); err != nil {
					fmt.Println(err)
				}
			},
			func(in prompt.Document) []prompt.Suggest {
				var sug []prompt.Suggest
				for _, text := range it.Suggest() {
					sug = append(sug, prompt.Suggest{Text: text})
				}
				return prompt.FilterHasPrefix(sug, in.GetWordBeforeCursor(), true)
			},
			prompt.OptionParser(parser),
			prompt.OptionPrefixTextColor(prompt.Yellow),
			prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
			prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
			prompt.OptionSuggestionBGColor(prompt.DarkGray),
		).Run()

		cancel()
		wg.Wait()

		return nil
	}
}

func getPort(port string) (midi.Out, error) {
	portNum, err := strconv.Atoi(port)
	if err == nil {
		return midi.OutByNumber(portNum)
	}
	return midi.OutByName(port)
}

func stdinOrFile(args []string) (io.ReadCloser, error) {
	if args[0] == "-" {
		return os.Stdin, nil
	} else if args[0] == "" {
		return nil, fmt.Errorf("file argument or '-' for stdin required")
	}
	f, err := os.Open(args[0])
	if err != nil {
		return nil, err
	}
	return f, nil
}
