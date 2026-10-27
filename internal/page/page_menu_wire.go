package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// MenuItem is one window context-menu row.
type MenuItem = host.MenuItem

// ContextMenuAt resolves the same painted target used by page clicks.
func (p *Page) ContextMenuAt(x, y float64) []host.MenuItem {
	if p == nil {
		return nil
	}
	spot := p.boxAt(x, y)
	return p.ContextMenuFor(spot.id)
}

// ContextMenuFor returns the rows for one hit-tested element. A page with no
// custom handler falls back to the built-in editing menu.
func (p *Page) ContextMenuFor(id string) []host.MenuItem {
	if p == nil {
		return nil
	}
	if p.handlers.ContextMenuFor != nil {
		if rows := p.handlers.ContextMenuFor(id); len(rows) > 0 {
			return rows
		}
	}

	return p.ContextMenu()
}

// PresentMenu lets the page paint the context menu. A page without a handler
// leaves the window chrome in place.
func (p *Page) PresentMenu(x, y int, items []host.MenuItem) bool {
	if p == nil || p.handlers.PresentMenu == nil {
		return false
	}

	return p.handlers.PresentMenu(x, y, items)
}

// MenuAction runs one custom row. A page without a handler ignores it.
func (p *Page) MenuAction(ctx context.Context, id string) error {
	if p == nil || p.handlers.MenuAction == nil {
		return nil
	}

	return p.handlers.MenuAction(ctx, id)
}
