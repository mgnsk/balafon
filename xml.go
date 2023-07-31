package balafon

import (
	"time"

	"github.com/eliothedeman/mxl"
	"golang.org/x/exp/slices"
)

// ToXML converts a balafon script to MusicXML.
func ToXML(input []byte) (*mxl.MXLDoc, error) {
	it := New()

	if err := it.Eval(input); err != nil {
		return nil, err
	}

	bars := it.Flush()

	parts := map[uint8]*mxl.Part{}

	for barNo, bar := range bars {
		tracks := map[uint8][]Event{}

		for _, ev := range bar.Events {
			tracks[ev.Track] = append(tracks[ev.Track], ev)
		}

		for trackNo, trackEvents := range tracks {
			part, ok := parts[trackNo]
			if !ok {
				part = &mxl.Part{}
				parts[trackNo] = part
			}

			measure := mxl.Measure{
				Number: barNo,
				Atters: mxl.Attributes{
					Time: mxl.Time{
						Beats:    int(bar.TimeSig[0]),
						BeatType: int(bar.TimeSig[1]),
					},
					Divisions: int(bar.Cap()), // TODO
				},
			}

			for _, ev := range trackEvents {
				if ev.Note != nil {
					measure.Notes = append(measure.Notes, mxl.Note{
						Duration: int(ev.Note.Props.NoteLen()),
						// Type: "whole", // TODO

						// Pitch    Pitch    `xml:"pitch"`
						// Duration int      `xml:"duration"`
						// Voice    int      `xml:"voice"`
						// Type     string   `xml:"type"`
						// Rest     xml.Name `xml:"rest"`
						// Chord    xml.Name `xml:"chord"`
						// Tie      Tie      `xml:"tie"`
					})
				}
			}

			part.Measures = append(part.Measures, measure)
		}
	}

	type trackPart struct {
		track int
		part  mxl.Part
	}

	trackParts := make([]trackPart, 0, len(parts))
	for trackNo, p := range parts {
		trackParts = append(trackParts, trackPart{
			track: int(trackNo),
			part:  *p,
		})
	}

	slices.SortFunc(trackParts, func(a, b trackPart) bool {
		return a.track < b.track
	})

	finalParts := make([]mxl.Part, 0, len(parts))
	for _, p := range trackParts {
		finalParts = append(finalParts, p.part)
	}

	return &mxl.MXLDoc{
		Identification: mxl.Identification{
			Encoding: mxl.Encoding{
				Software: "balafon",
				Date:     time.Now().Format(time.RFC3339),
			},
		},
		Parts: finalParts,
	}, nil
}
