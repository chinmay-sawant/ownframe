package ownframe

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
)

// Drop is one file or directory dropped on the window.
type Drop = host.Drop

// HoverHandler paints a hover state without layout when it returns true.
type HoverHandler = page.HoverHandler

// ContextMenuFor returns the custom right-click rows for one element id.
// A page with no custom provider falls back to the built-in edit menu.
func (p *Page) ContextMenuFor(id string) []MenuItem {
	if p == nil || p.inner == nil {
		return nil
	}

	return p.inner.ContextMenuFor(id)
}

// MenuAction runs one custom right-click row.
func (p *Page) MenuAction(ctx context.Context, id string) error {
	if p == nil || p.inner == nil {
		return nil
	}

	return p.inner.MenuAction(ctx, id)
}
