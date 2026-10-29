package main

import "github.com/chinmay-sawant/ownframe"

// contextMenuForPage converts the app's local menu rows to the public row
// type the page bridge carries.
func (a *App) contextMenuForPage(id string) []ownframe.MenuItem {
	rows := a.ContextMenuFor(id)
	if len(rows) == 0 {
		return nil
	}

	out := make([]ownframe.MenuItem, 0, len(rows))

	for _, row := range rows {
		out = append(out, ownframe.MenuItem{
			ID:      row.ID,
			Label:   row.Label,
			Hint:    row.Hint,
			Enabled: row.Enabled,
		})
	}

	return out
}
