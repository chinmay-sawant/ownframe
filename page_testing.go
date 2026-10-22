package ownframe

import (
	"context"
	"image"
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
