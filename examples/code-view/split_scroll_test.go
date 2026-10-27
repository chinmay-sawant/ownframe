package main

import (
	"context"
	"testing"
)

func TestSplitScrollIndependentOffsets(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	// Open split editor
	app.toggleSplit()
	if !app.split {
		t.Fatal("expected split mode to be active")
	}

	app.openFileIn(0, "backend/server.go")
	app.openFileIn(1, "backend/store.go")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	if app.scrollGroup[0] != 0 || app.scrollGroup[1] != 0 {
		t.Fatalf("initial offsets: group0=%d, group1=%d, want (0, 0)", app.scrollGroup[0], app.scrollGroup[1])
	}

	// Scroll Group 1 down by 160 px (8 rows)
	app.scrollGroupBy(1, 160)
	if app.scrollGroup[1] != 160 {
		t.Fatalf("group 1 scroll offset = %d, want 160", app.scrollGroup[1])
	}
	if app.scrollGroup[0] != 0 {
		t.Fatalf("group 0 scroll offset modified: got %d, want 0", app.scrollGroup[0])
	}

	// Slicing check: Group 0 starts at row 1, Group 1 starts at row 9 (160/20 = 8 skipped)
	v := app.View()
	if len(v.Rows) == 0 || v.Rows[0].No != 1 {
		t.Fatalf("group 0 visible first row = %v, want 1", v.Rows[0].No)
	}
	if len(v.Second.Rows) == 0 || v.Second.Rows[0].No != 9 {
		t.Fatalf("group 1 visible first row = %v, want 9", v.Second.Rows[0].No)
	}

	// Scroll Group 0 down by 100 px (5 rows)
	app.scrollGroupBy(0, 100)
	if app.scrollGroup[0] != 100 {
		t.Fatalf("group 0 scroll offset = %d, want 100", app.scrollGroup[0])
	}
	if app.scrollGroup[1] != 160 {
		t.Fatalf("group 1 scroll offset modified: got %d, want 160", app.scrollGroup[1])
	}

	// Verify new slicing: Group 0 starts at row 6 (100/20 = 5 skipped), Group 1 still at row 9
	v = app.View()
	if len(v.Rows) == 0 || v.Rows[0].No != 6 {
		t.Fatalf("group 0 visible first row = %v, want 6", v.Rows[0].No)
	}
	if len(v.Second.Rows) == 0 || v.Second.Rows[0].No != 9 {
		t.Fatalf("group 1 visible first row = %v, want 9", v.Second.Rows[0].No)
	}
}

func TestSplitScrollWheelRouting(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	app.toggleSplit()
	app.openFileIn(0, "backend/server.go")
	app.openFileIn(1, "backend/store.go")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	boxes := app.Page().Boxes()
	box1, ok1 := boxByID(boxes, "editor-window")
	box2, ok2 := boxByID(boxes, "editor-window-2")

	if !ok1 {
		t.Fatal("box editor-window not found")
	}
	if !ok2 {
		t.Fatal("box editor-window-2 not found")
	}

	// Initial offsets are zero
	if app.scrollGroup[0] != 0 || app.scrollGroup[1] != 0 {
		t.Fatalf("expected initial (0, 0), got (%d, %d)", app.scrollGroup[0], app.scrollGroup[1])
	}

	// Wheel scroll over editor-window-2 (Group 1) with dy = -1 (scrolling down)
	handled, err := app.onWheel(ctx, box2.X+box2.W/2, box2.Y+box2.H/2, 0, -1)
	if err != nil {
		t.Fatalf("onWheel error: %v", err)
	}
	if !handled {
		t.Fatal("expected onWheel over editor-window-2 to be handled")
	}

	if app.scrollGroup[1] != scrollStep {
		t.Fatalf("group 1 offset after wheel down = %d, want %d", app.scrollGroup[1], scrollStep)
	}
	if app.scrollGroup[0] != 0 {
		t.Fatalf("group 0 offset changed during group 1 wheel: %d", app.scrollGroup[0])
	}

	// Wheel scroll over editor-window (Group 0) with dy = -1 (scrolling down)
	handled, err = app.onWheel(ctx, box1.X+box1.W/2, box1.Y+box1.H/2, 0, -1)
	if err != nil {
		t.Fatalf("onWheel error: %v", err)
	}
	if !handled {
		t.Fatal("expected onWheel over editor-window to be handled")
	}

	if app.scrollGroup[0] != scrollStep {
		t.Fatalf("group 0 offset after wheel down = %d, want %d", app.scrollGroup[0], scrollStep)
	}
	if app.scrollGroup[1] != scrollStep {
		t.Fatalf("group 1 offset changed during group 0 wheel: %d", app.scrollGroup[1])
	}

	// Wheel scroll up over editor-window-2 (Group 1) with dy = +1
	handled, err = app.onWheel(ctx, box2.X+box2.W/2, box2.Y+box2.H/2, 0, 1)
	if err != nil {
		t.Fatalf("onWheel error: %v", err)
	}
	if !handled {
		t.Fatal("expected onWheel up to be handled")
	}

	if app.scrollGroup[1] != 0 {
		t.Fatalf("group 1 offset after wheel up = %d, want 0", app.scrollGroup[1])
	}
	if app.scrollGroup[0] != scrollStep {
		t.Fatalf("group 0 offset changed during group 1 wheel up: %d", app.scrollGroup[0])
	}
}

