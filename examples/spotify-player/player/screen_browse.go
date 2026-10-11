package player

import (
	"context"
	"strings"
)

// BrowseSelectHero highlights one featured card.
func (a *App) BrowseSelectHero(i int) {
	if i < 0 || i >= len(a.view.Browse.Hero) {
		return
	}

	for j := range a.view.Browse.Hero {
		a.view.Browse.Hero[j].Active = j == i
	}
}

// BrowseSelectTile highlights one genre tile.
func (a *App) BrowseSelectTile(i int) {
	if i < 0 || i >= len(a.view.Browse.Tiles) {
		return
	}

	for j := range a.view.Browse.Tiles {
		a.view.Browse.Tiles[j].Active = j == i
	}
}

// clickBrowse applies one action on the browse screen.
func (a *App) clickBrowse(ctx context.Context, action string) error {
	switch {
	case strings.HasPrefix(action, "browse-hero-"):
		a.BrowseSelectHero(slot(action, "browse-hero-"))
	case strings.HasPrefix(action, "browse-tile-"):
		a.BrowseSelectTile(slot(action, "browse-tile-"))
	}

	return nil
}
