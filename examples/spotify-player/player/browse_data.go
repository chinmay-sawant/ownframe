package player

// BrowseHero is one wide card in the featured band.
type BrowseHero struct {
	Title, Sub, Img string
	Active          bool
}

// BrowseTile is one square genre tile.
type BrowseTile struct {
	Title  string
	Tone   int
	Img    string
	Active bool
}

// BrowseData is the data behind the browse screen: the featured band and
// the genre grid.
type BrowseData struct {
	Hero  []BrowseHero
	Tiles []BrowseTile
}

// defaultBrowse returns the offline browse screen.
func defaultBrowse() BrowseData {
	titles := []string{
		"Chill", "Focus", "Workout", "Party", "Sleep", "Chillhop",
		"Indie", "Pop", "Rock", "Jazz", "Classical", "Electronic",
	}
	hero := []BrowseHero{
		{Title: "Midnight Drive", Sub: "Nova Waves · Album", Img: "cover-late-night"},
		{Title: "Neon Skyline", Sub: "Lumen Fields · Album", Img: "cover-ambient"},
		{Title: "Slow Motion", Sub: "Velvet Static · Album", Img: "cover-soul"},
	}
	covers := []string{
		"cover-chillhop", "cover-focus", "cover-funk", "cover-hybrid",
		"cover-ambient", "cover-chillhop", "cover-spring", "cover-soul",
		"cover-funk", "cover-jazz", "cover-asian-zen", "cover-late-night",
	}
	tiles := make([]BrowseTile, len(titles))
	for i, title := range titles {
		tiles[i] = BrowseTile{Title: title, Tone: i, Img: covers[i]}
	}
	return BrowseData{Hero: hero, Tiles: tiles}
}
