package main

import (
	"sort"

	"github.com/mgnsk/gong/internal/interpreter"
	"github.com/mgnsk/gong/internal/util"
	"github.com/spf13/cobra"
	"gitlab.com/gomidi/midi/v2/smf"
)

func main() {
	root := &cobra.Command{
		Use:   "gong2smf [file]",
		Short: "Compile gong script to SMF.",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			f, err := util.Open(args[0])
			if err != nil {
				return err
			}
			defer f.Close()

			it := interpreter.New()
			messages, err := it.EvalAll(f)
			if err != nil {
				return err
			}

			tracks := map[uint8]*midiTrack{}

			// First pass, create tracks.
			for _, msg := range messages {
				var ch uint8
				if msg.GetChannel(&ch) {
					if _, ok := tracks[ch]; !ok {
						tracks[ch] = newMidiTrack(ch)
					}
				}
			}

			// Second pass.
			for _, msg := range messages {
				if msg.Is(smf.MetaTempoMsg) || msg.Is(smf.MetaTimeSigMsg) {
					for _, t := range tracks {
						t.Add(msg)
					}
				} else {
					var ch uint8
					if msg.GetChannel(&ch) {
						tracks[ch].Add(msg)
					}
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
				t.track.Close(0)
				if err := s.Add(t.track); err != nil {
					return err
				}
			}

			return s.WriteFile(c.Flag("output").Value.String())
		},
	}
	root.Flags().StringP("output", "o", "out.mid", "Output file")
	root.MarkFlagRequired("output")

	if err := root.Execute(); err != nil {
		panic(err)
	}
}

type midiTrack struct {
	track    smf.Track
	lastTick uint32
	channel  uint8
}

func newMidiTrack(ch uint8) *midiTrack {
	return &midiTrack{
		track:   smf.Track{},
		channel: ch,
	}
}

func (t *midiTrack) Add(msg interpreter.Message) {
	t.track.Add(msg.Tick-t.lastTick, msg.Message)
	t.lastTick = msg.Tick
}
