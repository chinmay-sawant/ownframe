package player

// apply swaps the sample content for converted search results.
func (a *App) apply(term string, tracks []Track, picks, shelf []Card) {
	if len(tracks) == 0 {
		tracks = a.view.Tracks
	}

	if len(picks) > 0 {
		a.view.Picks = picks
	}

	if len(shelf) > 0 {
		a.view.Shelf = shelf
		a.view.Playlists = playlistsFrom(shelf)
	}

	a.view.Tracks = tracks
	paginateTracks(&a.view, 0)
	a.view.Now = tracks[0]
	a.view.Status = "Live from iTunes · " + term
	a.view.ShelfTitle = "Top results for " + term
	a.view.Playing = true
	a.ensureCredit()
	a.resetProgress()
	a.page.SetData(&a.view)
	a.page.Probe().SetChangedProperties([]string{"Tracks", "PageTracks", "Now", "Status", "ShelfTitle", "Playing", "TrackPage", "TrackPages", "TrackRange", "HasPrev", "HasNext"})
	if len(picks) > 0 {
		a.page.Probe().AddChangedProperty("Picks")
	}
	if len(shelf) > 0 {
		a.page.Probe().AddChangedProperties([]string{"Shelf", "Playlists"})
	}
}

// songsFrom keeps up to six song results.
func songsFrom(raw []searchResult) []searchResult {
	out := make([]searchResult, 0, maxSongs)

	for _, s := range raw {
		if s.Kind != "song" || s.TrackName == "" {
			continue
		}

		out = append(out, s)

		if len(out) == maxSongs {
			break
		}
	}

	return out
}

// albumsFrom keeps up to twelve unique album results.
func albumsFrom(raw []searchResult) []searchResult {
	out := make([]searchResult, 0, maxAlbums)
	seen := map[string]bool{}

	for _, al := range raw {
		if al.CollectionName == "" {
			continue
		}

		key := al.CollectionName + "|" + al.ArtistName
		if seen[key] {
			continue
		}

		seen[key] = true
		out = append(out, al)

		if len(out) == maxAlbums {
			break
		}
	}

	return out
}
