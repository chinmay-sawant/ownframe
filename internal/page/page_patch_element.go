package page

import (
	"context"

	"github.com/chinmay-sawant/blinkless/layout"
)

// ElementEdit is one change to an element in the retained tree: an attribute
// to set, or the text to replace. A nil Text and an empty Attribute change
// nothing.
//
// ID names the element the edit applies to. An empty ID means the element
// the patch is for. An edit on a descendant of the patched element rides
// along with the patch, because the patch lays the whole subtree out again.
type ElementEdit struct {
	ID        string
	Attribute string
	Value     string
	Text      *string
}

// PatchElement re-lays out one element of the retained page after editing it
// in the cached tree. The engine lays the element's subtree out again and
// hands the patched page back, so the page adopts it and marks the changed
// region dirty. It reports whether the patch applied; a false result leaves
// the caller to redraw the page.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) bool {
	if p == nil || p.cache == nil || p.display == nil || id == "" {
		return false
	}
	if !p.retainState {
		// The page keeps no layout state, so the engine cannot patch.
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
		return false
	}
	patch, err := layout.RelayoutElement(ctx, styled, p.display, id,
		layout.Options{Images: state.Images})
	if err != nil {
		return false
	}
	if !p.spliceElement(patch) {
		return false
	}
	p.cache.SetStyled(styled)

	return true
}
