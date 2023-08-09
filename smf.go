package balafon

import (
	"gitlab.com/gomidi/midi/v2/smf"
	"golang.org/x/exp/slices"
)

type channelTrack struct {
	track   smf.Track
	channel int
}

type track struct {
	track   smf.Track
	lastPos uint32
}

func (a *track) Add(ev TrackEvent) {
	a.track.Add(ev.AbsTicks-a.lastPos, ev.Message)
	a.lastPos = ev.AbsTicks
}

// ToSMF converts a balafon script to SMF2.
func ToSMF(input []byte) (*smf.SMF, error) {
	p := New()

	if err := p.Eval(input); err != nil {
		return nil, err
	}

	bars := p.Flush()

	seq := NewSequencer()
	seq.AddBars(bars...)

	events := seq.Flush()

	metaTrack := &track{}
	tracks := map[uint8]*track{}

	for _, ev := range events {
		if ev.Message == nil {
			continue
		}

		var ch uint8
		if ev.Message.GetChannel(&ch) {
			if tracks[ch] == nil {
				tracks[ch] = &track{}
			}
			tracks[ch].Add(ev)
		} else {
			metaTrack.Add(ev)
		}
	}

	smfTracks := make([]channelTrack, 0, len(tracks)+1)

	metaTrack.track.Close(0)
	smfTracks = append(smfTracks, channelTrack{
		channel: -1,
		track:   metaTrack.track,
	})

	for ch, t := range tracks {
		t.track.Close(0)
		smfTracks = append(smfTracks, channelTrack{
			channel: int(ch),
			track:   t.track,
		})
	}

	slices.SortFunc(smfTracks, func(a, b channelTrack) bool {
		return a.channel < b.channel
	})

	song := smf.New()
	for _, t := range smfTracks {
		song.Add(t.track)
	}

	return song, nil
}
