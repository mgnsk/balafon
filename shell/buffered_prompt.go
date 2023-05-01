package shell

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/c-bata/go-prompt"
)

// BufferedPrompt is a buffered prompt.
type BufferedPrompt struct {
	pt *prompt.Prompt

	livePrefix        string
	livePrefixEnabled bool
	buffer            bytes.Buffer
}

// NewBufferedPrompt creates a buffered prompt.
func NewBufferedPrompt(
	parser prompt.ConsoleParser,
	writer prompt.ConsoleWriter,
	execute prompt.Executor,
	complete prompt.Completer,
	history []string,
) *BufferedPrompt {
	p := &BufferedPrompt{}

	p.pt = prompt.New(
		func(in string) {
			in = strings.TrimSpace(in)

			if strings.HasPrefix(in, ":bar ") && len(in) > len(":bar ") {
				p.buffer.WriteString(in)
				p.buffer.WriteString("; ")

				p.livePrefix = "...    "
				p.livePrefixEnabled = true

				return
			}

			if strings.HasSuffix(in, ":end") {
				goto finish
			}

			if p.livePrefixEnabled {
				p.buffer.WriteString(in)
				p.buffer.WriteString("; ")

				p.livePrefix = "...    "
				p.livePrefixEnabled = true

				return
			}

		finish:
			p.buffer.WriteString(in)

			p.livePrefix = in
			p.livePrefixEnabled = false

			execute(p.buffer.String())
			p.buffer.Reset()
		},
		complete,
		prompt.OptionHistory(history),
		prompt.OptionShowCompletionAtStart(),
		prompt.OptionCompletionWordSeparator(func() string {
			var s strings.Builder

			s.WriteString(" ")

			return s.String()
		}()),
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
		prompt.OptionParser(parser),
		prompt.OptionWriter(writer),
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

// Run the buffered prompt.
func (p *BufferedPrompt) Run() {
	p.pt.Run()
}

// RestoreTerminal restores terminal after exit.
func RestoreTerminal() {
	if strings.Contains(runtime.GOOS, "linux") {
		// TODO: eventually remove this when the bugs get fixed.
		// Fix Ctrl+C not working after exit (https://github.com/c-bata/go-prompt/issues/228)
		rawModeOff := exec.Command("/bin/stty", "-raw", "echo")
		rawModeOff.Stdin = os.Stdin
		_ = rawModeOff.Run()
		rawModeOff.Wait()
	}
}
