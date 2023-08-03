package balafon_test

import "github.com/mgnsk/balafon"

func FindNote(bar *balafon.Bar) (ch, key, velocity uint8, ok bool) {
	for _, ev := range bar.Events {
		if ev.Message.GetNoteOn(&ch, &key, &velocity) {
			return ch, key, velocity, true
		}
	}

	return ch, key, velocity, false
}
