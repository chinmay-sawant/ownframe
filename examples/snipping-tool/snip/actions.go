package snip

import (
	"context"

	"github.com/chinmay-sawant/ownframe"
)

// onClick routes a toolbar button to Handle. The library redraws after.
func (a *App) onClick(ctx context.Context, box ownframe.Box) error {
	action := box.Action
	if action == "" {
		action = box.ID
	}
	return a.Handle(ctx, action)
}

// onChange re-crops when an x/y/w/h field is edited. Other fields are
// ignored so future controls stay harmless.
func (a *App) onChange(_ context.Context, box ownframe.Box) error {
	switch box.ID {
	case "x", "y", "w", "h":
		a.selectFields()
	}
	return nil
}

// Handle runs one action: capture, select, clear, save, or copy.
// Unknown actions are ignored so new buttons stay harmless.
func (a *App) Handle(ctx context.Context, action string) error {
	switch action {
	case "capture":
		a.capture(ctx)
	case "select":
		a.selectFields()
	case "clear":
		a.clear()
	case "save":
		a.save()
	case "copy":
		a.copy()
	}
	return nil
}

// capture takes a shot, or the demo image when no tool runs.
func (a *App) capture(ctx context.Context) {
	raw, err := CaptureWith(ctx, a.capturer)
	a.demo = false
	if err != nil || len(raw) == 0 {
		raw = DemoImage(320, 200)
		a.demo = true
	}
	a.setShot(raw)
	if a.demo {
		a.view.Status = "demo image; capture unavailable"
	} else {
		a.view.Status = "captured " + a.view.Size
	}
}
