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
