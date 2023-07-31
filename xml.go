package balafon

import (
	"bytes"
	"encoding/xml"
	"strconv"

	"github.com/mgnsk/balafon/internal/constants"
	"golang.org/x/exp/slices"
)

var notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

// GetPitch returns the step and octave for MIDI note.
func GetPitch(note uint8) (step string, octave uint8) {
	octave = (note / 12) - 1
	noteIndex := note % 12

	return notes[noteIndex], octave
}

type xmlTrack struct {
	channel int
	bars    []Bar
}

// ToXML converts a balafon script to MusicXML.
func ToXML(input []byte) ([]byte, error) {
	it := New()

	if err := it.Eval(input); err != nil {
		return nil, err
	}

	bars := it.Flush()

	channels := map[uint8]*xmlTrack{}
	for _, bar := range bars {
		for _, ev := range bar.Events {
			t, ok := channels[ev.Track]
			if !ok {
				t = &xmlTrack{}
				channels[ev.Track] = t
			}
		}
	}

	parts := map[uint8]*Part{}

	for i, bar := range bars {
		events := map[uint8][]Event{}

		for _, ev := range bar.Events {
			events[ev.Track] = append(events[ev.Track], ev)
		}

		for ch := range channels {
			p, ok := parts[ch]
			if !ok {
				p = &Part{
					Id: strconv.Itoa(int(ch)),
				}
				parts[ch] = p
			}

			measure := Measure{
				Number: i,
				Atters: Attributes{
					Time: &Time{
						Beats:    int(bar.TimeSig[0]),
						BeatType: int(bar.TimeSig[1]),
					},
					Divisions: int(constants.TicksPerWhole),
					// Key       Key  `xml:"key"`
					// Time      Time `xml:"time"`
					// Divisions int  `xml:"divisions"`
					// Clef      Clef `xml:"clef"`

				},
			}

			if barEvents, ok := events[ch]; ok {
				for _, ev := range barEvents {
					if ev.Note != nil {
						if ev.Note.IsPause() {
							measure.Notes = append(measure.Notes, Note{
								// Pitch: mxl.Pitch{
								// 	// Accidental int8   `xml:"alter"`
								// 	Step:   "C",
								// 	Octave: 4,
								// },
								Duration: int(ev.Note.Props.NoteLen()),
								// Voice    int      `xml:"voice"`
								// Type     string   `xml:"type"`
								Rest: &xml.Name{
									Local: "rest",
								},
								// Chord    xml.Name `xml:"chord"`
								// Tie      Tie      `xml:"tie"`

							})
						} else {
							var c, k, v uint8
							if !ev.Message.GetNoteStart(&c, &k, &v) {
								panic("expected GetNoteStart() to succeed")
							}

							step, octave := GetPitch(k)

							measure.Notes = append(measure.Notes, Note{
								Pitch: Pitch{
									// Accidental int8   `xml:"alter"`
									Step:   step,
									Octave: int(octave),
								},
								Duration: int(ev.Note.Props.NoteLen()),
								// Voice    int      `xml:"voice"`
								// Type     string   `xml:"type"`
								// Chord    xml.Name `xml:"chord"`
								// Tie      Tie      `xml:"tie"`
							})
						}
					}
				}
			}

			p.Measures = append(p.Measures, measure)
		}
	}

	type trackPart struct {
		ch   uint8
		part *Part
	}
	tps := make([]trackPart, 0, len(parts))
	for ch, p := range parts {
		tps = append(tps, trackPart{
			ch:   ch,
			part: p,
		})
	}
	slices.SortFunc(tps, func(a, b trackPart) bool {
		return a.ch < b.ch
	})

	finalParts := make([]Part, len(tps))
	for i := 0; i < len(tps); i++ {
		finalParts[i] = *tps[i].part
	}

	score := Score{
		Version: "4.0",
		Parts:   finalParts,
	}

	for _, p := range parts {
		score.PartList.Parts = append(score.PartList.Parts, ScorePart{
			ID:   p.Id,
			Name: p.Id,
		})
	}

	var buf bytes.Buffer

	buf.WriteString(`<?xml version="1.0" encoding="UTF-8" standalone="no"?>
	<!DOCTYPE score-partwise PUBLIC
	"-//Recordare//DTD MusicXML 4.0 Partwise//EN"
	"http://www.musicxml.org/dtds/partwise.dtd">
	`)

	enc := xml.NewEncoder(&buf)
	enc.Indent("", "    ")
	if err := enc.Encode(score); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// Score holds all data for a music xml file
type Score struct {
	XMLName        xml.Name        `xml:"score-partwise"`
	Version        string          `xml:"version,attr"`
	Identification *Identification `xml:"identification,omitempty"`
	PartList       PartList        `xml:"part-list,omitempty"`
	Parts          []Part          `xml:"part,omitempty"`
}

type PartList struct {
	Parts []ScorePart `xml:"score-part,omitempty"`
}

type ScorePart struct {
	ID   string `xml:"id,attr"`
	Name string `xml:"part-name"`
}

// Identification holds all of the ident information for a music xml file
type Identification struct {
	Composer string    `xml:"creator,omitempty"`
	Encoding *Encoding `xml:"encoding,omitempty"`
	Rights   string    `xml:"rights,omitempty"`
	Source   string    `xml:"source,omitempty"`
	Title    string    `xml:"movement-title,omitempty"`
}

// Encoding holds encoding info
type Encoding struct {
	Software string `xml:"software,omitempty"`
	Date     string `xml:"encoding-date,omitempty"`
}

// Part represents a part in a piece of music
type Part struct {
	Id       string    `xml:"id,attr"`
	Measures []Measure `xml:"measure"`
}

// Measure represents a measure in a piece of music
type Measure struct {
	Number int        `xml:"number,attr"`
	Atters Attributes `xml:"attributes"`
	Notes  []Note     `xml:"note"`
}

// Attributes represents
type Attributes struct {
	Key       *Key  `xml:"key,omitempty"`
	Time      *Time `xml:"time,omitempty"`
	Divisions int   `xml:"divisions,omitempty"`
	Clef      *Clef `xml:"clef,omitempty"`
}

// Clef represents a clef change
type Clef struct {
	Sign string `xml:"sign"`
	Line int    `xml:"line"`
}

// Key represents a key signature change
type Key struct {
	Fifths int    `xml:"fifths"`
	Mode   string `xml:"mode"`
}

// Time represents a time signature change
type Time struct {
	Beats    int `xml:"beats"`
	BeatType int `xml:"beat-type"`
}

// Note represents a note in a measure
type Note struct {
	Pitch    Pitch     `xml:"pitch"`
	Duration int       `xml:"duration"`
	Voice    int       `xml:"voice,omitempty"`
	Type     string    `xml:"type,omitempty"`
	Rest     *xml.Name `xml:"rest,omitempty"`
	Chord    *xml.Name `xml:"chord,omitempty"`
	Tie      *Tie      `xml:"tie,omitempty"`
}

// Pitch represents the pitch of a note
type Pitch struct {
	Accidental int8   `xml:"alter,omitempty"`
	Step       string `xml:"step"`
	Octave     int    `xml:"octave"`
}

// Tie represents whether or not a note is tied.
type Tie struct {
	Type string `xml:"type,attr"`
}
