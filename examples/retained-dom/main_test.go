package main

import (
	"context"
	"strings"
	"testing"
	"time"
)

func displayHasText(t *testing.T, app *App, want string) bool {
	t.Helper()
	for _, op := range app.Page().Probe().Display().Ops {
		if strings.Contains(op.Text, want) {
			return true
		}
	}

	return false
}

func TestRetainedDOMExampleActions(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := New(true, true)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	now := time.Date(2026, 11, 28, 11, 0, 0, 0, time.UTC)
	app.now = func() time.Time { return now }

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("initial Redraw failed: %v", err)
	}

	if !displayHasText(t, app, "OwnFrame v0.0.9") {
		t.Fatal("initial display missing title")
	}

	actions := []struct {
		name     string
		wantText string
	}{
		{"toggle-class", "ToggleClass(.active)"},
		{"flag-on", "AddClass(.flagged)"},
		{"flag-off", "RemoveClass(.flagged)"},
		{"attr-set", `SetAttr(data-priority="high")`},
		{"attr-remove", "RemoveAttr(data-priority)"},
		{"patch-text", "text updates: 1"},
		{"batch-patch", "Slot C: 1"},
		{"append-item", "#3 Appended via ElementEdit.AppendHTML"},
		{"remove-last", "Remove(#feed-item-3)"},
		{"reset-feed", "#1 Reset via ElementEdit.InnerHTML"},
	}

	for _, tc := range actions {
		if err := app.handleAction(ctx, tc.name); err != nil {
			t.Fatalf("action %q failed: %v", tc.name, err)
		}
		if !displayHasText(t, app, tc.wantText) {
			t.Fatalf("after action %q, display missing %q", tc.name, tc.wantText)
		}
	}

	if displayHasText(t, app, "#3 Appended") {
		t.Fatal("expected removed and reset item #3 to be absent from Display.Ops")
	}

	// Verify real click dispatch routes through ClickPaint without a full rebuild.
	var appendBox struct {
		x, y  float64
		found bool
	}
	for _, b := range app.Page().Boxes() {
		if b.Action == "append-item" {
			appendBox.x = b.X + b.W/2
			appendBox.y = b.Y + b.H/2
			appendBox.found = true
			break
		}
	}
	if !appendBox.found {
		t.Fatal("append-item box not found in Page.Boxes()")
	}
	if err := app.Page().Click(ctx, appendBox.x, appendBox.y); err != nil {
		t.Fatalf("Click(append-item) failed: %v", err)
	}
	if !displayHasText(t, app, "#3 Appended via ElementEdit.AppendHTML") {
		t.Fatal("Click(append-item) did not append item #3 to Display.Ops")
	}

	// Idempotent / no-op edits return ErrPatchRefused from PatchElement and must
	// surface in the bottom-right alert card instead of failing the event loop.
	for _, noop := range []string{"flag-off", "attr-remove", "flag-on", "flag-on"} {
		if err := app.handleAction(ctx, noop); err != nil {
			t.Fatalf("no-op action %q returned error: %v", noop, err)
		}
	}
	if !displayHasText(t, app, "Patch Refused (No-op)") {
		t.Fatal("expected duplicate AddClass to show Patch Refused (No-op) toast title")
	}
	if !displayHasText(t, app, "AddClass(.flagged)") {
		t.Fatal("expected duplicate AddClass to mention AddClass(.flagged) in toast body")
	}
	if !app.toastVisible {
		t.Fatal("expected toast alert card to be visible after action")
	}

	// Advancing the clock past 5 seconds must remove all toast ops from Display.Ops.
	now = now.Add(toastDuration + time.Millisecond)
	if err := app.handleTick(ctx); err != nil {
		t.Fatalf("handleTick after 5s failed: %v", err)
	}
	if app.toastVisible {
		t.Fatal("expected toast alert card to auto-dismiss after 5s")
	}
	if displayHasText(t, app, "Patch Refused (No-op)") {
		t.Fatal("expected toast text to be completely removed from Display.Ops after 5s timeout")
	}

	// Clicking the X button on the toast card must remove all toast ops immediately.
	if err := app.handleAction(ctx, "toggle-class"); err != nil {
		t.Fatalf("toggle-class failed: %v", err)
	}
	if !displayHasText(t, app, "Retained Patch Applied") {
		t.Fatal("expected toast alert card text in Display.Ops before clicking X")
	}
	var closeBox struct {
		x, y  float64
		found bool
	}
	for _, b := range app.Page().Boxes() {
		if b.Action == "dismiss-toast" {
			closeBox.x = b.X + b.W/2
			closeBox.y = b.Y + b.H/2
			closeBox.found = true
			break
		}
	}
	if !closeBox.found {
		t.Fatal("dismiss-toast X button box not found in Page.Boxes()")
	}
	if err := app.Page().Click(ctx, closeBox.x, closeBox.y); err != nil {
		t.Fatalf("Click(dismiss-toast) failed: %v", err)
	}
	if app.toastVisible {
		t.Fatal("expected clicking X button to dismiss toast card immediately")
	}
	if displayHasText(t, app, "Retained Patch Applied") {
		t.Fatal("expected clicking X button to remove toast text from Display.Ops")
	}

	// Simulate narrower viewport when the DevTools dock is open on the right;
	// fixed-size patched containers must still patch without size-change refusals.
	app.Page().SetSize(720, 760)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("narrow Redraw failed: %v", err)
	}
	for _, name := range []string{"toggle-class", "patch-text", "batch-patch", "append-item", "remove-last", "reset-feed"} {
		if err := app.handleAction(ctx, name); err != nil {
			t.Fatalf("narrow-viewport action %q failed: %v", name, err)
		}
	}
	if !displayHasText(t, app, "text updates: 2") {
		t.Fatal("narrow-viewport patch-text did not update #target-meta")
	}
}

func TestToastStaysStationaryAcrossMotionTicks(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app, err := New(true, true)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("Redraw failed: %v", err)
	}
	if err := app.handleAction(ctx, "toggle-class"); err != nil {
		t.Fatalf("toggle-class failed: %v", err)
	}

	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("first Tick failed: %v", err)
	}
	time.Sleep(150 * time.Millisecond)
	if err := app.Page().Testing().Tick(ctx); err != nil {
		t.Fatalf("second Tick failed: %v", err)
	}

	d := app.Page().Probe().Display()
	toastBoxes := map[int]string{}
	for i, b := range d.Boxes {
		if b.ID == "toast-card" || b.ID == "toast-title" || b.ID == "toast-body" {
			toastBoxes[i] = b.ID
		}
	}
	if len(toastBoxes) != 3 {
		t.Fatalf("expected 3 toast boxes, found %d: %v", len(toastBoxes), toastBoxes)
	}
	for i, op := range d.Ops {
		if i >= len(d.OpOwners) {
			continue
		}
		if id, ok := toastBoxes[d.OpOwners[i]]; ok {
			m := op.Transform()
			if m.E != 0 || m.F != 0 {
				t.Errorf("op %d (owner %s, kind %v, text %q) unexpectedly has Transform=%+v", i, id, op.Kind, op.Text, m)
			}
		}
	}
}
