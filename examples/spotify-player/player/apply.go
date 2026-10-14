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
	a.page.SetChangedProperties([]string{"Tracks", "PageTracks", "Now", "Status", "ShelfTitle", "Playing", "TrackPage", "TrackPages", "TrackRange", "HasPrev", "HasNext"})
	if len(picks) > 0 {
		a.page.AddChangedProperty("Picks")
	}
	if len(shelf) > 0 {
		a.page.AddChangedProperties([]string{"Shelf", "Playlists"})
	}
}

// songsFrom keeps up to six song results.
func songsFrom(raw []searchResult) []searchResult {
	out := make([]searchResult, 0, 6)

	for _, s := range raw {
		if s.Kind != "song" || s.TrackName == "" {
			continue
		}

		out = append(out, s)

		if len(out) == 6 {
			break
		}
	}

	return out
}

// albumsFrom keeps up to twelve unique album results.
func albumsFrom(raw []searchResult) []searchResult {
	out := make([]searchResult, 0, 12)
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

		if len(out) == 12 {
			break
		}
	}

	return out
}
