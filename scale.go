package balafon

import (
	"fmt"

	"gitlab.com/gomidi/midi/v2/smf"
)

var notes = []string{"C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"}

func getPitch(note int) (step string, octave int) {
	octave = (note / 12) - 1
	noteIndex := note % 12

	return notes[noteIndex], octave
}

var (
	sharps = []string{"F", "C", "G", "D", "A", "E"}
	flats  = []string{"B", "E", "A", "D", "G", "C"}
)

func getScale(scale string) (func() smf.Message, []string, []string) {
	switch scale {
	// Major scales.
	case "C", "CMaj":
		return smf.CMaj, nil, nil
	case "G", "GMaj":
		return smf.GMaj, sharps[:1], nil
	case "D", "DMaj":
		return smf.DMaj, sharps[:2], nil
	case "A", "AMaj":
		return smf.AMaj, sharps[:3], nil
	case "E", "EMaj":
		return smf.EMaj, sharps[:4], nil
	case "B", "BMaj":
		return smf.BMaj, sharps[:5], nil
	case "F#", "FsharpMaj":
		return smf.FsharpMaj, sharps[:6], nil

	case "F", "FMaj":
		return smf.FMaj, nil, flats[:1]
	case "Bb", "BbMaj":
		return smf.BbMaj, nil, flats[:2]
	case "Eb", "EbMaj":
		return smf.EbMaj, nil, flats[:3]
	case "Ab", "AbMaj":
		return smf.AbMaj, nil, flats[:4]
	case "Db", "DbMaj":
		return smf.DbMaj, nil, flats[:5]
	case "Gb", "GbMaj":
		return smf.GbMaj, nil, flats[:6]

	// Minor scales.
	case "Am", "AMin":
		return smf.AMin, nil, nil
	case "Em", "EMin":
		return smf.EMin, sharps[:1], nil
	case "Bm", "BMin":
		return smf.BMin, sharps[:2], nil
	case "F#m", "FsharpMin":
		return smf.FsharpMin, sharps[:3], nil
	case "C#m", "CsharpMin":
		return smf.CsharpMin, sharps[:4], nil
	case "G#m", "GsharpMin":
		return smf.GsharpMin, sharps[:5], nil
	case "D#m", "DsharpMin":
		return smf.DsharpMin, sharps[:6], nil

	case "Dm", "DMin":
		return smf.DMin, nil, flats[:1]
	case "Gm", "GMin":
		return smf.GMin, nil, flats[:2]
	case "Cm", "CMin":
		return smf.CMin, nil, flats[:3]
	case "Fm", "FMin":
		return smf.FMin, nil, flats[:4]
	case "Bbm", "BbMin":
		return smf.BbMin, nil, flats[:5]
	case "Ebm", "EbMin":
		return smf.EbMin, nil, flats[:6]

	default:
		panic(fmt.Sprintf("invalid scale %q", scale))
	}
}
