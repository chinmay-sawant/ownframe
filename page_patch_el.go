package ownframe

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/page"
)

// ElementPatch is one element's share of a batched patch: the element id
// plus the tree edits for it.
type ElementPatch = page.ElementPatch

// PatchElement re-lays out one retained element after editing it.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) error {
	return innerPage(p).PatchElement(ctx, id, edits...)
}

// PatchElements edits several retained elements in one batch with a single
// shared style re-resolve. An empty batch returns nil.
func (p *Page) PatchElements(ctx context.Context, patches []ElementPatch) error {
	return innerPage(p).PatchElements(ctx, patches)
}
