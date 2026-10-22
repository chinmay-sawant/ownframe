package ownframe

import (
	"context"
	"time"
)

// PNGWithContext threads ctx to the on-demand render.
func (t Testing) PNGWithContext(ctx context.Context) []byte {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.PNGWithContext(ctx)
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

// TakeChangedProperties returns a snapshot and drains local metadata.
func (t Testing) TakeChangedProperties() []string {
	p := innerPage(t.page)
	if p == nil {
		return nil
	}

	return p.TakeChangedProperties()
}

// TakeScroll returns the pending scroll request and clears it.
func (t Testing) TakeScroll() (ScrollRequest, bool) {
	p := innerPage(t.page)
	if p == nil {
		return ScrollRequest{}, false
	}
	req, ok := p.TakeScroll()
	return ScrollRequest{X: req.X, Y: req.Y, Absolute: req.Absolute}, ok
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
