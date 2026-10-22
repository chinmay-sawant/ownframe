package page

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
)

// PatchElement re-lays out one element of the retained page after editing it
// in the cached tree. The engine lays the element's subtree out again and
// hands the patched page back, so the page adopts it and marks the changed
// region dirty.
//
// The patch is atomic: a refused patch puts the tree back to the source's
// state.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) error {
	if err := p.patchReady(id); err != nil {
		return err
	}
	if !p.applyElementEdits(id, edits) {
		return fmt.Errorf("no change: %w", ErrPatchRefused)
	}
	state := p.renderState()
	styled, err := p.cache.RelayoutFor(ctx, p.width, p.height, state)
	if err != nil {
		return p.refusePatch(err)
	}
	patch, err := layout.RelayoutElement(ctx, styled, p.draw.Display, id,
		layout.Options{Images: state.Images})
	if err != nil {
		return p.refusePatch(err)
	}
	if !p.spliceElement(patch, id, edits) {
		return p.refusePatch(nil)
	}
	p.cache.SetStyled(styled)
	p.statePseudos = p.cache.StatePseudos()

	return nil
}
