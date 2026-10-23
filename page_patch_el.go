package ownframe

import "context"

// PatchElement re-lays out one retained element after editing it.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) error {
	return innerPage(p).PatchElement(ctx, id, edits...)
}
