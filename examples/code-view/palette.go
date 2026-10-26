package main

import (
	"context"
	"strings"
)

const (
	paletteDefaultRows = 5
	paletteMaxRows     = 8
)

// paletteView is the quick-open overlay data.
type paletteView struct {
	Open bool
	Rows []paletteRow
}

// paletteRow is one file suggestion.
type paletteRow struct {
	Path   string
	Name   string
	Dir    string
	Lang   string
	Badge  string
	Active bool
}

// paletteRows returns the open tabs followed by workspace files for an empty
// query, or path matches for a query. The list never leaves the open folder.
func (a *App) paletteRows() []paletteRow {
	query := strings.ToLower(strings.TrimSpace(a.paletteQuery))

	if query == "" {
		rows := make([]paletteRow, 0, paletteDefaultRows)
		seen := map[string]bool{}

		for _, path := range a.tabs {
			if seen[path] || len(rows) >= paletteDefaultRows {
				continue
			}

			if file := a.ws.ByPath[path]; file != nil {
				rows = append(rows, paletteRowOf(file))
				seen[path] = true
			}
		}

		for _, fresh := range a.recentRows() {
			if len(rows) >= paletteDefaultRows {
				break
			}

			if seen[fresh.Path] {
				continue
			}

			if file := a.ws.ByPath[fresh.Path]; file != nil {
				rows = append(rows, paletteRowOf(file))
				seen[fresh.Path] = true
			}
		}

		for _, file := range a.ws.Files {
			if len(rows) >= paletteDefaultRows {
				break
			}

			if seen[file.Path] {
				continue
			}

			rows = append(rows, paletteRowOf(file))
		}

		return rows
	}

	prefix := make([]paletteRow, 0, paletteMaxRows)
	contains := make([]paletteRow, 0, paletteMaxRows)

	for _, file := range a.ws.Files {
		path := strings.ToLower(file.Path)

		switch {
		case strings.HasPrefix(path, query):
			prefix = append(prefix, paletteRowOf(file))
		case strings.Contains(path, query):
			contains = append(contains, paletteRowOf(file))
		}
	}

	rows := make([]paletteRow, 0, len(prefix)+len(contains))
	rows = append(rows, prefix...)
	rows = append(rows, contains...)

	if len(rows) > paletteMaxRows {
		rows = rows[:paletteMaxRows]
	}

	return rows
}

// paletteRowOf builds one suggestion.
func paletteRowOf(file *fileMeta) paletteRow {
	return paletteRow{
		Path:  file.Path,
		Name:  file.Name,
		Dir:   file.Dir,
		Lang:  file.Lang,
		Badge: languageInfo(file.Lang).Badge,
	}
}

// paletteView marks the selected row.
func (a *App) paletteView() paletteView {
	out := paletteView{Open: a.paletteOpen}
	if !a.paletteOpen {
		return out
	}

	rows := a.paletteRows()

	switch {
	case len(rows) == 0:
		a.paletteIndex = 0
	case a.paletteIndex >= len(rows):
		a.paletteIndex = len(rows) - 1
	case a.paletteIndex < 0:
		a.paletteIndex = 0
	}

	for i := range rows {
		rows[i].Active = i == a.paletteIndex
	}

	out.Rows = rows

	return out
}

// openPalette shows the overlay and focuses its input.
func (a *App) openPalette(ctx context.Context) error {
	a.paletteOpen = true
	a.paletteQuery = ""
	a.paletteIndex = 0
	a.refresh()

	if err := a.page.Redraw(ctx); err != nil {
		return err
	}

	return a.page.Form().Focus(ctx, "palette")
}

// closePalette hides the overlay.
func (a *App) closePalette(ctx context.Context) error {
	a.paletteOpen = false
	a.paletteQuery = ""
	a.paletteIndex = 0
	a.refresh()

	return a.page.Form().Focus(ctx, "")
}

// movePalette moves the selected suggestion.
func (a *App) movePalette(delta int) {
	rows := a.paletteRows()
	if len(rows) == 0 {
		return
	}

	a.paletteIndex += delta
	if a.paletteIndex < 0 {
		a.paletteIndex = 0
	}

	if a.paletteIndex >= len(rows) {
		a.paletteIndex = len(rows) - 1
	}

	a.refresh()
}
