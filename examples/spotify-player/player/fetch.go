package player

import (
	"context"
	"fmt"
	"strings"
)

// itunesBase is the live iTunes Search API host.
const itunesBase = "https://itunes.apple.com"

// DefaultTerm is the initial term shown in the search field.
const DefaultTerm = "daft punk"

// searchResult is the fields the player reads from one iTunes result.
type searchResult struct {
	Kind           string `json:"kind"`
	TrackName      string `json:"trackName"`
	ArtistName     string `json:"artistName"`
	CollectionName string `json:"collectionName"`
	Genre          string `json:"primaryGenreName"`
	TrackMillis    int    `json:"trackTimeMillis"`
	Artwork        string `json:"artworkUrl100"`
	ReleaseDate    string `json:"releaseDate"`
}

// searchResponse is the iTunes envelope.
type searchResponse struct {
	Results []searchResult `json:"results"`
}

// Load replaces the sample data with iTunes search results for term.
// Any failure keeps the current content and reports Sample data (offline).
func (a *App) Load(ctx context.Context, term string) error {
	term = strings.TrimSpace(term)
	if term == "" {
		return a.offline(fmt.Errorf("player: empty search term"))
	}

	rawSongs, err := a.search(ctx, term, "song", maxSongs)
	if err != nil {
		return a.offline(err)
	}

	rawAlbums, err := a.search(ctx, term, "album", maxAlbums)
	if err != nil {
		return a.offline(err)
	}

	songs := songsFrom(rawSongs)
	albums := albumsFrom(rawAlbums)
	picks, shelf := cardsFrom(albums)

	a.apply(term, tracksFrom(songs), picks, shelf)

	for name, body := range a.fetchCovers(ctx, coverJobs(songs, albums)) {
		a.page.SetImage(name, body)
	}

	return nil
}

// offline marks the view offline and returns err unchanged.
func (a *App) offline(err error) error {
	a.view.Status = sampleStatus
	a.ensureCredit()
	a.page.SetData(&a.view)
	a.page.SetChangedProperties([]string{"Status", "Credit"})

	return err
}
