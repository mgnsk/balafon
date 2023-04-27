package interpreter

type midiKey struct {
	channel uint8
	note    rune
}

// KeyMap is a note keymap.
type KeyMap struct {
	m map[midiKey]int
}

// NewKeyMap creates a new keymap.
func NewKeyMap() *KeyMap {
	return &KeyMap{
		m: map[midiKey]int{},
	}
}

// Get a note key on channel.
func (m *KeyMap) Get(channel uint8, note rune) (key int, exists bool) {
	key, ok := m.m[midiKey{channel, note}]
	return key, ok
}

// Set a note key on channel.
func (m *KeyMap) Set(channel uint8, note rune, key int) (success bool) {
	if _, exists := m.m[midiKey{channel, note}]; exists {
		return false
	}
	m.m[midiKey{channel, note}] = key
	return true
}
