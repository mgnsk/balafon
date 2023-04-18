package interpreter

type midiKey struct {
	channel uint8
	note    rune
}

// KeyMap is a note keymap.
type KeyMap struct {
	m map[midiKey]uint8
}

// NewKeyMap creates a new keymap.
func NewKeyMap() *KeyMap {
	return &KeyMap{
		m: map[midiKey]uint8{},
	}
}

// Get a note key on channel.
func (m *KeyMap) Get(channel uint8, note rune) (key uint8, exists bool) {
	key, ok := m.m[midiKey{channel, note}]
	return key, ok
}

// Set a note key on channel.
func (m *KeyMap) Set(channel uint8, note rune, key uint8) (success bool) {
	if _, exists := m.m[midiKey{channel, note}]; exists {
		return false
	}
	m.m[midiKey{channel, note}] = key
	return true
}
