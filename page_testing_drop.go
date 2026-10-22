package ownframe

import "context"

// Drop calls the drop handler with the files dropped on the window.
func (t Testing) Drop(ctx context.Context, files []Drop) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Drop(ctx, files)
}
