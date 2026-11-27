package ui

import (
	"context"
	"strings"

	"github.com/chinmay-sawant/ownframe/components/chadcn"
	"github.com/chinmay-sawant/ownframe/ui"
)

// Scrolling model. The engine does not scroll nested overflow containers, so
// every scrollable area is either a pane (children translated by a negative
// margin, range measured from layout) or a virtual list (grid, palette,
// editor) whose offset lives in its own model. Both draw the same themed thin
// scrollbar, chadcn.Scrollbar, driven by a chadcn.ScrollState per region. The
// bar is an overlay: it disappears while the content fits.

// region is one scrollable area's range state and the setter that pushes a new
// pixel offset into whatever model owns it.
type region struct {
	st    chadcn.ScrollState
	set   func(offset int)
	label string
	// unit is the pixels one wheel notch scrolls; zero uses wheelPixels.
	unit int
	// horizontal regions drag and page along x.
	horizontal bool
	// Thumb drag: the offset at the press and the pointer travel since.
	dragging  bool
	dragStart int
	dragSum   float64
}

func (a *App) region(key string) *region {
	if a.regions == nil {
		a.regions = map[string]*region{}
	}
	r := a.regions[key]
	if r == nil {
		r = &region{}
		r.set = func(o int) { r.st.ScrollTo(o) }
		a.regions[key] = r
	}
	return r
}

// scrollbarFor draws the overlay bar for region key from explicit metrics.
// set receives the new offset after a track click.
func (a *App) scrollbarFor(key, label string, content, viewport, offset int, set func(int)) *ui.Element {
	r := a.region(key)
	r.label, r.set = label, set
	r.st.SetRange(content, viewport)
	r.st.ScrollTo(offset)
	return chadcn.Scrollbar(chadcn.ScrollbarOptions{
		ID: key + "-bar", ThumbID: key + "-thumb", ControlsID: key + "-vp", Label: label,
		State: &r.st, TrackPixels: viewport,
		OnChange: func(ctx context.Context) error {
			set(r.st.Offset())
			return a.redraw(ctx)
		},
	})
}

// pane is a clipped viewport whose children scroll by a translated offset. Its
// range is measured after layout (syncPanes), so it needs no explicit height:
// give the viewport a size with the class or Height on the returned element.
func (a *App) pane(key, label, class string, children ...*ui.Element) *ui.Element {
	r := a.region(key)
	r.label = label
	r.set = func(o int) { r.st.ScrollTo(o) }
	if a.livePanes == nil {
		a.livePanes = map[string]bool{}
	}
	a.livePanes[key] = true
	innerClass := "qr-pane-inner"
	if a.modal != "" {
		innerClass = "qr-pane-inner qr-pane-inner-row"
	}
	inner := ui.Div().ID(key + "-inner").Class(innerClass).MarginTop(ui.Px(-float64(r.st.Offset()))).Children(children...)
	vp := ui.Div().ID(key + "-vp").Class("qr-pane " + class).Children(inner)
	m := r.st.Metrics()
	if a.modal != "" {
		// In-flow bar beside the content (see scrollbarFlow).
		return vp.Class("qr-pane qr-pane-row " + class).Children(a.scrollbarFlow(key, label, m.Content, m.Viewport, m.Offset, r.set))
	}
	return vp.Children(a.scrollbarFor(key, label, m.Content, m.Viewport, m.Offset, r.set))
}

// syncPanes measures every pane built in the last render and, when a range
// changed, rebuilds once so the thumb and offset match the real layout.
func (a *App) syncPanes(ctx context.Context) error {
	if len(a.livePanes) == 0 || a.syncing {
		return nil
	}
	boxes := map[string][2]float64{}
	for _, b := range a.page.Boxes() {
		if strings.HasSuffix(b.ID, "-vp") || strings.HasSuffix(b.ID, "-inner") {
			boxes[b.ID] = [2]float64{b.Y, b.H}
		}
	}
	changed := false
	for key := range a.livePanes {
		vp, ok1 := boxes[key+"-vp"]
		in, ok2 := boxes[key+"-inner"]
		if !ok1 || !ok2 {
			continue
		}
		if a.region(key).st.SetRange(int(in[1]), int(vp[1])) {
			changed = true
		}
	}
	if !changed {
		return nil
	}
	a.syncing = true
	defer func() { a.syncing = false }()
	return a.view.Update(ctx, a.screen())
}

