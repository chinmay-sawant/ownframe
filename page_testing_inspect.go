package ownframe

import (
	"context"
	"image"
	"time"
)

// TakeDirty drains the host dirty bounding box.
func (page *Page) TakeDirty() (image.Rectangle, bool) {
	p := innerPage(page)
	if p == nil {
		return image.Rectangle{}, false
	}

	return p.TakeDirty()
}

// TakeDirtyRects drains host dirty regions.
func (page *Page) TakeDirtyRects() ([]image.Rectangle, bool) {
	p := innerPage(page)
	if p == nil {
		return nil, false
	}

	return p.TakeDirtyRects()
}

// MarkRect adds a CSS-pixel region to the dirty rect from a tick.
func (page *Page) MarkRect(r image.Rectangle) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.MarkRect(r)
}

// UseFrameDirty opts the next window draw into the buffer path.
func (page *Page) UseFrameDirty() {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.UseFrameDirty()
}

// FrameDirty reports and clears the frame-dirty opt-in.
func (page *Page) FrameDirty() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.FrameDirty()
}

// PNG encodes the page to PNG bytes.
func (page *Page) PNG() []byte {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.PNG()
}

// PNGWithContext threads ctx to the on-demand render.
func (t Testing) PNGWithContext(ctx context.Context) []byte {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PNGWithContext(ctx)
}

// Display returns the last display list, or nil for a rasterized page.
func (page *Page) Display() *Display {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Display()
}

// Image returns the last rasterized picture, or nil for a display-list page.
func (page *Page) Image() image.Image {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Image()
}

// Element returns a retained handle for id.
func (page *Page) Element(id string) *Element {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Element(id)
}

// Elements returns all retained handles.
func (page *Page) Elements() []*Element {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Elements()
}

// ElementRegistry exposes the retained element registry.
func (page *Page) ElementRegistry() *ElementRegistry {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.ElementRegistry()
}

// Invalidate marks the region the element id changed.
func (page *Page) Invalidate(id string) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.Invalidate(id)
}

// Stats returns one snapshot of the page counters an inspector shows.
func (page *Page) Stats() Stats {
	p := innerPage(page)
	if p == nil {
		return Stats{}
	}

	return p.Stats()
}

// Generation increases by one on every successful Redraw.
func (page *Page) Generation() uint64 {
	p := innerPage(page)
	if p == nil {
		return 0
	}

	return p.Generation()
}

// ChangedProperties returns changed-property metadata surfaced to the page.
func (t Testing) ChangedProperties() []string {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.ChangedProperties()
}

// ClearChangedProperties clears surfaced metadata.
func (t Testing) ClearChangedProperties() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.ClearChangedProperties()
}

// AddChangedProperty appends one property path.
func (page *Page) AddChangedProperty(prop string) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.AddChangedProperty(prop)
}

// AddChangedProperties appends multiple property paths.
func (page *Page) AddChangedProperties(props []string) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.AddChangedProperties(props)
}

// SetChangedProperties records property metadata surfaced from a handler.
func (page *Page) SetChangedProperties(props []string) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetChangedProperties(props)
}

// TakeChangedProperties returns a snapshot and drains local metadata.
func (t Testing) TakeChangedProperties() []string {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.TakeChangedProperties()
}

// TakeScroll returns the pending scroll request and clears it.
func (t Testing) TakeScroll() (Scroll, bool) {
	p := innerPage(t.page)
	if p == nil {
		return Scroll{}, false
	}
	req, ok := p.TakeScroll()
	return Scroll{X: req.X, Y: req.Y, Absolute: req.Absolute}, ok
}

// StepScrollWindow prepares visible rows and reports whether they changed.
func (t Testing) StepScrollWindow() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.StepScrollWindow()
}

// NextFrameAt returns the deadline set by WantFrameAt and clears it.
func (t Testing) NextFrameAt() time.Time {
	p := innerPage(t.page)
	if p == nil {
		return time.Time{}
	}

	return p.NextFrameAt()
}

// ViewLocked reports that the window should scale this page to the screen.
func (page *Page) ViewLocked() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.ViewLocked()
}

// Title returns the window title.
func (page *Page) Title() string {
	p := innerPage(page)
	if p == nil {
		return ""
	}

	return p.Title()
}

// Ticking reports whether a frame callback is registered.
func (page *Page) Ticking() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.Ticking()
}

// Boxes returns a copy of the last hit-test boxes.
func (page *Page) Boxes() []Box {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Boxes()
}

// AccessibilityTree returns current semantics in document order.
func (page *Page) AccessibilityTree() (AccessibilityTree, error) {
	p := innerPage(page)
	if p == nil {
		return AccessibilityTree{}, nil
	}

	return p.AccessibilityTree()
}

// AccessibilityPerform applies a tree action.
// The accessibility example drives this from app code.
func (page *Page) AccessibilityPerform(ctx context.Context, request AccessibilityRequest) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.AccessibilityPerform(ctx, request)
}

// ScrollOffset returns the last offset the window reported.
func (page *Page) ScrollOffset() (int, int) {
	p := innerPage(page)
	if p == nil {
		return 0, 0
	}

	return p.ScrollOffset()
}

// SetScrollOffset records the window offset. Tests drive the row window with it.
func (page *Page) SetScrollOffset(x, y int) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetScrollOffset(x, y)
}

// SetWindowing opts into redraws on scroll changes.
func (page *Page) SetWindowing(on bool) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetWindowing(on)
}

// Windowing reports whether scroll changes redraw this page.
func (page *Page) Windowing() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.Windowing()
}

// ScrollTo queues an absolute scroll move for the window.
func (page *Page) ScrollTo(x, y int) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.ScrollTo(x, y)
}

// ScrollBy queues a scroll delta for the window.
func (page *Page) ScrollBy(dx, dy int) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.ScrollBy(dx, dy)
}

// SetAllowScroll turns window panning on or off.
func (page *Page) SetAllowScroll(on bool) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.SetAllowScroll(on)
}

// AllowScroll reports that the window may pan this page.
func (page *Page) AllowScroll() bool {
	p := innerPage(page)
	if p == nil {
		return false
	}

	return p.AllowScroll()
}
