package ownframe

import (
	"context"
	"image"
	"time"
)

// Testing exposes host-loop helpers for tests and engine examples.
// The window already calls these on the inner screen.
type Testing struct{ page *Page }

func (p *Page) Testing() Testing { return Testing{page: p} }

// TakeDirty drains the host dirty bounding box.
func (t Testing) TakeDirty() (image.Rectangle, bool) {
	p := innerPage(t.page)
	if p == nil {
		return image.Rectangle{}, false
	}

	return p.TakeDirty()
}

// TakeDirtyRects drains host dirty regions.
func (t Testing) TakeDirtyRects() ([]image.Rectangle, bool) {
	p := innerPage(t.page)
	if p == nil {
		return nil, false
	}

	return p.TakeDirtyRects()
}

// PollReload rereads a watched source file.
func (t Testing) PollReload(ctx context.Context) (bool, error) {
	p := innerPage(t.page)
	if p == nil {
		return false, nil
	}

	return p.PollReload(ctx)
}

// IMEReplace replaces a byte range in the focused value.
func (t Testing) IMEReplace(ctx context.Context, start, end int, insert string, caret int) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.IMEReplace(ctx, start, end, insert, caret)
}

// PatchOp mutates one retained operation.
func (t Testing) PatchOp(i int, fn func(o *DisplayOp)) bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.PatchOp(i, fn)
}

// Tick runs the frame callback registered with SetTick and the caret blink.
func (t Testing) Tick(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Tick(ctx)
}

// SetAfterRedraw registers fn to run after a redraw commits.
func (t Testing) SetAfterRedraw(fn func()) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetAfterRedraw(fn)
}

// BeginRedrawBatch defers draws until EndRedrawBatch.
func (t Testing) BeginRedrawBatch() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.BeginRedrawBatch()
}

// EndRedrawBatch closes a batch and draws the final state once.
func (t Testing) EndRedrawBatch(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.EndRedrawBatch(ctx)
}

// DispatchDeepLink validates a host-delivered URL and calls the app handler.
func (t Testing) DispatchDeepLink(ctx context.Context, target string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.DispatchDeepLink(ctx, target)
}

// ViewLocked reports that the window should scale this page to the screen.
func (t Testing) ViewLocked() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.ViewLocked()
}

// Click hit-tests the last picture and applies the click.
func (t Testing) Click(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Click(ctx, x, y)
}

// Hover sets the hovered element from a point.
func (t Testing) Hover(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Hover(ctx, x, y)
}

// Press sets the pressed element from a point.
func (t Testing) Press(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Press(ctx, x, y)
}

// Release clears the pressed element.
func (t Testing) Release(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Release(ctx)
}

// Type appends text to the focused field.
func (t Testing) Type(ctx context.Context, text string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Type(ctx, text)
}

// MarkRect adds a CSS-pixel region to the dirty rect from a tick.
func (t Testing) MarkRect(r image.Rectangle) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.MarkRect(r)
}

// UseFrameDirty opts the next window draw into the buffer path.
func (t Testing) UseFrameDirty() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.UseFrameDirty()
}

// FrameDirty reports and clears the frame-dirty opt-in.
func (t Testing) FrameDirty() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.FrameDirty()
}

// Drop calls the drop handler with the files dropped on the window.
func (t Testing) Drop(ctx context.Context, files []Drop) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Drop(ctx, files)
}

// Element returns a retained handle for id.
func (t Testing) Element(id string) *Element {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Element(id)
}

// Elements returns all retained handles.
func (t Testing) Elements() []*Element {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Elements()
}

// ElementRegistry exposes the retained element registry.
func (t Testing) ElementRegistry() *ElementRegistry {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.ElementRegistry()
}

// Invalidate marks the region the element id changed.
func (t Testing) Invalidate(id string) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.Invalidate(id)
}

// LongPress fires a held press at a point.
func (t Testing) LongPress(ctx context.Context, x, y float64) (bool, error) {
	p := innerPage(t.page)
	if p == nil {
		return false, nil
	}

	return p.LongPress(ctx, x, y)
}

// Swipe delivers a swipe delta to the handler.
func (t Testing) Swipe(ctx context.Context, dx, dy float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Swipe(ctx, dx, dy)
}

// BeginDrag starts a pointer drag.
func (t Testing) BeginDrag(ctx context.Context, x, y float64) (bool, error) {
	p := innerPage(t.page)
	if p == nil {
		return false, nil
	}

	return p.BeginDrag(ctx, x, y)
}

// MoveDrag continues a pointer drag.
func (t Testing) MoveDrag(ctx context.Context, x, y float64) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.MoveDrag(ctx, x, y)
}