// paneAt returns the pane whose viewport contains the point, if any.
func (a *App) paneAt(x, y float64) (string, bool) {
	for _, b := range a.page.Boxes() {
		if !strings.HasSuffix(b.ID, "-vp") {
			continue
		}
		key := strings.TrimSuffix(b.ID, "-vp")
		if !a.livePanes[key] {
			continue
		}
		if x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H {
			return key, true
		}
	}
	return "", false
}

// thumbKey returns the region a "<key>-thumb" element belongs to.
func (a *App) thumbKey(id string) (string, bool) {
	if strings.HasSuffix(id, "-hthumb") {
		key := strings.TrimSuffix(id, "-hthumb") + "-h"
		_, ok := a.regions[key]
		return key, ok
	}
	if !strings.HasSuffix(id, "-thumb") {
		return "", false
	}
	key := strings.TrimSuffix(id, "-thumb")
	_, ok := a.regions[key]
	return key, ok
}

// trackKey returns the region a track-page button ("<key>-bar-before/after")
// belongs to.
func (a *App) trackKey(id string) (string, bool) {
	for _, suf := range []string{"-hbar-before", "-hbar-after"} {
		if strings.HasSuffix(id, suf) {
			key := strings.TrimSuffix(id, suf) + "-h"
			_, ok := a.regions[key]
			return key, ok
		}
	}
	for _, suf := range []string{"-bar-before", "-bar-after"} {
		if strings.HasSuffix(id, suf) {
			key := strings.TrimSuffix(id, suf)
			_, ok := a.regions[key]
			return key, ok
		}
	}
	return "", false
}

// pressTrack puts the thumb under the pointer when the track is pressed.
func (a *App) pressTrack(ctx context.Context, key string, x, y float64) error {
	r := a.regions[key]
	var barY float64
	id, pos := key+"-bar", y
	if r.horizontal {
		id, pos = strings.TrimSuffix(key, "-h")+"-hbar", x
	}
	for _, b := range a.page.Boxes() {
		if b.ID == id {
			barY = b.Y
			if r.horizontal {
				barY = b.X
			}
		}
	}
	m := r.st.Metrics()
	r.set(m.OffsetAt(pos-barY, m.Viewport))
	return a.redraw(ctx)
}

// beginThumbDrag starts dragging region key's thumb.
func (a *App) beginThumbDrag(key string) {
	r := a.regions[key]
	r.dragging, r.dragStart, r.dragSum = true, r.st.Offset(), 0
	a.thumbDrag = key
}

// dragThumb moves the thumb by dy pixels of pointer travel.
func (a *App) dragThumb(ctx context.Context, dx, dy float64) error {
	r := a.regions[a.thumbDrag]
	if r == nil || !r.dragging {
		return nil
	}
	if r.horizontal {
		dy = dx
	}
	m := r.st.Metrics()
	_, th, vis := m.Thumb(m.Viewport)
	travel := m.Viewport - th
	if !vis || travel <= 0 {
		return nil
	}
	r.dragSum += dy
	off := r.dragStart + int(r.dragSum*float64(m.MaxOffset())/float64(travel))
	r.set(min(max(off, 0), m.MaxOffset()))
	return a.redraw(ctx)
}

func (a *App) endThumbDrag() {
	if r := a.regions[a.thumbDrag]; r != nil {
		r.dragging = false
	}
	a.thumbDrag = ""
}

