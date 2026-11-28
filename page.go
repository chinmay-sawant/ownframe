package ownframe

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"time"
)

// SetData stores the value the template prints on the next Redraw.
func (p *Page) SetData(data any) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.SetData(data)
}

// Handle registers the input functions. A later call replaces the whole set.
func (p *Page) Handle(h Handlers) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.Handle(h)
}

// AddHandlers merges only non-nil callbacks into the existing set.
func (p *Page) AddHandlers(h Handlers) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.AddHandlers(h)
}

// SetTheme stores an extra stylesheet the next Redraw applies.
func (p *Page) SetTheme(source string) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.SetTheme(source)
}

// SetTick registers fn to run once per frame before the window draws.
func (p *Page) SetTick(fn func(ctx context.Context) error) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.SetTick(fn)
}

// SetImage registers encoded image bytes for one source name.
func (p *Page) SetImage(src string, data []byte) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.SetImage(src, data)
}

// SetAudio stores one encoded clip for the web preview.
func (p *Page) SetAudio(data []byte, contentType string) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.SetAudio(data, contentType)
}

// SetAudioPlayback publishes whether the preview should play and at what
// volume, from 0 to 1. The browser player follows it.
func (p *Page) SetAudioPlayback(playing bool, volume float64) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.SetAudioPlayback(playing, volume)
}

// Redraw renders the current size.
func (p *Page) Redraw(ctx context.Context) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.Redraw(ctx)
}

// RedrawIfChanged renders unless the last committed frame is unchanged.
func (p *Page) RedrawIfChanged(ctx context.Context) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.RedrawIfChanged(ctx)
}

// Click hit-tests the last picture and applies the click.
func (p *Page) Click(ctx context.Context, x, y float64) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.Click(ctx, x, y)
}

// WantFrameAt declares the next time the page needs the window to draw.
func (p *Page) WantFrameAt(at time.Time) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.WantFrameAt(at)
}

// WantFrameIn declares the next frame in d.
func (p *Page) WantFrameIn(d time.Duration) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.WantFrameIn(d)
}

// Title returns the window title.
func (p *Page) Title() string {
	inner := innerPage(p)
	if inner == nil {
		return ""
	}

	return inner.Title()
}

// Boxes returns a copy of the last hit-test boxes.
func (p *Page) Boxes() []Box {
	inner := innerPage(p)
	if inner == nil {
		return nil
	}

	return inner.Boxes()
}

// BoxByID returns the first placed box with id without cloning all boxes.
func (p *Page) BoxByID(id string) (Box, bool) {
	inner := innerPage(p)
	if inner == nil {
		return Box{}, false
	}

	return inner.BoxByID(id)
}

// Size returns the frame size in CSS pixels.
func (p *Page) Size() (int, int) {
	inner := innerPage(p)
	if inner == nil {
		return 0, 0
	}

	return inner.Size()
}

// MinSize returns the configured desktop minimum.
func (p *Page) MinSize() (int, int) {
	inner := innerPage(p)
	if inner == nil {
		return 0, 0
	}

	return inner.MinSize()
}

// MaxSize returns the largest window the page asks its host for.
func (p *Page) MaxSize() (int, int) {
	inner := innerPage(p)
	if inner == nil {
		return 0, 0
	}

	return inner.MaxSize()
}

// Clamp applies the desktop minimum or the native mobile viewport.
func (p *Page) Clamp(width, height int) (int, int) {
	inner := innerPage(p)
	if inner == nil {
		return 0, 0
	}

	return inner.Clamp(width, height)
}

// SetSize stores the frame size used by the next Redraw.
func (p *Page) SetSize(width, height int) {
	inner := innerPage(p)
	if inner == nil {
		return
	}
	inner.SetSize(width, height)
}

// ElementPatch is one element's share of a batched patch: the element id
// plus the tree edits for it.
type ElementPatch = page.ElementPatch

// PatchElement re-lays out one retained element after editing it.
func (p *Page) PatchElement(ctx context.Context, id string, edits ...ElementEdit) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.PatchElement(ctx, id, edits...)
}

// PatchElements edits several retained elements in one batch with a single
// shared style re-resolve. An empty batch returns nil.
func (p *Page) PatchElements(ctx context.Context, patches []ElementPatch) error {
	inner := innerPage(p)
	if inner == nil {
		return ErrNilPage
	}

	return inner.PatchElements(ctx, patches)
}

// ElementEdit is one change to an element in the retained tree: text, an
// attribute write or removal, a class-list mutation, an element removal, or
// an HTML fragment insertion/replacement. Leaving all action fields empty/nil
// changes nothing.
type ElementEdit = page.ElementEdit
