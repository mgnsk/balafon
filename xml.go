package balafon

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/mgnsk/balafon/internal/constants"
	"github.com/mgnsk/balafon/internal/mxl"
	"gitlab.com/gomidi/midi/v2/smf"
	"golang.org/x/exp/slices"
)

// ToXML converts a balafon script to MusicXML.
func ToXML(w io.Writer, input []byte) error {
	it := New()

	if err := it.Eval(input); err != nil {
		return err
	}

	bars := it.Flush()

	var channels []Channel
	{
		seen := map[Channel]struct{}{}
		for _, bar := range bars {
			for _, ev := range bar.Events {
				seen[ev.Channel] = struct{}{}
			}
		}

		for ch := range seen {
			channels = append(channels, ch)
		}

		slices.Sort(channels)
	}

	// The partwise MusicXML structure.
	parts := map[Channel]*mxl.Part{}

	for i, bar := range bars {
		events := map[Channel][]Event{}

		for _, ev := range bar.Events {
			events[ev.Channel] = append(events[ev.Channel], ev)
		}

		for _, ch := range channels {
			p, ok := parts[ch]
			if !ok {
				p = &mxl.Part{
					ID: fmt.Sprintf("ch%d", ch),
				}
				parts[ch] = p
			}

			var key *mxl.Key
			if i == 0 {
				// Set default CMaj on each channel's first bar.
				key = &mxl.Key{
					Fifths: 0,
					Mode:   "major",
				}
			}

			measure := mxl.Measure{
				Number: i + 1,
				Atters: mxl.Attributes{
					Time: &mxl.Time{
						Beats:    int(bar.TimeSig[0]),
						BeatType: int(bar.TimeSig[1]),
					},
					Divisions: int(constants.TicksPerWhole) / int(bar.TimeSig[1]),
					Key:       key,
					// Clef      Clef `xml:"clef"`
				},
			}

			if barEvents, ok := events[ch]; ok {
				// Treat notes of the same voice on the same position as chords.
				chords := map[uint32][]Event{}
				uniqPositions := map[uint32]struct{}{}
				uniqVoices := map[Voice]struct{}{}
				for _, ev := range barEvents {
					chords[ev.Pos] = append(chords[ev.Pos], ev)
					uniqPositions[ev.Pos] = struct{}{}
					uniqVoices[ev.Voice] = struct{}{}
				}

				positions := make([]uint32, 0, len(uniqPositions))
				for pos := range uniqPositions {
					positions = append(positions, pos)
				}

				voices := make([]Voice, 0, len(uniqVoices))
				for voice := range uniqVoices {
					voices = append(voices, voice)
				}

				slices.Sort(positions)
				slices.Sort(voices)

				prevVoiceDur := 0

				for i, voice := range voices {
					if i > 0 && prevVoiceDur > 0 {
						measure.Notes = append(measure.Notes, mxl.Backup{
							Duration: prevVoiceDur,
						})
						prevVoiceDur = 0
					}

					for i, pos := range positions {
						if i == 0 && pos != 0 {
							panic("first position in bar must be zero")
						}

						hasMultipleVoicesInPosition := false
						{
							var prevVoice Voice
							for i, ev := range chords[pos] {
								if i == 0 {
									prevVoice = ev.Voice
								} else if ev.Voice != prevVoice {
									hasMultipleVoicesInPosition = true
									break
								}
							}
						}

						prevNoteDur := 0
						noteCount := 0

						for _, ev := range chords[pos] {
							if ev.Voice != voice {
								continue
							}

							if ev.Note == nil {
								// Meta event.
								var smfKey smf.Key
								if ev.Message.GetMetaKey(&smfKey) {
									var fifths int
									{
										_, sharps, flats := getScale(smfKey.String())
										fifths += len(sharps)
										fifths -= len(flats)
									}

									mode := "major"
									if !smfKey.IsMajor {
										mode = "minor"
									}

									measure.Atters.Key = &mxl.Key{
										Fifths: fifths,
										Mode:   mode,
									}
								}
							} else {
								// Note event.

								var chord *xml.Name
								if noteCount > 0 && !hasMultipleVoicesInPosition {
									chord = &xml.Name{}
								}

								if noteCount > 0 && prevNoteDur > 0 {
									// Multiple notes on same position, same voice, need to back up.
									measure.Notes = append(measure.Notes, mxl.Backup{
										Duration: prevNoteDur,
									})
								}

								dur := int(ev.Note.Props.NoteLen())
								prevVoiceDur += dur
								prevNoteDur = dur

								if ev.Note.IsPause() {
									measure.Notes = append(measure.Notes, mxl.Note{
										// Pitch: mxl.Pitch{
										// 	// Accidental int8   `xml:"alter"`
										// 	Step:   "C",
										// 	Octave: 4,
										// },
										Duration: dur,
										Voice:    int(ev.Voice),
										// Voice    int      `xml:"voice"`
										// Type     string   `xml:"type"`
										Rest: &xml.Name{
											Local: "rest",
										},
										Chord: chord,
										// Chord    xml.Name `xml:"chord"`
										// Tie      Tie      `xml:"tie"`
									})
								} else {
									var c, k, v uint8
									if !ev.Message.GetNoteStart(&c, &k, &v) {
										panic("expected GetNoteStart() to succeed")
									}

									step, octave := getPitch(int(k))
									if len(step) == 1 && ev.IsFlat {
										panic("invariant failure: natural note cannot be flat")
									}

									pitch := &mxl.Pitch{
										Step:   step,
										Octave: octave,
									}

									setSharp := func() {
										tmpStep, tmpOctave := getPitch(int(k - 1))
										pitch.Accidental = 1
										pitch.Step = tmpStep
										pitch.Octave = tmpOctave
									}

									setFlat := func() {
										tmpStep, tmpOctave := getPitch(int(k + 1))
										pitch.Accidental = -1
										pitch.Step = tmpStep
										pitch.Octave = tmpOctave
									}

									if strings.HasSuffix(step, "#") {
										if ev.IsFlat {
											setFlat()
										} else {
											setSharp()
										}
									} else {
										// Note: by interpreter not allowing sharp/flat properties
										// on notes that are assigned to a non-natural key,
										// we avoid an ambiguity here.
										if ev.Note.Props.IsSharp() {
											setSharp()
										} else if ev.Note.Props.IsFlat() {
											setFlat()
										}
									}

									measure.Notes = append(measure.Notes, mxl.Note{
										Pitch:    pitch,
										Duration: dur,
										Voice:    int(ev.Voice),
										Chord:    chord,
										// Type     string   `xml:"type"`
										// Chord    xml.Name `xml:"chord"`
										// Tie      Tie      `xml:"tie"`
									})
								}

								noteCount++
							}
						}
					}
				}
			}

			p.Measures = append(p.Measures, measure)
		}
	}

	tps := make([]mxl.Part, 0, len(parts))
	{
		type channelPart struct {
			part    *mxl.Part
			channel Channel
		}

		tmpParts := make([]channelPart, 0, len(parts))
		for ch, p := range parts {
			tmpParts = append(tmpParts, channelPart{
				channel: ch,
				part:    p,
			})
		}

		slices.SortFunc(tmpParts, func(a, b channelPart) bool {
			return a.channel < b.channel
		})

		for _, p := range tmpParts {
			tps = append(tps, *p.part)
		}
	}

	score := mxl.Score{
		Version: "4.0",
		Parts:   tps,
	}

	for _, p := range tps {
		score.PartList.Parts = append(score.PartList.Parts, mxl.ScorePart{
			ID:   p.ID,
			Name: p.ID,
		})
	}

	if _, err := io.WriteString(w, `<?xml version="1.0" encoding="UTF-8" standalone="no"?>
    <!DOCTYPE score-partwise PUBLIC
    "-//Recordare//DTD MusicXML 4.0 Partwise//EN"
    "http://www.musicxml.org/dtds/partwise.dtd">
`); err != nil {
		return err
	}

	enc := xml.NewEncoder(w)
	enc.Indent("", "    ")
	return enc.Encode(score)
}