// scrollbarFlow is the in-flow variant of the bar for content inside a
// position:fixed overlay (dialogs, the palette): the engine does not paint an
// absolutely positioned child of a fixed box, so the bar is a column laid out
// beside the content instead, with the same ids and behaviour.
func (a *App) scrollbarFlow(key, label string, content, viewport, offset int, set func(int)) *ui.Element {
	r := a.region(key)
	r.label, r.set = label, set
	r.st.SetRange(content, viewport)
	r.st.ScrollTo(offset)
	top, h, vis := r.st.Metrics().Thumb(viewport)
	if !vis {
		return nil
	}
	page := func(dir int) func(context.Context) error {
		return func(ctx context.Context) error {
			if r.st.Page(dir) {
				set(r.st.Offset())
			}
			return a.redraw(ctx)
		}
	}
	col := ui.Div().ID(key + "-bar").Class("qr-flowbar").Height(ui.Px(float64(viewport)))
	if top > 0 {
		col.Children(ui.Button("").ID(key+"-bar-before").Class("qr-flowgap").Attr("type", "button").Attr("tabindex", "-1").
			Attr("aria-hidden", "true").Height(ui.Px(float64(top))).OnClick(page(-1)))
	}
	col.Children(ui.Button("").ID(key+"-thumb").Class("qr-flowthumb").Attr("type", "button").Attr("role", "scrollbar").
		Attr("aria-label", label).Attr("aria-orientation", "vertical").Height(ui.Px(float64(h))))
	if rest := viewport - top - h; rest > 0 {
		col.Children(ui.Button("").ID(key+"-bar-after").Class("qr-flowgap").Attr("type", "button").Attr("tabindex", "-1").
			Attr("aria-hidden", "true").Height(ui.Px(float64(rest))).OnClick(page(1)))
	}
	return col
}

// rowPane is a fixed-height pane of 30 px rows that scrolls one row per notch,
// so a row is never cut at the pane edge.
func (a *App) rowPane(key, label string, content *ui.Element, height int) *ui.Element {
	p := a.pane(key, label, "qr-pane-fixed", content).Height(ui.Px(float64(height)))
	a.region(key).unit = paletteRowH
	return p
}

// hbar draws a horizontal scrollbar of width viewport px, positioned by the
// caller (absolutely, at the bottom of a clipped area). It is built from the
// same ScrollState geometry as the vertical bars and hides when the content fits.
func (a *App) hbar(key, label string, content, viewport, offset int, set func(int)) *ui.Element {
	r := a.region(key + "-h")
	r.horizontal, r.label, r.set = true, label, set
	r.st.SetRange(content, viewport)
	r.st.ScrollTo(offset)
	left, w, vis := r.st.Metrics().Thumb(viewport)
	if !vis {
		return nil
	}
	page := func(dir int) func(context.Context) error {
		return func(ctx context.Context) error {
			if r.st.Page(dir) {
				set(r.st.Offset())
			}
			return a.redraw(ctx)
		}
	}
	bar := ui.Div().ID(key + "-hbar").Class("qr-hbar").Width(ui.Px(float64(viewport)))
	if left > 0 {
		bar.Children(ui.Button("").ID(key+"-hbar-before").Class("qr-hgap").Attr("type", "button").Attr("tabindex", "-1").
			Attr("aria-hidden", "true").Width(ui.Px(float64(left))).OnClick(page(-1)))
	}
	bar.Children(ui.Button("").ID(key+"-hthumb").Class("qr-hthumb").Attr("type", "button").Attr("role", "scrollbar").
		Attr("aria-label", label).Attr("aria-orientation", "horizontal").Width(ui.Px(float64(w))))
	if rest := viewport - left - w; rest > 0 {
		bar.Children(ui.Button("").ID(key+"-hbar-after").Class("qr-hgap").Attr("type", "button").Attr("tabindex", "-1").
			Attr("aria-hidden", "true").Width(ui.Px(float64(rest))).OnClick(page(1)))
	}
	return bar
}
