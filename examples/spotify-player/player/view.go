package player

// Track is one row of the tracklist and the now-playing item.
// Index is the zero-based slot used in row ids and actions.
type Track struct {
	Index         int
	Num           string
	Title, Artist string
	Description   string
	Album, Year   string
	Genre         string
	Length        string
	Cover         string // registered static artwork for this track's category
	Active        bool
}

// Card is one greeting tile or shelf card.
type Card struct {
	Index      int
	Title, Sub string
	Cover      string // registered local cover or a search result slot
	Active     bool
}

// Playlist is one row of the sidebar library.
type Playlist struct {
	Index      int
	Name, Meta string
	Cover      string
	Active     bool
}

// Profile is the signed-in user the avatar opens.
type Profile struct {
	Name                            string
	Playlists, Followers, Following int
}

// View is the data the player template prints.
type View struct {
	Nav        string // "home" | "search" | "library" | "profile"
	CanBack    bool
	CanForward bool
	Playing    bool
	Shuffle    bool
	Repeat     bool
	Liked      bool
	Volume     int
	Progress   int
	Elapsed    string
	Remaining  string
	Now        Track
	Tracks     []Track
	PageTracks []Track
	TrackPage  int
	TrackPages int
	TrackRange string
	HasPrev    bool
	HasNext    bool
	Picks      []Card
	Shelf      []Card
	Playlists  []Playlist
	Status     string
	Query      string
	ShelfTitle string
	Credit     string
	Profile    Profile
	Search     SearchData
	Library    LibraryData
	LikedSongs LikedData
	Browse     BrowseData
	Radio      RadioData
	Queue      QueueData
}
