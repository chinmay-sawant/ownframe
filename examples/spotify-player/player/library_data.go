package player

// LibCard is one library card. Kind 1 playlist, 2 album, 3 artist, 4 podcast.
type LibCard struct {
	Index, Kind int
	Title, Sub  string
	Cover       string
	Active      bool
}

// LibraryData holds the library filter pills and the media grid.
type LibraryData struct {
	Filter int
	All    []LibCard
	Cards  []LibCard
}

// defaultLibrary is the offline library screen.
func defaultLibrary() LibraryData {
	all := []LibCard{
		{0, 1, "Neon Nights", "Playlist · Chinmay", "cover-late-night", false},
		{1, 1, "Focus Flow", "Playlist · Chinmay", "cover-focus", false},
		{2, 1, "Late Drive", "Playlist · Spotify", "cover-late-night", false},
		{3, 1, "Rainy Day", "Playlist · Spotify", "cover-spring", false},
		{4, 2, "Neon Horizon", "Album · Nova Waves", "cover-late-night", false},
		{5, 2, "Static Bloom", "Album · The Far Coast", "cover-spring", false},
		{6, 2, "Signal Fire", "Album · Aster Field", "cover-ambient", false},
		{7, 3, "Nova Waves", "Artist", "cover-funk", false},
		{8, 3, "Aster Field", "Artist", "cover-jazz", false},
		{9, 4, "Signal Path", "Podcast · Weekly", "cover-jazz", false},
	}
	return LibraryData{All: all, Cards: append([]LibCard(nil), all...)}
}
