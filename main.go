//go:generate gocc -o internal/parser gong.bnf

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/c-bata/go-prompt"
	"github.com/mgnsk/gong/internal/frontend"
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

	compileSMF := &cobra.Command{
		Use:   "smf [file]",
		Short: "Compile a gong file to SMF",
		Args:  cobra.MaximumNArgs(1),
		RunE:  compileToSMF,
	}
	compileSMF.Flags().StringP("output", "o", "out.mid", "Output file")
	root.AddCommand(compileSMF)

	compileToGong := &cobra.Command{
		Use:   "compile [file]",
		Short: "Compile a YAML file to gong script",
		Args:  cobra.MaximumNArgs(1),
		RunE:  compileYAML,
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
	channel  uint8
	track    *smf.Track
	lastTick uint32
}

func newMidiTrack(ch uint8) *midiTrack {
	return &midiTrack{
		channel: ch,
		track:   smf.NewTrack(),
	}
}

func (t *midiTrack) Add(msg interpreter.Message) {
	t.track.Add(msg.Tick-t.lastTick, msg.Msg.Data)
	t.lastTick = msg.Tick
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

func compileYAML(c *cobra.Command, args []string) error {
	f, err := stdinOrFile(args)
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}

	script, err := frontend.Compile(b)
	if err != nil {
		return err
	}

	fmt.Printf(string(script))

	return nil
}

func compileToSMF(c *cobra.Command, args []string) error {
	f, err := stdinOrFile(args)
	if err != nil {
		return err
	}
	defer f.Close()

	it := interpreter.New()
	messages, err := it.EvalAll(f)
	if err != nil {
		return err
	}

	tracks := map[int8]*midiTrack{}

	// First pass, create tracks.
	for _, msg := range messages {
		if ch := msg.Msg.Channel(); ch >= 0 {
			if _, ok := tracks[ch]; !ok {
				tracks[ch] = newMidiTrack(uint8(ch))
			}
		}
	}

	// Second pass.
	for _, msg := range messages {
		if msg.Msg.Is(midi.MetaTempoMsg) || msg.Msg.Channel() == -1 {
			for _, t := range tracks {
				t.Add(msg)
			}
			continue
		}
		tracks[msg.Msg.Channel()].Add(msg)
	}

	trackList := make([]*midiTrack, 0, len(tracks))
	for _, track := range tracks {
		trackList = append(trackList, track)
	}

	sort.Slice(trackList, func(i, j int) bool {
		return trackList[i].channel < trackList[j].channel
	})

	s := smf.New()
	for _, t := range tracks {
		s.AddAndClose(0, t.track)
	}

	return s.WriteFile(c.Flag("output").Value.String())
}

func playFile(c *cobra.Command, args []string) error {
	f, err := stdinOrFile(args)
	if err != nil {
		return err
	}
	defer f.Close()

	it := interpreter.New()
	messages, err := it.EvalAll(f)
	if err != nil {
		return err
	}

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

func startPlayer(ctx context.Context, out midi.Sender, resultC <-chan result, tempo uint16) {
	runtime.LockOSThread()

	p := player.New(out)
	if tempo > 0 {
		p.SetTempo(tempo)
	}
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
