package player

import (
	"context"
	"strings"
)

// SearchTile is one Browse all category tile.
type SearchTile struct {
	Index  int
	Title  string
	Tone   int
	Img    string
	Active bool
}

// SearchData is the data behind the search screen: the Browse all tiles and
// the recent searches.
type SearchData struct {
	Tiles  []SearchTile
	Recent []Track
}

// defaultSearch returns the offline search screen.
func defaultSearch() SearchData {
	titles := []string{"Music", "Podcasts", "Live Events", "Made For You", "New Releases", "Charts", "Pop", "Hip-Hop", "Rock", "Indie", "Jazz", "Mood"}

	tiles := make([]SearchTile, 0, len(titles))
	covers := []string{
		"cover-hybrid", "cover-jazz", "cover-late-night", "cover-focus",
		"cover-spring", "cover-funk", "cover-soul", "cover-chillhop",
		"cover-funk", "cover-spring", "cover-jazz", "cover-ambient",
	}
	for i, title := range titles {
		tiles = append(tiles, SearchTile{Index: i, Title: title, Tone: i, Img: covers[i]})
	}

	recent := sampleTracks()[:5]

	return SearchData{Tiles: tiles, Recent: recent}
}

// clickSearch applies one action on the search screen.
func (a *App) clickSearch(ctx context.Context, action string) error {
	switch {
	case strings.HasPrefix(action, "search-tile-"):
		a.searchSelectTile(slot(action, "search-tile-"))
	case strings.HasPrefix(action, "search-recent-"):
		a.searchPlayRecent(ctx, slot(action, "search-recent-"))
	case strings.HasPrefix(action, "search-song-"):
		a.selectRow(slot(action, "search-song-"))
		a.playNow(ctx)
	}

	return nil
}

// searchSelectTile highlights tile i and clears the other tiles.
func (a *App) searchSelectTile(i int) {
	for j := range a.view.Search.Tiles {
		a.view.Search.Tiles[j].Active = j == i
	}
}

// searchPlayRecent makes recent search i the now-playing item.
func (a *App) searchPlayRecent(ctx context.Context, i int) {
	if i < 0 || i >= len(a.view.Search.Recent) {
		return
	}

	a.view.Now = a.view.Search.Recent[i]
	a.resetProgress()
	a.playNow(ctx)
}
