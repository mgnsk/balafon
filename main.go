//go:generate gocc -o internal/parser gong.bnf

package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/mgnsk/gong/internal/player"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
	"gitlab.com/gomidi/midi/v2/smf"
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
		RunE:          createRunShellCommand(interpreter.New()),
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
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}

			it := interpreter.New()

			_, err = it.EvalAll(io.TeeReader(f, os.Stdout))
			if err != nil {
				fmt.Println(err)
				f.Close()
				return nil
			}
			f.Close()

			return createRunShellCommand(it)(c, args)
		},
	})

	root.AddCommand(&cobra.Command{
		Use:   "play [file]",
		Short: "Play a file",
		Args:  cobra.ExactArgs(1),
		RunE:  playFile,
	})

	compile := &cobra.Command{
		Use:   "compile [file]",
		Short: "Compile a gong file to SMF",
		Args:  cobra.ExactArgs(1),
		RunE:  compileFile,
	}
	compile.Flags().StringP("output", "o", "out.mid", "Output file")
	root.AddCommand(compile)

	root.AddCommand(&cobra.Command{
		Use:   "lint [file]",
		Short: "Lint a file",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			f, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			it := interpreter.New()
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

func createRunShellCommand(it *interpreter.Interpreter) func(*cobra.Command, []string) error {
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

		fmt.Printf("Welcome to the gong shell on MIDI port '%d: %s'!\n", out.Number(), out.String())

		resultC := make(chan result)
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			startPlayer(ctx, out, resultC)
		}()

		prompt.New(
			func(input string) {
				messages, err := it.Eval(input)
				if err != nil {
					fmt.Println(err)
					return
				}
				resultC <- result{"", messages}
			},
			func(in prompt.Document) []prompt.Suggest {
				var sug []prompt.Suggest
				for _, text := range it.Suggest() {
					sug = append(sug, prompt.Suggest{Text: text})
				}
				return prompt.FilterHasPrefix(sug, in.GetWordBeforeCursor(), true)
			},
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

type midiTrack struct {
	track    *smf.Track
	lastTick uint32
}

func newMidiTrack() *midiTrack {
	return &midiTrack{
		track: smf.NewTrack(),
	}
}

func (t *midiTrack) Add(msg interpreter.Message) {
	if msg.Tempo > 0 {
		t.track.Add(msg.Tick-t.lastTick, midi.MetaTempo(float64(msg.Tempo)))
		return
	}
	t.track.Add(msg.Tick-t.lastTick, msg.Msg.Data)
	t.lastTick = msg.Tick
}

func compileFile(c *cobra.Command, args []string) error {
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}

	it := interpreter.New()
	messages, err := it.EvalAll(f)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return nil
	}

	f.Close()

	tracks := map[int8]*midiTrack{}

	// First pass, create tracks.
	for _, msg := range messages {
		if msg.Msg.IsNoteStart() {
			ch := msg.Msg.Channel()
			if _, ok := tracks[ch]; !ok {
				tracks[ch] = newMidiTrack()
			}
		}
	}

	// Second pass.
	for _, msg := range messages {
		if msg.Tempo > 0 {
			for _, t := range tracks {
				t.Add(msg)
			}
			continue
		}
		tracks[msg.Msg.Channel()].Add(msg)
	}

	s := smf.New()
	for _, t := range tracks {
		s.AddAndClose(0, t.track)
	}

	return s.WriteFile(c.Flag("output").Value.String())
}

func playFile(c *cobra.Command, args []string) error {
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}

	it := interpreter.New()
	messages, err := it.EvalAll(f)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return nil
	}

	f.Close()

	out, err := getPort(c.Flag("port").Value.String())
	if err != nil {
		return err
	}

	if err := out.Open(); err != nil {
		return err
	}

	playAll(context.Background(), out, messages)

	return nil
}

func playAll(ctx context.Context, out midi.Sender, messages []interpreter.Message) {
	runtime.LockOSThread()

	p := player.New(out)
	for _, msg := range messages {
		if err := p.Play(ctx, msg); err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			log.Fatal(err)
		}
	}
}

func startPlayer(ctx context.Context, out midi.Sender, resultC <-chan result) {
	runtime.LockOSThread()

	p := player.New(out)
	for {
		select {
		case <-ctx.Done():
			return
		case res, ok := <-resultC:
			if !ok {
				return
			}
			if res.input != "" {
				fmt.Println(res.input)
			}
			for _, msg := range res.messages {
				if err := p.Play(ctx, msg); err != nil {
					if errors.Is(err, context.Canceled) {
						return
					}
					log.Fatal(err)
				}
			}
		}
	}
}

func getPort(port string) (midi.Out, error) {
	portNum, err := strconv.Atoi(port)
	if err == nil {
		return midi.OutByNumber(portNum)
	}
	return midi.OutByName(port)
}
