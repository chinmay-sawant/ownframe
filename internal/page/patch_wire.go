package page

import (
	"context"

	"github.com/chinmay-sawant/ownframe/internal/element"
	"github.com/chinmay-sawant/ownframe/internal/page/patch"
	"github.com/chinmay-sawant/ownframe/internal/render"
	"github.com/chinmay-sawant/ownframe/internal/state"
)

// DisplayPatch reports how a retained display changed.
type DisplayPatch = patch.DisplayPatch

type opBuffer = patch.OpBuffer

// ElementEdit is one change to an element in the retained tree: an attribute
// to set, or the text to replace. A nil Text and an empty Attribute change
// nothing.
//
// ID names the element the edit applies to. An empty ID means the element
// the patch is for, and an edit on a descendant rides along with the patch.
// Text replaces the text content and drops the children.
//
// Text set this way does not go through the emoji pass. The pass runs on the
// source before it is parsed, and a text edit lands on the parsed tree, so an
// emoji in patched text paints as text rather than as the color image the
// same emoji gets from the template. The engine's tree API only writes a text
// node, so the image cannot be built inside the tree.
type ElementEdit = patch.ElementEdit

// LastPatch returns the last patch report; use it.
func (p *Page) LastPatch() DisplayPatch {
	return p.lastPatch
}

// Deprecated: use LastPatch.
func (p *Page) PatchReport() DisplayPatch {
	return p.LastPatch()
}

// Deprecated: use LastPatch.
func (p *Page) DisplayPatch() DisplayPatch {
	return p.LastPatch()
}

// PatchDisplay mutates the retained display and records the regions that
// changed. It returns the patch report. A nil display reports nothing.
func (p *Page) PatchDisplay(fn func(d *Display)) DisplayPatch {
	if p.draw.Display == nil || fn == nil {
		return p.lastPatch
	}

	p.tickMutated.Store(true)
	diff, buf := patch.Update(&displayOpPool, p.draw.Display, fn)
	if diff.Patch.Full {
		p.markFull()
	} else {
		for i := range diff.N {
			p.markRect(diff.Rects[i])
		}
	}

	report := patch.WithProps(diff.Patch, p.ChangedProperties())
	p.lastPatch = report
	patch.Free(&displayOpPool, buf)

	return report
}

// PatchOp mutates one retained operation and dirties its old and new bounds.
func (p *Page) PatchOp(i int, fn func(o *DisplayOp)) bool {
	p.tickMutated.Store(true)
	r, ok := patch.OpDamage(p.draw.Display, i, fn)
	if !ok {
		return false
	}

	p.markRect(r)
	p.lastPatch = patch.WithProps(patch.DisplayPatch{Changed: 1, Regions: 1}, p.ChangedProperties())

	return true
}

func (p *Page) rentOps(n int) *opBuffer {
	return patch.Rent(&displayOpPool, n)
}

func (p *Page) freeOps(buf *opBuffer) {
	patch.Free(&displayOpPool, buf)
}

// PatchElement re-lays out one element of the retained page after editing it
// in the cached tree. The engine lays the element's subtree out again and
// hands the patched page back, so the page adopts it and marks the changed
// region dirty.
//
// The patch is atomic: a refused patch puts the tree back to the source's
// state.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) error {
	if p == nil {
		return ErrNilPage
	}

	p.tickMutated.Store(true)
	err := patch.ApplyElement(ctx, patch.ElementIn{
		Refused:  ErrPatchRefused,
		ID:       id,
		Edits:    edits,
		Cache:    p.cache,
		Display:  p.draw.Display,
		Retain:   p.retainState,
		Width:    p.width,
		Height:   p.height,
		State:    func() render.State { return p.renderState() },
		Adopt:    func(d *Display) { p.spliceElement(d, id, edits) },
		Rollback: p.rollbackPatch,
	})
	if err != nil {
		return err
	}

	p.statePseudos = p.cache.StatePseudos()

	return nil
}