func TestSplitScrollKeyboardRouting(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	app.toggleSplit()
	app.openFileIn(0, "backend/server.go")
	app.openFileIn(1, "backend/store.go")

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	// Focus Group 0
	app.setFocusGroup(0)
	if app.focusGroup != 0 {
		t.Fatalf("expected focusGroup 0, got %d", app.focusGroup)
	}

	// Press arrowdown: should scroll Group 0, leaving Group 1 at 0
	if err := app.onKeyDown(ctx, "arrowdown"); err != nil {
		t.Fatalf("onKeyDown arrowdown: %v", err)
	}
	if app.scrollGroup[0] != scrollStep {
		t.Fatalf("group 0 offset = %d, want %d", app.scrollGroup[0], scrollStep)
	}
	if app.scrollGroup[1] != 0 {
		t.Fatalf("group 1 offset changed: %d", app.scrollGroup[1])
	}

	// Press pagedown: should scroll Group 0 further
	if err := app.onKeyDown(ctx, "pagedown"); err != nil {
		t.Fatalf("onKeyDown pagedown: %v", err)
	}
	expectedG0 := scrollStep + pageScroll
	if app.scrollGroup[0] != expectedG0 {
		t.Fatalf("group 0 offset = %d, want %d", app.scrollGroup[0], expectedG0)
	}
	if app.scrollGroup[1] != 0 {
		t.Fatalf("group 1 offset changed: %d", app.scrollGroup[1])
	}

	// Press arrowup: should decrease Group 0 offset
	if err := app.onKeyDown(ctx, "arrowup"); err != nil {
		t.Fatalf("onKeyDown arrowup: %v", err)
	}
	if app.scrollGroup[0] != expectedG0-scrollStep {
		t.Fatalf("group 0 offset = %d, want %d", app.scrollGroup[0], expectedG0-scrollStep)
	}

	// Press home: should reset Group 0 to 0
	if err := app.onKeyDown(ctx, "home"); err != nil {
		t.Fatalf("onKeyDown home: %v", err)
	}
	if app.scrollGroup[0] != 0 {
		t.Fatalf("group 0 offset after home = %d, want 0", app.scrollGroup[0])
	}

	// Now switch focus to Group 1
	app.setFocusGroup(1)
	if app.focusGroup != 1 {
		t.Fatalf("expected focusGroup 1, got %d", app.focusGroup)
	}

	// Press arrowdown: should scroll Group 1, leaving Group 0 at 0
	if err := app.onKeyDown(ctx, "arrowdown"); err != nil {
		t.Fatalf("onKeyDown arrowdown: %v", err)
	}
	if app.scrollGroup[1] != scrollStep {
		t.Fatalf("group 1 offset = %d, want %d", app.scrollGroup[1], scrollStep)
	}
	if app.scrollGroup[0] != 0 {
		t.Fatalf("group 0 offset changed during group 1 key scroll: %d", app.scrollGroup[0])
	}

	// Press pagedown: should scroll Group 1 further
	if err := app.onKeyDown(ctx, "pagedown"); err != nil {
		t.Fatalf("onKeyDown pagedown: %v", err)
	}
	expectedG1 := scrollStep + pageScroll
	if app.scrollGroup[1] != expectedG1 {
		t.Fatalf("group 1 offset = %d, want %d", app.scrollGroup[1], expectedG1)
	}
	if app.scrollGroup[0] != 0 {
		t.Fatalf("group 0 offset changed: %d", app.scrollGroup[0])
	}

	// Press pageup: should decrease Group 1 offset
	if err := app.onKeyDown(ctx, "pageup"); err != nil {
		t.Fatalf("onKeyDown pageup: %v", err)
	}
	if app.scrollGroup[1] != scrollStep {
		t.Fatalf("group 1 offset after pageup = %d, want %d", app.scrollGroup[1], scrollStep)
	}

	// Press end: should scroll Group 1 to bottom
	if err := app.onKeyDown(ctx, "end"); err != nil {
		t.Fatalf("onKeyDown end: %v", err)
	}
	maxG1 := app.maxGroupScroll(1)
	if app.scrollGroup[1] != maxG1 {
		t.Fatalf("group 1 offset after end = %d, want %d", app.scrollGroup[1], maxG1)
	}
	if app.scrollGroup[0] != 0 {
		t.Fatalf("group 0 offset changed: %d", app.scrollGroup[0])
	}
}

