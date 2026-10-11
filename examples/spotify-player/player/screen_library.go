package player

import (
	"context"
	"strings"
)

// LibraryFilter rebuilds Cards for the current Filter.
func (a *App) LibraryFilter() {
	lib := &a.view.Library
	lib.Cards = lib.Cards[:0]
	for _, c := range lib.All {
		if lib.Filter == 0 || c.Kind == lib.Filter {
			c.Index = len(lib.Cards)
			c.Active = false
			lib.Cards = append(lib.Cards, c)
		}
	}
}

// clickLibrary applies one action on the library screen.
func (a *App) clickLibrary(ctx context.Context, action string) error {
	switch {
	case strings.HasPrefix(action, "library-filter-"):
		f := slot(action, "library-filter-")
		if f < 0 || f > 4 {
			return nil
		}

		a.view.Library.Filter = f
		a.LibraryFilter()
	case strings.HasPrefix(action, "library-card-"):
		cards := a.view.Library.Cards
		i := slot(action, "library-card-")
		for j := range cards {
			cards[j].Active = j == i
		}
	}

	return nil
}
