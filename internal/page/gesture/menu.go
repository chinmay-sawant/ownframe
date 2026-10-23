package gesture

import "github.com/chinmay-sawant/ownframe/internal/host"

// Menu is the page state the context menu reads. The page decides each flag.
type Menu struct {
	Selection bool
	Editable  bool
	SelectAll bool
	Undo      bool
	Redo      bool
}

// ContextMenu returns the context-menu rows. Cut and copy need a selection,
// paste needs a focused editable field, select all needs either, and undo
// and redo need their handler.
func ContextMenu(m Menu) []host.MenuItem {
	return []host.MenuItem{
		{ID: "cut", Label: "Cut", Enabled: m.Selection},
		{ID: "copy", Label: "Copy", Enabled: m.Selection},
		{ID: "paste", Label: "Paste", Enabled: m.Editable},
		{ID: "select-all", Label: "Select all", Enabled: m.SelectAll},
		{ID: "undo", Label: "Undo", Enabled: m.Undo},
		{ID: "redo", Label: "Redo", Enabled: m.Redo},
	}
}
