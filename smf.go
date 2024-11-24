package balafon

import (
	"maps"
	"slices"

	"gitlab.com/gomidi/midi/v2/smf"
)

type track struct {
	track   smf.Track
	lastPos uint32
}

func (a *track) Add(ev TrackEvent) {
	a.track.Add(ev.AbsTicks-a.lastPos, ev.Message)
	a.lastPos = ev.AbsTicks
}

// ToSMF converts a balafon script to SMF1.
func ToSMF(input []byte) (*smf.SMF, error) {
	it := New()

	if err := it.Eval(input); err != nil {
		return nil, err
	}

	bars := it.Flush()

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

	song := smf.New()

	metaTrack.track.Close(0)
	if err := song.Add(metaTrack.track); err != nil {
		return nil, err
	}

	for _, ch := range slices.Sorted(maps.Keys(tracks)) {
		t := tracks[ch]
		t.track.Close(0)
		if err := song.Add(t.track); err != nil {
			return nil, err
		}
	}

	return song, nil
}
