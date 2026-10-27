package ownframe

import "context"

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
