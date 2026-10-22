package page

import (
	"context"

	"github.com/chinmay-sawant/blinkless/layout"
)

// PatchElement re-lays out one element of the retained page after editing it
// in the cached tree. The engine lays the element's subtree out again and
// hands the patched page back, so the page adopts it and marks the changed
// region dirty. It reports whether the patch applied.
//
// The patch is atomic: a refused patch puts the tree back to the source's
// state.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) bool {
	if p == nil || p.cache == nil || p.display == nil || id == "" {
		return false
	}
	if !p.retainState {
		// The page keeps no layout state.
		return false
	}
	changed := false
	for _, e := range edits {
		target := e.ID
		if target == "" {
			target = id
		}
		switch {
		case e.Text != nil:
			changed = p.cache.SetText(target, *e.Text) || changed
		case e.Attribute != "":
			changed = p.cache.SetAttribute(target, e.Attribute, e.Value) || changed
		}
	}
	if !changed {
		return false
	}
	state := p.renderState()
	styled, err := p.cache.RelayoutFor(ctx, p.width, p.height, state)
	if err != nil {
		p.rollbackPatch()
		return false
	}
	patch, err := layout.RelayoutElement(ctx, styled, p.display, id,
		layout.Options{Images: state.Images})
	if err != nil {
		p.rollbackPatch()
		return false
	}
	if !p.spliceElement(patch, id, edits) {
		p.rollbackPatch()
		return false
	}
	p.cache.SetStyled(styled)
	p.statePseudos = p.cache.StatePseudos()

	return true
}
