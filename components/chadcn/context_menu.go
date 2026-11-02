package chadcn

import (
	"context"

	"github.com/chinmay-sawant/ownframe/ui"
)

func contextMenuSkin() menuSkin {
	return menuSkin{
		surface:   "of-context-menu",
		item:      "of-context-menu-item",
		danger:    "of-context-menu-item--danger",
		shortcut:  "of-context-menu-shortcut",
		label:     "of-context-menu-label",
		separator: "of-context-menu-separator",
		chevron:   "of-context-menu-chevron",
		row:       "of-context-menu-row",
		submenu:   "of-context-menu-submenu",
	}
}

// ContextMenuOptions describes a controlled context menu. Open comes from the
// host after it reads a right-click in Handlers and records the pointer
// position in X and Y. ID names the menu. Items are flat rows; Groups add
// labeled sections after them. OnSelect receives the activated item label.
type ContextMenuOptions struct {
	ID       string
	Label    string
	Open     bool
	X        int
	Y        int
	Items    []MenuItem
	Groups   []MenuGroup
	OnSelect func(ctx context.Context, label string) error
}

// ContextMenu renders a menu at the host pointer coordinates. It returns nil
// while closed; the host opens it from a right-click handler and closes it on
// Escape, click-away, or selection. The menu is absolutely positioned, so
// mount it inside the block container that owns the clickable area.
func ContextMenu(options ContextMenuOptions) *ui.Element {
	if !options.Open {
		return nil
	}
	skin := contextMenuSkin()
	pane := menuPane(&skin, options.ID, options.Label, options.Items, options.Groups, options.OnSelect)
	return pane.Position(ui.PositionAbsolute).
		Left(ui.Px(float64(options.X))).
		Top(ui.Px(float64(options.Y)))
}
