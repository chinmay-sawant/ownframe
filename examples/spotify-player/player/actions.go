package player

import (
	"strconv"
	"strings"
)

// Demo boundary: this player is a single-user local demo with no
// authentication or authorization; every rendered action runs as the
// viewer. Add server-side permission checks before adding accounts,
// tenants, or privileged operations.

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

// slot parses the trailing number of an action like "row-3".
func slot(action, prefix string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(action, prefix))
	if err != nil {
		return -1
	}

	return n
}
