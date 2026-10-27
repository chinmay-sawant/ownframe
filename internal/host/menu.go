package host

import "context"

// MenuItem is one row of the window's context menu. ID names the action:
// cut, copy, paste, select-all, undo, or redo, or a custom id the screen
// handles through MenuActioner.
type MenuItem struct {
	ID      string
	Label   string
	Enabled bool
}

// ContextMenu is a screen that reports the context menu rows for its current
// state.
type ContextMenu interface {
	ContextMenu() []MenuItem
}

// ContextMenuFor is a screen that reports rows for one hit-tested element.
// The window prefers it over ContextMenu when the screen implements it; the
// id is the innermost box under the pointer, or "" when no box was hit.
type ContextMenuFor interface {
	ContextMenuFor(id string) []MenuItem
}

// MenuActioner is a screen that runs the custom rows it reported. The window
// calls it for an id that is not one of the built-in cut, copy, paste,
// select-all, undo, or redo rows.
type MenuActioner interface {
	MenuAction(ctx context.Context, id string) error
}
