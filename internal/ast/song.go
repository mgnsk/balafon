package ast

// Song is a single single source file.
type Song []interface{}

// NewSong creates a new song.
func NewSong(decl, inner interface{}) (song Song) {
	if innerList, ok := inner.(Song); ok {
		song = make(Song, len(innerList)+1)
		song[0] = decl
		copy(song[1:], innerList)
	} else {
		song = Song{decl}
	}

	return song
}
