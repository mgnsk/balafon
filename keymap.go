package balafon

type midiKey struct {
	channel Channel
	note    rune
}

// keyMap is a note keymap.
type keyMap struct {
	m map[midiKey]int
}

// newKeyMap creates a new keymap.
func newKeyMap() *keyMap {
	return &keyMap{
		m: map[midiKey]int{},
	}
}

// Range loops over the mapped keys.
func (m *keyMap) Range(f func(channel Channel, note rune, key int)) {
	for k, v := range m.m {
		f(k.channel, k.note, v)
	}
}

// Get a note key on channel.
func (m *keyMap) Get(channel Channel, note rune) (key int, exists bool) {
	key, ok := m.m[midiKey{channel, note}]
	return key, ok
}

// Set a note key on channel.
func (m *keyMap) Set(channel Channel, note rune, key int) (success bool) {
	if _, exists := m.m[midiKey{channel, note}]; exists {
		return false
	}
	m.m[midiKey{channel, note}] = key
	return true
}
