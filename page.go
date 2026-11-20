package ownframe

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"time"
)

// SetData stores the value the template prints on the next Redraw.
func (p *Page) SetData(data any) { innerPage(p).SetData(data) }

// Handle registers the input functions. A later call replaces the whole set.
func (p *Page) Handle(h Handlers) { innerPage(p).Handle(h) }

// AddHandlers merges only non-nil callbacks into the existing set.
func (p *Page) AddHandlers(h Handlers) { innerPage(p).AddHandlers(h) }

// SetTheme stores an extra stylesheet the next Redraw applies.
func (p *Page) SetTheme(source string) error {
	return innerPage(p).SetTheme(source)
}

// SetTick registers fn to run once per frame before the window draws.
func (p *Page) SetTick(fn func(ctx context.Context) error) {
	innerPage(p).SetTick(fn)
}

// SetImage registers encoded image bytes for one source name.
func (p *Page) SetImage(src string, data []byte) {
	innerPage(p).SetImage(src, data)
}

// SetAudio stores one encoded clip for the web preview.
func (p *Page) SetAudio(data []byte, contentType string) {
	innerPage(p).SetAudio(data, contentType)
}

// SetAudioPlayback publishes whether the preview should play and at what
// volume, from 0 to 1. The browser player follows it.
func (p *Page) SetAudioPlayback(playing bool, volume float64) {
	innerPage(p).SetAudioPlayback(playing, volume)
}

// Redraw renders the current size.
func (p *Page) Redraw(ctx context.Context) error {
	return innerPage(p).Redraw(ctx)
}

// RedrawIfChanged renders unless the last committed frame is unchanged.
func (p *Page) RedrawIfChanged(ctx context.Context) error {
	return innerPage(p).RedrawIfChanged(ctx)
}

// Click hit-tests the last picture and applies the click.
func (page *Page) Click(ctx context.Context, x, y float64) error {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Click(ctx, x, y)
}

// WantFrameAt declares the next time the page needs the window to draw.
func (page *Page) WantFrameAt(at time.Time) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.WantFrameAt(at)
}

// WantFrameIn declares the next frame in d.
func (page *Page) WantFrameIn(d time.Duration) {
	p := innerPage(page)
	if p == nil {
		return
	}
	p.WantFrameIn(d)
}

// Title returns the window title.
func (page *Page) Title() string {
	p := innerPage(page)
	if p == nil {
		return ""
	}

	return p.Title()
}

// Boxes returns a copy of the last hit-test boxes.
func (page *Page) Boxes() []Box {
	p := innerPage(page)
	if p == nil {
		return nil
	}

	return p.Boxes()
}

// Size returns the frame size in CSS pixels.
func (p *Page) Size() (int, int) { return innerPage(p).Size() }

// MinSize returns the configured desktop minimum.
func (p *Page) MinSize() (int, int) { return innerPage(p).MinSize() }

// MaxSize returns the largest window the page asks its host for.
func (p *Page) MaxSize() (int, int) { return innerPage(p).MaxSize() }

// Clamp applies the desktop minimum or the native mobile viewport.
func (p *Page) Clamp(width, height int) (int, int) {
	return innerPage(p).Clamp(width, height)
}

// SetSize stores the frame size used by the next Redraw.
func (p *Page) SetSize(width, height int) {
	innerPage(p).SetSize(width, height)
}

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

// ElementEdit is one change to an element in the retained tree: an
// attribute to set, or the text to replace. A nil Text and an empty
// Attribute change nothing.
type ElementEdit = page.ElementEdit
