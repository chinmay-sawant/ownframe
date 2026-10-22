package player

// sampleShelf is the six-card Recently played shelf.
func sampleShelf() []Card {
	tracks := sampleTracks()
	shelf := make([]Card, 0, maxSongs)
	for i, track := range tracks[:6] {
		shelf = append(shelf, Card{
			Index: i, Title: track.Title, Sub: track.Artist, Cover: track.Cover,
		})
	}
	return shelf
}

// samplePlaylists is the five-row sidebar library.
func samplePlaylists() []Playlist {
	return []Playlist{
		{Index: 0, Name: "Liked Songs", Meta: "Playlist · 214 songs", Cover: "cover-soul"},
		{Index: 1, Name: "Neon Nights", Meta: "Playlist · Chinmay", Cover: "cover-late-night"},
		{Index: sampleIdx2, Name: "Focus Flow", Meta: "Playlist · Chinmay", Cover: "cover-focus"},
		{Index: sampleIdx3, Name: "Late Drive", Meta: "Playlist · Spotify", Cover: "cover-late-night"},
		{Index: sampleIdx4, Name: "Rainy Day", Meta: "Playlist · Spotify", Cover: "cover-spring"},
	}
}
