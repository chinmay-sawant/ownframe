package player

import (
	"strconv"
	"strings"
)

// selectRow makes track i the now-playing item.
func (a *App) selectRow(i int) {
	if i < 0 || i >= len(a.view.Tracks) {
		return
	}

	// Mutate the owned track slice in place: drop the previous active mark
	// instead of rebuilding every row's Active flag.
	for j := range a.view.Tracks {
		if j != i && a.view.Tracks[j].Active {
			a.view.Tracks[j].Active = false
		}
	}
	a.view.Tracks[i].Active = true

	a.view.Now = a.view.Tracks[i]
	paginateTracks(&a.view, i/trackPageSize)
	a.resetProgress()
}

// step moves the active track by delta with wrap-around.
func (a *App) step(delta int) {
	n := len(a.view.Tracks)
	if n == 0 {
		return
	}

	cur := 0
	for i := range a.view.Tracks {
		if a.view.Tracks[i].Active {
			cur = i
		}
	}

	a.selectRow((cur + delta + n) % n)
}

// selectPick plays the greeting tile and clears the shelf highlight.
func (a *App) selectPick(i int) {
	if i < 0 || i >= len(a.view.Picks) {
		return
	}

	for j := range a.view.Picks {
		a.view.Picks[j].Active = j == i
	}

	for j := range a.view.Shelf {
		a.view.Shelf[j].Active = false
	}

	a.setNow(a.view.Picks[i])
}

// selectShelf plays the shelf card and clears the greeting highlight.
func (a *App) selectShelf(i int) {
	if i < 0 || i >= len(a.view.Shelf) {
		return
	}

	for j := range a.view.Shelf {
		a.view.Shelf[j].Active = j == i
	}

	for j := range a.view.Picks {
		a.view.Picks[j].Active = false
	}

	a.setNow(a.view.Shelf[i])
}

// selectList highlights the sidebar playlist and opens the library nav, or
// the Liked Songs screen for the pinned first entry.
func (a *App) selectList(i int) {
	if i < 0 || i >= len(a.view.Playlists) {
		return
	}

	for j := range a.view.Playlists {
		a.view.Playlists[j].Active = j == i
	}

	if strings.EqualFold(a.view.Playlists[i].Name, "Liked Songs") {
		a.navigate("liked")
		return
	}

	a.navigate("library")
}

// slot parses the trailing number of an action like "row-3".
func slot(action, prefix string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(action, prefix))
	if err != nil {
		return -1
	}

	return n
}
