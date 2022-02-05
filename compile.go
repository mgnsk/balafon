package main

import (
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/mgnsk/gong/internal/frontend"
	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
)

func compileToGong(c *cobra.Command, args []string) error {
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

type midiTrack struct {
	track    *smf.Track
	lastTick uint32
	channel  uint8
}

func newMidiTrack(ch uint8) *midiTrack {
	return &midiTrack{
		track:   smf.NewTrack(),
		channel: ch,
	}
}

func (t *midiTrack) Add(msg interpreter.Message) {
	t.track.Add(msg.Tick-t.lastTick, msg.Msg.Data)
	t.lastTick = msg.Tick
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
		if msg.Msg.Is(midi.MetaMsg) {
			for _, t := range tracks {
				t.Add(msg)
			}
		} else {
			tracks[msg.Msg.Channel()].Add(msg)
		}
	}

	trackList := make([]*midiTrack, 0, len(tracks))
	for _, track := range tracks {
		trackList = append(trackList, track)
	}

	sort.Slice(trackList, func(i, j int) bool {
		return trackList[i].channel < trackList[j].channel
	})

	s := smf.New()
	for _, t := range trackList {
		s.AddAndClose(0, t.track)
	}

	return s.WriteFile(c.Flag("output").Value.String())
}