func TestSplitScrollClamping(t *testing.T) {
	app := newTestApp(t)

	app.toggleSplit()
	app.openFileIn(0, "backend/server.go")
	app.openFileIn(1, "README.md") // 52 lines

	// Upward scroll should clamp to 0
	app.scrollGroupBy(0, -500)
	if app.scrollGroup[0] != 0 {
		t.Fatalf("group 0 clamp at top: got %d, want 0", app.scrollGroup[0])
	}
	app.scrollGroupBy(1, -500)
	if app.scrollGroup[1] != 0 {
		t.Fatalf("group 1 clamp at top: got %d, want 0", app.scrollGroup[1])
	}

	// Downward scroll past bottom should clamp to maxGroupScroll
	app.scrollGroupBy(0, 100000)
	max0 := app.maxGroupScroll(0)
	if app.scrollGroup[0] != max0 {
		t.Fatalf("group 0 clamp at bottom: got %d, want %d", app.scrollGroup[0], max0)
	}

	app.scrollGroupBy(1, 100000)
	max1 := app.maxGroupScroll(1)
	if app.scrollGroup[1] != max1 {
		t.Fatalf("group 1 clamp at bottom: got %d, want %d", app.scrollGroup[1], max1)
	}
}

func TestSplitScrollToggleReset(t *testing.T) {
	app := newTestApp(t)

	app.toggleSplit()
	app.openFileIn(0, "backend/server.go")
	app.openFileIn(1, "backend/store.go")

	app.scrollGroupBy(0, 100)
	app.scrollGroupBy(1, 150)

	if app.scrollGroup[0] != 100 || app.scrollGroup[1] != 150 {
		t.Fatalf("setup scroll: (%d, %d)", app.scrollGroup[0], app.scrollGroup[1])
	}

	// Toggle split off: should reset both scroll offsets to 0
	app.toggleSplit()
	if app.split {
		t.Fatal("expected split to be closed")
	}

	if app.scrollGroup[0] != 0 || app.scrollGroup[1] != 0 {
		t.Fatalf("after split closed: group0=%d, group1=%d, want (0, 0)", app.scrollGroup[0], app.scrollGroup[1])
	}
}

func TestSplitScrollTabSwitchPreservesOtherPane(t *testing.T) {
	app := newTestApp(t)

	app.toggleSplit()
	app.openFileIn(0, "backend/server.go")
	app.openFileIn(1, "backend/store.go")
	app.openFileIn(1, "README.md") // second tab in group 1

	app.scrollGroupBy(0, 120)
	app.scrollGroupBy(1, 80)

	if app.scrollGroup[0] != 120 || app.scrollGroup[1] != 80 {
		t.Fatalf("setup offsets: (%d, %d)", app.scrollGroup[0], app.scrollGroup[1])
	}

	// Switch tab in group 1 to tab 0 ("backend/store.go")
	app.selectTabIn(1, 0)

	// Group 1 scroll should reset to 0 for the newly focused tab
	if app.scrollGroup[1] != 0 {
		t.Fatalf("group 1 scroll after tab switch = %d, want 0", app.scrollGroup[1])
	}
	// Group 0 scroll should be preserved!
	if app.scrollGroup[0] != 120 {
		t.Fatalf("group 0 scroll unexpectedly modified: got %d, want 120", app.scrollGroup[0])
	}
}
