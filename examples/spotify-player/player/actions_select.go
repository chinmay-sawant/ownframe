package player

import (
	"context"
	"strings"
)

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
func (a *App) selectList(ctx context.Context, i int) {
	if i < 0 || i >= len(a.view.Playlists) {
		return
	}

	for j := range a.view.Playlists {
		a.view.Playlists[j].Active = j == i
	}

	if strings.EqualFold(a.view.Playlists[i].Name, "Liked Songs") {
		a.navigate(ctx, "liked")
		return
	}

	a.navigate(ctx, "library")
}
