package ownframe

import (
	"context"
	"image"
)

// Probe reads paint, dirty regions, and counters.
// Example programs call these methods. Simulated input stays on Testing.
type Probe struct{ page *Page }

// Probe returns paint, dirty-region, and counter reads for this page.
func (p *Page) Probe() Probe { return Probe{page: p} }

// TakeDirty drains the host dirty bounding box.
func (pr Probe) TakeDirty() (image.Rectangle, bool) {
	p := innerPage(pr.page)
	if p == nil {
		return image.Rectangle{}, false
	}

	return p.TakeDirty()
}

// TakeDirtyRects drains host dirty regions.
func (pr Probe) TakeDirtyRects() ([]image.Rectangle, bool) {
	p := innerPage(pr.page)
	if p == nil {
		return nil, false
	}

	return p.TakeDirtyRects()
}

// MarkRect adds a CSS-pixel region to the dirty rect from a tick.
func (pr Probe) MarkRect(r image.Rectangle) {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.MarkRect(r)
}

// UseFrameDirty opts the next window draw into the buffer path.
func (pr Probe) UseFrameDirty() {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.UseFrameDirty()
}

// FrameDirty reports and clears the frame-dirty opt-in.
func (pr Probe) FrameDirty() bool {
	p := innerPage(pr.page)
	if p == nil {
		return false
	}

	return p.FrameDirty()
}

// PNG encodes the page to PNG bytes.
func (pr Probe) PNG() []byte {
	p := innerPage(pr.page)
	if p == nil {
		return nil
	}

	return p.PNG()
}

// Display returns the last display list, or nil for a rasterized page.
func (pr Probe) Display() *Display {
	p := innerPage(pr.page)
	if p == nil {
		return nil
	}

	return p.Display()
}

// Image returns the last rasterized picture, or nil for a display-list page.
func (pr Probe) Image() image.Image {
	p := innerPage(pr.page)
	if p == nil {
		return nil
	}

	return p.Image()
}

// Element returns a retained handle for id.
func (pr Probe) Element(id string) *Element {
	p := innerPage(pr.page)
	if p == nil {
		return nil
	}

	return p.Element(id)
}

// Elements returns all retained handles.
func (pr Probe) Elements() []*Element {
	p := innerPage(pr.page)
	if p == nil {
		return nil
	}

	return p.Elements()
}

// ElementRegistry exposes the retained element registry.
func (pr Probe) ElementRegistry() *ElementRegistry {
	p := innerPage(pr.page)
	if p == nil {
		return nil
	}

	return p.ElementRegistry()
}

// Invalidate marks the region the element id changed.
func (pr Probe) Invalidate(id string) {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.Invalidate(id)
}

// Stats returns one snapshot of the page counters an inspector shows.
func (pr Probe) Stats() Stats {
	p := innerPage(pr.page)
	if p == nil {
		return Stats{}
	}

	return p.Stats()
}

// Generation increases by one on every successful Redraw.
func (pr Probe) Generation() uint64 {
	p := innerPage(pr.page)
	if p == nil {
		return 0
	}

	return p.Generation()
}

// AddChangedProperty appends one property path.
func (pr Probe) AddChangedProperty(prop string) {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.AddChangedProperty(prop)
}

// AddChangedProperties appends multiple property paths.
func (pr Probe) AddChangedProperties(props []string) {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.AddChangedProperties(props)
}

// SetChangedProperties records property metadata surfaced from a handler.
func (pr Probe) SetChangedProperties(props []string) {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.SetChangedProperties(props)
}

// ViewLocked reports that the window should scale this page to the screen.
func (pr Probe) ViewLocked() bool {
	p := innerPage(pr.page)
	if p == nil {
		return false
	}

	return p.ViewLocked()
}

// PatchOp mutates one retained operation.
func (pr Probe) PatchOp(i int, fn func(o *DisplayOp)) bool {
	p := innerPage(pr.page)
	if p == nil {
		return false
	}

	return p.PatchOp(i, fn)
}

// PatchDisplay mutates the retained display and records changed regions.
func (pr Probe) PatchDisplay(fn func(d *Display)) DisplayPatch {
	p := innerPage(pr.page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.PatchDisplay(fn)
}

// LastPatch returns the last display patch report.
func (pr Probe) LastPatch() DisplayPatch {
	p := innerPage(pr.page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.LastPatch()
}

// BeginRedrawBatch defers draws until EndRedrawBatch.
func (pr Probe) BeginRedrawBatch() {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.BeginRedrawBatch()
}

// EndRedrawBatch closes a batch and draws the final state once.
func (pr Probe) EndRedrawBatch(ctx context.Context) error {
	p := innerPage(pr.page)
	if p == nil {
		return nil
	}

	return p.EndRedrawBatch(ctx)
}

// SetAfterRedraw registers fn to run after a redraw commits.
func (pr Probe) SetAfterRedraw(fn func()) {
	p := innerPage(pr.page)
	if p == nil {
		return
	}
	p.SetAfterRedraw(fn)
}

// Ticking reports whether a frame callback is registered.
func (pr Probe) Ticking() bool {
	p := innerPage(pr.page)
	if p == nil {
		return false
	}

	return p.Ticking()
}

// PressedID is the control held by the current pointer gesture.
func (pr Probe) PressedID() string {
	p := innerPage(pr.page)
	if p == nil {
		return ""
	}

	return p.PressedID()
}
