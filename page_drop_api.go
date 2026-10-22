package ownframe

import "context"

// Drop calls the drop handler with the files dropped on the window.
func (p *Page) Drop(ctx context.Context, files []Drop) error {
	return innerPage(p).Drop(ctx, files)
}
