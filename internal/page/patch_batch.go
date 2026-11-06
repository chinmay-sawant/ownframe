package page

import (
	"context"
	"fmt"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/patch"
	"github.com/chinmay-sawant/ownframe/internal/render"
)

// ElementPatch is one element's share of a batched patch: the element id
// plus the tree edits for it. An edit with an empty ID targets the element
// the patch names, the same rule ElementEdit documents.
type ElementPatch struct {
	ID    string
	Edits []ElementEdit
}

// PatchElements edits several retained elements in one batch and re-lays
// them out together. An empty batch returns nil without touching the page.
//
// OPT-07-02: every patch's tree edits land first, then the page runs ONE
// shared style re-resolve (RelayoutFor) for the whole batch instead of one
// per element. Each touched element is re-laid out in order against that
// single styled document, chaining from one element's patched display to
// the next, and the final display is spliced in and drawn once, so a batch
// of N patches pays one re-resolve, N subtree relayouts, and one adoption.
//
// The batch is atomic: a refused element rolls the tree back to the
// source's state and the error wraps ErrPatchRefused, so callers fall back
// to a refresh. An entry whose edits match nothing rides along without
// failing the batch; the batch is refused only when no entry changes
// anything, the way one PatchElement with no change is refused.
func (p *Page) PatchElements(ctx context.Context, patches []ElementPatch) error {
	if p == nil {
		return ErrNilPage
	}
	if len(patches) == 0 {
		return nil
	}
	defer p.notifyFrame()
	clear(p.hoverScopes)

	p.tickMutated.Store(true)
	for i := range patches {
		if err := patch.Ready(ErrPatchRefused, patches[i].ID, p.cache != nil, p.draw.Display != nil, p.retainState); err != nil {
			return err
		}
	}

	touched := make([]bool, len(patches))
	changed := false
	for i := range patches {
		if patch.Apply(p.cache, patches[i].ID, patches[i].Edits) {
			touched[i] = true
			changed = true
		}
	}
	if !changed {
		return fmt.Errorf("no change: %w", ErrPatchRefused)
	}

	state := p.renderState()
	// OPT-07-02: one shared style re-resolve for the whole batch. The
	// per-element loop below re-places subtrees only; it never re-resolves.
	styled, err := p.cache.RelayoutFor(ctx, p.width, p.height, state)
	if err != nil {
		return p.failBatch(err)
	}

	current := p.draw.Display
	for i := range patches {
		if !touched[i] {
			continue
		}
		ep, err := render.RelayoutElement(ctx, styled, current, patches[i].ID, layout.Options{Images: state.Images})
		if err != nil {
			return p.failBatch(err)
		}
		next, ok := patch.AdoptedDisplay(ep)
		if !ok {
			return p.failBatch(nil)
		}
		current = next
	}

	captures := p.captureTransitions(p.draw.Display, current)
	prev := p.adoptSpliced(current)
	p.bindMotion(prev, current)
	p.bindTransitions(captures)
	for i := range patches {
		if touched[i] {
			p.refreshAccessibility(patches[i].ID, patches[i].Edits)
		}
	}
	p.cache.SetStyled(styled)
	p.statePseudos = p.cache.StatePseudos()
	// The patched display is now the committed frame. Record its key so the
	// redraw a click dispatches after its handler sees the frame as unchanged
	// instead of laying the whole page out again.
	p.remember(state)

	return nil
}

// failBatch rolls a refused batch back and wraps the cause in
// ErrPatchRefused, so callers fall back to a refresh.
func (p *Page) failBatch(err error) error {
	p.rollbackPatch()

	return patch.Refuse(ErrPatchRefused, err)
}