// spliceElement adopts the patched page. The engine hands over the whole
// display, because a patch that changes the operation count moves every later
// operation index. The page rebuilds its box index, marks the changed
// operations dirty, drops the picture, and refreshes the a11y tree.
func (p *Page) spliceElement(d *Display, id string, edits []ElementEdit) {
	captures := p.captureTransitions(p.draw.Display, d)
	prev := p.adoptSpliced(d)
	p.bindMotion(prev, d)
	p.bindTransitions(captures)
	p.refreshAccessibility(id, edits)
}

// adoptSpliced installs a display the engine spliced and returns the list the
// page drew before it. It is the shared adoption body for an element edit and
// for the layout driver's relayout: boxes and hits are rebuilt, the difference
// between the two displays is marked dirty, and the generation is bumped. A
// moved sibling is outside the patched element's own box, so the display diff
// is the damage source, not ElementPatch.Old and New alone.
func (p *Page) adoptSpliced(d *Display) *Display {
	p.tickMutated.Store(true)
	prev := p.draw.SwapList(d)
	p.setDisplayBoxes(d)
	p.dirtyFromDisplay(prev, d)
	p.generation++
	p.layerGen++
	p.stats.layouts++
	p.stats.repaints++

	return prev
}

// rollbackPatch undoes the edits of a refused patch. The cached tree is
// dropped, so the next redraw parses the source again and the page shows the
// source's state. The render key is forgotten, so it redraws.
func (p *Page) rollbackPatch() {
	p.invalidateCache()
	p.hasLast = false
}

// ChangedProperties returns a snapshot, falling back to the bound store when empty.
func (p *Page) ChangedProperties() []string {
	if p == nil {
		return nil
	}

	return p.stateBinding.ChangedProperties()
}

// TakeChangedProperties drains local metadata; the store fallback remains a snapshot.
func (p *Page) TakeChangedProperties() []string {
	if p == nil {
		return nil
	}

	return p.stateBinding.TakeChangedProperties()
}

// SetChangedProperties copies handler metadata into the binding.
func (p *Page) SetChangedProperties(props []string) {
	if p == nil {
		return
	}

	p.stateBinding.SetChangedProperties(props)
}

// AddChangedProperty appends a nonempty property path.
func (p *Page) AddChangedProperty(prop string) {
	if p == nil {
		return
	}

	p.stateBinding.AddChangedProperties(prop)
}

// AddChangedProperties appends nonempty property paths.
func (p *Page) AddChangedProperties(props []string) {
	if p == nil {
		return
	}

	p.stateBinding.AddChangedProperties(props...)
}

// ClearChangedProperties clears local metadata without clearing the store.
func (p *Page) ClearChangedProperties() {
	if p == nil {
		return
	}

	p.stateBinding.ClearChangedProperties()
}

// State returns the bound state store.
func (p *Page) State() *state.Store { return p.stateBinding.Store() }

// BindState replaces the subscription. Bind nil to detach.
func (p *Page) BindState(s *state.Store) {
	if p != nil {
		p.stateBinding.Bind(s)
	}
}

// SyncState refreshes surfaced metadata from the current binding only.
func (p *Page) SyncState() {
	if p != nil {
		p.stateBinding.Sync()
	}
}

func (p *Page) ensureElements() *element.Registry {
	if p == nil {
		return patch.Ensure(nil)
	}

	p.elements = patch.Ensure(p.elements)

	return p.elements
}

// Element returns a retained handle for id.
func (p *Page) Element(id string) *element.Node {
	return patch.Node(p.ensureElements(), id)
}

// Elements returns all retained handles.
func (p *Page) Elements() []*element.Node {
	return patch.All(p.ensureElements())
}

// RetainedRoot returns the retained tree root when set.
func (p *Page) RetainedRoot() *element.Node {
	return patch.Root(p.ensureElements())
}

// ElementRegistry exposes the underlying registry.
func (p *Page) ElementRegistry() *element.Registry {
	return p.ensureElements()
}