// EndDrag finishes a pointer drag.
func (t Testing) EndDrag(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.EndDrag(ctx)
}

// IMEContext returns the focused field's box and caret text.
func (t Testing) IMEContext() (Box, int, string, string, bool) {
	p := innerPage(t.page)
	if p == nil {
		return Box{}, 0, "", "", false
	}

	return p.IMEContext()
}

// Backspace drops the last rune of the focused field.
func (t Testing) Backspace(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Backspace(ctx)
}

// DeleteWord deletes the word behind the caret.
func (t Testing) DeleteWord(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.DeleteWord(ctx)
}

// Submit calls the submit handler and draws.
func (t Testing) Submit(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Submit(ctx)
}

// KeyDown calls the key-down handler and caret keys.
func (t Testing) KeyDown(ctx context.Context, key string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.KeyDown(ctx, key)
}

// KeyUp calls the key-up handler.
func (t Testing) KeyUp(ctx context.Context, key string) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.KeyUp(ctx, key)
}

// PrepareMobile prepares this page for its native mobile viewport.
func (t Testing) PrepareMobile(ctx context.Context) error {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PrepareMobile(ctx)
}

// NextFrameAt returns the deadline set by WantFrameAt and clears it.
func (t Testing) NextFrameAt() time.Time {
	p := innerPage(t.page)
	if p == nil {
		return time.Time{}
	}

	return p.NextFrameAt()
}

// SetPerf turns pipeline timing, dirty counters, and alloc tracking on or off.
func (t Testing) SetPerf(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetPerf(on)
}

// Perf reports whether pipeline timing and dirty counters are recorded.
func (t Testing) Perf() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.Perf()
}

// SetDevTools turns the window devtools overlay on or off.
func (t Testing) SetDevTools(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetDevTools(on)
}

// DevTools reports whether the window devtools overlay starts on.
func (t Testing) DevTools() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.DevTools()
}

// Stats returns one snapshot of the page counters an inspector shows.
func (t Testing) Stats() Stats {
	p := innerPage(t.page)
	if p == nil {
		return Stats{}
	}

	return p.Stats()
}

// PatchDisplay mutates the retained display and records changed regions.
func (t Testing) PatchDisplay(fn func(d *Display)) DisplayPatch {
	p := innerPage(t.page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.PatchDisplay(fn)
}

// LastPatch returns the last display patch report.
func (t Testing) LastPatch() DisplayPatch {
	p := innerPage(t.page)
	if p == nil {
		return DisplayPatch{}
	}

	return p.LastPatch()
}

// SetPicker installs the dialog a file control opens.
func (t Testing) SetPicker(pick func(ctx context.Context, title string) (string, error)) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetPicker(pick)
}

// CancelTouches queues gesture cancellation for the window loop.
func (t Testing) CancelTouches() {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.CancelTouches()
}

// SetTouchZoomAllowed controls two-finger touch zoom.
func (t Testing) SetTouchZoomAllowed(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetTouchZoomAllowed(on)
}

// TouchZoomAllowed reports whether two-finger touch may zoom.
func (t Testing) TouchZoomAllowed() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.TouchZoomAllowed()
}

// SetTouchScrollOptions tunes finger sensitivity and inertial decay.
func (t Testing) SetTouchScrollOptions(options TouchScrollOptions) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetTouchScrollOptions(options)
}

// PressedID is the control held by the current pointer gesture.
func (t Testing) PressedID() string {
	p := innerPage(t.page)
	if p == nil {
		return ""
	}

	return p.PressedID()
}

// Generation increases by one on every successful Redraw.
func (t Testing) Generation() uint64 {
	p := innerPage(t.page)
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
func (t Testing) AddChangedProperty(prop string) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.AddChangedProperty(prop)
}

// AddChangedProperties appends multiple property paths.
func (t Testing) AddChangedProperties(props []string) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.AddChangedProperties(props)
}

// SetChangedProperties records property metadata surfaced from a handler.
func (t Testing) SetChangedProperties(props []string) {
	p := innerPage(t.page)
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

// SetHotReload turns the file watch on or off.
func (t Testing) SetHotReload(on bool) {
	p := innerPage(t.page)
	if p == nil {
		return
	}
	p.SetHotReload(on)
}

// Watching reports whether a poll can reread a source file.
func (t Testing) Watching() bool {
	p := innerPage(t.page)
	if p == nil {
		return false
	}

	return p.Watching()
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

// PNG encodes the page to PNG bytes.
func (t Testing) PNG() []byte {
	p := innerPage(t.page)
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
func (t Testing) Display() *Display {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Display()
}

// Image returns the last rasterized picture, or nil for a display-list page.
func (t Testing) Image() image.Image {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.Image()
}
