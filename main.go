//go:generate gocc -o internal/parser gong.bnf

package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
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
)

func main() {
	defer midi.CloseDriver()

	root := &cobra.Command{
		Short: "gong is a MIDI control language and interpreter.",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
		RunE: runShell,
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
		Use:   "play [file]",
		Short: "Play a file",
		Args:  cobra.ExactArgs(1),
		RunE:  playFile,
	})

	if err := root.Execute(); err != nil {
		fmt.Println(err)
	}
}

type result struct {
	input    string
	messages []interpreter.Message
}

func runShell(c *cobra.Command, _ []string) error {
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

	it := interpreter.New()

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

func playFile(c *cobra.Command, args []string) error {
	f, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer f.Close()

	out, err := getPort(c.Flag("port").Value.String())
	if err != nil {
		return err
	}

	if err := out.Open(); err != nil {
		return err
	}

	it := interpreter.New()
	s := bufio.NewScanner(f)

	resultC := make(chan result)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		startPlayer(context.Background(), out, resultC)
	}()

	for s.Scan() {
		input := s.Text()
		messages, err := it.Eval(input)
		if err != nil {
			fmt.Println(input)
			return err
		}
		resultC <- result{input, messages}
	}

	close(resultC)
	wg.Wait()

	return s.Err()
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
