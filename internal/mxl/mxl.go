package mxl

import "encoding/xml"

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
	Number int           `xml:"number,attr"`
	Atters Attributes    `xml:"attributes"`
	Notes  []interface{} // Note or Backup
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
	XMLName  xml.Name  `xml:"note"`
	Pitch    *Pitch    `xml:"pitch,omitempty"`
	Duration int       `xml:"duration"`
	Voice    int       `xml:"voice,omitempty"`
	Type     string    `xml:"type,omitempty"`
	Rest     *xml.Name `xml:"rest,omitempty"`
	Chord    *xml.Name `xml:"chord,omitempty"`
	Tie      *Tie      `xml:"tie,omitempty"`
}

// Backup represents the backup element.
type Backup struct {
	XMLName  xml.Name `xml:"backup"`
	Duration int      `xml:"duration"`
}

// Pitch represents the pitch of a note
type Pitch struct {
	Accidental int8   `xml:"alter"`
	Step       string `xml:"step"`
	Octave     int    `xml:"octave"`
}

// Tie represents whether or not a note is tied.
type Tie struct {
	Type string `xml:"type,attr"`
}
