package main

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/c-bata/go-prompt"
)

const (
	// Keycode for Ctrl+D.
	eot = 4

	defaultReso = 16

	gridBG    = "🟦"
	beatBG    = "⭕"
	currentBG = "🔴"
)

func printGrid(buf *bytes.Buffer, reso, current uint) {
	for i := uint(0); i < reso; i++ {
		if i == current {
			buf.WriteString(currentBG)
		} else if i%4 == 0 {
			buf.WriteString(beatBG)
		} else {
			buf.WriteString(gridBG)
		}
	}
}

type bufferedPrompt struct {
	pt *prompt.Prompt

	livePrefix        string
	livePrefixEnabled bool
	buffer            bytes.Buffer
}

func newBufferedPrompt(execute prompt.Executor, complete prompt.Completer) *bufferedPrompt {
	p := &bufferedPrompt{}

	p.pt = prompt.New(
		func(in string) {
			if strings.HasPrefix(in, "bar") {
				p.buffer.WriteString(in)
				p.buffer.WriteString("; ")

				p.livePrefix = "...    "
				p.livePrefixEnabled = true

				return
			}

			if strings.HasSuffix(in, "end") {
				p.buffer.WriteString(in)

				p.livePrefix = in
				p.livePrefixEnabled = false

				execute(p.buffer.String())
				p.buffer.Reset()

				return
			}

			if p.livePrefixEnabled {
				p.buffer.WriteString(in)
				p.buffer.WriteString("; ")

				p.livePrefix = "...    "
				p.livePrefixEnabled = true

				return
			}

			p.buffer.WriteString(in)
			p.livePrefix = in
			p.livePrefixEnabled = false

			execute(p.buffer.String())
			p.buffer.Reset()

			return
		},
		complete,
		// prompt.OptionCompletionWordSeparator(func() string {
		// 	// TODO: build a list of separators
		// 	s := []rune{' '}
		// 	// Notes.
		// 	for note := 'a'; note < 'z'; note++ {
		// 		s = append(s, note)
		// 	}
		// 	for note := 'A'; note < 'Z'; note++ {
		// 		s = append(s, note)
		// 	}
		// 	// Rest.
		// 	s = append(s, '-')
		// 	// Note group parenthesis.
		// 	s = append(s, '[', ']')
		// 	// Note properties.
		// 	// TODO: tuplet and uint?
		// 	s = append(s, []rune{'#', '$', '^', ')', '.', '*'}...)
		// 	// TODO: dont eval when selecting suggestion
		// 	// need to treat the currently selected as part of buffer
		// 	// buf not evaluate yet
		// 	// TODO how to know when user wants to evaluate?
		// 	// terminator?
		// 	// what is live prefix option?
		// 	// https://github.com/c-bata/go-prompt/issues/25

		// 	return string(s)
		// }()),
		prompt.OptionPrefix(">>> "),
		prompt.OptionLivePrefix(func() (string, bool) {
			return p.livePrefix, p.livePrefixEnabled
		}),
		prompt.OptionPrefixTextColor(prompt.Yellow),
		prompt.OptionPreviewSuggestionTextColor(prompt.Blue),
		prompt.OptionSelectedSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionBGColor(prompt.DarkGray),
	)

	return p
}

func (p *bufferedPrompt) Run() {
	defer restoreTerminal()
	p.pt.Run()
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

// func (s *shell) runMetronomePrinter(ctx context.Context, reso uint) {
// 	tickDuration := time.Duration(float64(time.Minute) / float64(s.it.Tempo()) / float64(constants.TicksPerQuarter))
// 	resoDuration := time.Duration(must(s.it.Parse(fmt.Sprintf("x%d", reso))).(ast.NoteList)[0].Length() * uint32(tickDuration))

// 	ticker := time.NewTicker(resoDuration)
// 	defer ticker.Stop()

// 	var buf bytes.Buffer

// 	i := uint(0)
// 	for {
// 		select {
// 		case <-ctx.Done():
// 			return
// 		case <-ticker.C:
// 			printGrid(&buf, reso, i%reso)

// 			fmt.Print("\r" + buf.String())
// 			buf.Reset()

// 			i++
// 		}
// 	}
// }

// func (s *shell) runLive(reso uint) error {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	go s.runMetronomePrinter(ctx, reso)

// 	for {
// 		b, err := s.parser.Read()
// 		if err != nil {
// 			return err
// 		}

// 		r, _ := utf8.DecodeRune(b)
// 		if r == eot {
// 			return nil
// 		}

// 		messages, err := s.it.NoteOn(r)
// 		if err != nil {
// 			// Ignore errors.
// 			continue
// 		}

// 		s.results <- result{"", messages}
// 	}
// }

// func (s *shell) handleInputLine(input string) error {
// 	if input == "live" {
// 		fmt.Print("Entered live mode. Press Ctrl+D to exit.\n")

// 		if err := s.runLive(defaultReso); err != nil {
// 			return err
// 		}
// 		return nil
// 	}

// 	messages, err := s.it.Eval(input)
// 	if err != nil {
// 		return err
// 	}

// 	s.results <- result{"", messages}

// 	return nil
// }

// func must(res interface{}, err error) interface{} {
// 	if err != nil {
// 		panic(err)
// 	}
// 	return res
// }
