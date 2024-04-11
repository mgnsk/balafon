package balafon

import (
	"cmp"
	"encoding/xml"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/mgnsk/balafon/internal/constants"
	"github.com/mgnsk/balafon/internal/mxl"
	"gitlab.com/gomidi/midi/v2/smf"
)

// ToXML converts a balafon script to MusicXML.
func ToXML(w io.Writer, input []byte) error {
	it := New()

	if err := it.Eval(input); err != nil {
		return err
	}

	bars := it.Flush()

	var tracks []uint8
	{
		seen := map[uint8]struct{}{}
		for _, bar := range bars {
			for _, ev := range bar.Events {
				seen[ev.Track] = struct{}{}
			}
		}

		for ch := range seen {
			tracks = append(tracks, ch)
		}

		slices.Sort(tracks)
	}

	// The partwise MusicXML structure.
	parts := make(map[uint8]*mxl.Part, len(tracks))
	timesig := [2]uint8{4, 4}

	for i, bar := range bars {
		events := map[uint8][]Event{}

		for _, ev := range bar.Events {
			events[ev.Track] = append(events[ev.Track], ev)
		}

		for _, tr := range tracks {
			p, ok := parts[tr]
			if !ok {
				p = &mxl.Part{
					ID: fmt.Sprintf("%d", tr),
				}
				parts[tr] = p
			}

			var (
				clef *mxl.Clef
				key  *mxl.Key
			)

			if tr == constants.PercussionTrack {
				clef = &mxl.Clef{
					Sign: "percussion",
					Line: 2,
				}
			} else if i == 0 {
				// Set default CMaj key on each channel's first bar.
				key = &mxl.Key{
					Fifths: 0,
					Mode:   "major",
				}
			}

			measure := mxl.Measure{
				Number: i + 1,
				Atters: mxl.Attributes{
					Divisions: int(constants.TicksPerWhole) / int(bar.timeSig[1]),
					Key:       key,
					Clef:      clef,
				},
			}

			if barEvents, ok := events[tr]; ok {
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

						hasMultipleVoicesInPos := false
						{
							var prevVoice Voice
							for i, ev := range chords[pos] {
								if i == 0 {
									prevVoice = ev.Voice
								} else if ev.Voice != prevVoice {
									hasMultipleVoicesInPos = true
									break
								}
							}
						}

						prevNoteDur := 0
						noteCountInPos := 0

						for _, ev := range chords[pos] {
							if ev.Voice != voice {
								continue
							}

							if ev.Note == nil {
								// Meta event.
								if smfKey := (smf.Key{}); ev.Message.GetMetaKey(&smfKey) {
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
								} else if num, denom := uint8(0), uint8(0); ev.Message.GetMetaMeter(&num, &denom) {
									if num != timesig[0] || denom != timesig[1] {
										measure.Atters.Time = &mxl.Time{
											Beats:    int(num),
											BeatType: int(denom),
										}
										timesig[0] = num
										timesig[1] = denom
									}
								}
							} else {
								// Note event.

								var chord *xml.Name
								if noteCountInPos > 0 && !hasMultipleVoicesInPos {
									chord = &xml.Name{}
								}

								if noteCountInPos > 0 && prevNoteDur > 0 {
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
										// Type     string   `xml:"type"`
										Rest: &xml.Name{
											Local: "rest",
										},
										Chord: chord,
										// Tie      Tie      `xml:"tie"`
									})
								} else {
									var c, k, v uint8
									if !ev.Message.GetNoteStart(&c, &k, &v) {
										panic("expected GetNoteStart() to succeed")
									}

									note := mxl.Note{
										Duration: dur,
										Voice:    int(ev.Voice),
										Chord:    chord,
										NoteHead: &mxl.NoteHead{
											Filled:      "yes",
											Parentheses: "no",
											Value:       "normal",
										},
									}

									// TODO: not working
									if ev.Note.Props.NumGhost() > 0 {
										note.NoteHead.Parentheses = "yes"
									}

									switch tr {
									// case constants.PercussionTrack:
									// get the pitch and notehead

									default:
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

										note.Pitch = pitch
									}

									measure.Notes = append(measure.Notes, note)
								}

								noteCountInPos++
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
		type trackPart struct {
			part  *mxl.Part
			track uint8
		}

		tmpParts := make([]trackPart, 0, len(parts))
		for tr, p := range parts {
			tmpParts = append(tmpParts, trackPart{
				part:  p,
				track: tr,
			})
		}

		slices.SortFunc(tmpParts, func(a, b trackPart) int {
			return cmp.Compare(a.track, b.track)
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
		name := fmt.Sprintf("# %s", p.ID)

		score.PartList.Parts = append(score.PartList.Parts, mxl.ScorePart{
			ID:   p.ID,
			Name: name,
			ScoreInstrument: &mxl.ScoreInstrument{
				ID:   p.ID,
				Name: name,
			},
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
