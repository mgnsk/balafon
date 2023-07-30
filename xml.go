package balafon

import (
	"github.com/eliothedeman/mxl"
)

// ToXML converts a balafon script to MusicXML.
func ToXML(input []byte) ([]byte, error) {
	it := New()

	if err := it.Eval(input); err != nil {
		return nil, err
	}

	bars := it.Flush()

	_ = bars
	// spew.Dump(bars)

	// seq := NewSequencer()
	// seq.AddBars(bars...)

	// events := seq.Flush()

	doc := mxl.MXLDoc{}
	_ = doc
	return nil, nil
}
