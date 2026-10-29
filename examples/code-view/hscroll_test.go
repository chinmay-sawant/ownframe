package main

import (
	"context"
	"html/template"
	"strconv"
	"strings"
	"testing"
)

func renderAppHTML(t *testing.T, app *App) string {
	t.Helper()

	tmpl, err := template.New("test").Parse(app.Page().HTML())
	if err != nil {
		t.Fatalf("parse template: %v", err)
	}

	var buf strings.Builder
	v := app.View()
	if err := tmpl.Execute(&buf, &v); err != nil {
		t.Fatalf("execute template: %v", err)
	}

	return buf.String()
}

// TestHScrollOffsetClamping verifies that horizontal scroll offset clamps to 0
// at the left edge and cannot exceed the active document's maximum line width.
func TestHScrollOffsetClamping(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	app.wrap = false
	app.scrollX = 0
	app.refresh()

	maxW := app.maxLineWidth()
	if maxW <= 0 {
		t.Fatalf("maxLineWidth = %d, want > 0", maxW)
	}

	// Cannot scroll past left edge (0).
	app.scrollHorizontal(-100)
	if app.ScrollX() != 0 {
		t.Fatalf("scrollX after scrolling left = %d, want 0", app.ScrollX())
	}

	// Scrolling right increases scrollX.
	app.scrollHorizontal(50)
	if app.ScrollX() != 50 {
		t.Fatalf("scrollX after scrolling right = %d, want 50", app.ScrollX())
	}

	// Clamping at max line width.
	app.scrollHorizontal(maxW * 2)
	if app.ScrollX() != maxW {
		t.Fatalf("scrollX = %d, want clamped to maxLineWidth %d", app.ScrollX(), maxW)
	}

	// Direct SetScrollX also clamps to [0, maxLineWidth].
	app.SetScrollX(-50)
	if app.ScrollX() != 0 {
		t.Fatalf("SetScrollX(-50) = %d, want 0", app.ScrollX())
	}

	app.SetScrollX(maxW + 500)
	if app.ScrollX() != maxW {
		t.Fatalf("SetScrollX(maxW+500) = %d, want %d", app.ScrollX(), maxW)
	}
}

// TestHScrollWheelEvents verifies that horizontal wheel events (dx != 0 or Shift+Wheel)
// adjust the horizontal scroll offset when wrap is disabled and the pointer is
// over the editor pane.
func TestHScrollWheelEvents(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()

	app.wrap = false
	app.scrollX = 0
	app.refresh()

	editorBox, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		editorBox, ok = boxByID(app.Page().Boxes(), "editor-window")
	}
	editorX := float64(chromeLeft + 50)
	editorY := float64(chromeTop + 50)
	if ok {
		editorX = editorBox.X + editorBox.W/2
		editorY = editorBox.Y + editorBox.H/2
	}

	// 1. Horizontal wheel dx < 0 (rolling right) scrolls right.
	handled, err := app.onWheel(ctx, editorX, editorY, -1, 0)
	if err != nil {
		t.Fatalf("onWheel: %v", err)
	}
	if !handled {
		t.Fatal("onWheel was not handled for horizontal wheel over editor")
	}
	if app.ScrollX() <= 0 {
		t.Fatalf("scrollX after wheel right = %d, want > 0", app.ScrollX())
	}

	prevScroll := app.ScrollX()

	// 2. Horizontal wheel dx > 0 (rolling left) scrolls left.
	handled, err = app.onWheel(ctx, editorX, editorY, 1, 0)
	if err != nil {
		t.Fatalf("onWheel: %v", err)
	}
	if !handled {
		t.Fatal("onWheel was not handled for horizontal wheel left")
	}
	if app.ScrollX() >= prevScroll {
		t.Fatalf("scrollX after wheel left = %d, want < %d", app.ScrollX(), prevScroll)
	}

	// 3. Shift + Wheel down (dy < 0 with shift) scrolls right.
	app.shift = true
	handled, err = app.onWheel(ctx, editorX, editorY, 0, -1)
	app.shift = false
	if err != nil {
		t.Fatalf("onWheel shift: %v", err)
	}
	if !handled {
		t.Fatal("onWheel was not handled for Shift+Wheel over editor")
	}
	if app.ScrollX() <= 0 {
		t.Fatalf("scrollX after Shift+Wheel = %d, want > 0", app.ScrollX())
	}

	// 4. When wrap is ON, horizontal wheel is ignored.
	app.wrap = true
	app.scrollX = 0
	app.refresh()
	handled, err = app.onWheel(ctx, editorX, editorY, -1, 0)
	if err != nil {
		t.Fatalf("onWheel when wrapped: %v", err)
	}
	if handled {
		t.Fatal("onWheel was handled for horizontal wheel when wrap is on")
	}
	if app.ScrollX() != 0 {
		t.Fatalf("scrollX with wrap=true = %d, want 0", app.ScrollX())
	}

	// 5. Wheel outside editor pane (e.g. over sidebar at x=10) does not scroll editor.
	app.wrap = false
	app.scrollX = 20
	app.refresh()
	_, _ = app.onWheel(ctx, 10, 100, -1, 0)
	if app.ScrollX() != 20 {
		t.Fatalf("scrollX after wheel outside editor = %d, want 20", app.ScrollX())
	}
}

// TestEditorVerticalWheelReachesTheWindow checks that a vertical wheel over
// the editor is left for the window, which pans a long file.
func TestEditorVerticalWheelReachesTheWindow(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()

	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("missing code body")
	}

	handled, err := app.onWheel(ctx, body.X+body.W/2, body.Y+body.H/2, 0, -1)
	if err != nil {
		t.Fatal(err)
	}

	if handled {
		t.Fatal("vertical wheel over the editor was consumed")
	}
}

// TestHScrollKeyboardActions verifies that arrow keys scroll horizontally when
// wrap is disabled and clamp properly.
func TestHScrollKeyboardActions(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()

	app.wrap = false
	app.scrollX = 0
	app.refresh()

	// ArrowRight scrolls right by hscrollStep.
	if err := app.onKeyDown(ctx, "arrowright"); err != nil {
		t.Fatalf("arrowright: %v", err)
	}
	if app.ScrollX() != hscrollStep {
		t.Fatalf("scrollX after arrowright = %d, want %d", app.ScrollX(), hscrollStep)
	}

	// ArrowLeft scrolls left by hscrollStep.
	if err := app.onKeyDown(ctx, "arrowleft"); err != nil {
		t.Fatalf("arrowleft: %v", err)
	}
	if app.ScrollX() != 0 {
		t.Fatalf("scrollX after arrowleft = %d, want 0", app.ScrollX())
	}

	// ArrowLeft at 0 stays at 0.
	if err := app.onKeyDown(ctx, "arrowleft"); err != nil {
		t.Fatalf("arrowleft at 0: %v", err)
	}
	if app.ScrollX() != 0 {
		t.Fatalf("scrollX = %d, want 0", app.ScrollX())
	}

	// With wrap ON, arrow keys do not change horizontal scroll.
	app.wrap = true
	app.refresh()
	if err := app.onKeyDown(ctx, "arrowright"); err != nil {
		t.Fatalf("arrowright with wrap: %v", err)
	}
	if app.ScrollX() != 0 {
		t.Fatalf("scrollX with wrap ON = %d, want 0", app.ScrollX())
	}
}

// TestHScrollWrapToggleResets verifies that toggling wrap back ON resets horizontal
// scroll so wrapping takes precedence.
func TestHScrollWrapToggleResets(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")

	// Start with wrap off.
	app.wrap = false
	app.scrollHorizontal(80)
	app.refresh()

	if app.ScrollX() != 80 {
		t.Fatalf("scrollX = %d, want 80", app.ScrollX())
	}
	if app.View().ScrollX != 80 {
		t.Fatalf("View().ScrollX = %d, want 80", app.View().ScrollX)
	}

	// Toggle wrap back ON.
	app.toggleWrap()

	if !app.wrap {
		t.Fatal("wrap did not toggle ON")
	}
	if app.ScrollX() != 0 {
		t.Fatalf("scrollX = %d, want reset to 0", app.ScrollX())
	}
	if app.View().ScrollX != 0 {
		t.Fatalf("View().ScrollX = %d, want 0", app.View().ScrollX)
	}

	// Calling scrollHorizontal while wrap is ON keeps scrollX at 0.
	app.scrollHorizontal(50)
	if app.ScrollX() != 0 {
		t.Fatalf("scrollHorizontal when wrap is ON = %d, want 0", app.ScrollX())
	}
}

// TestHScrollHTMLTemplateStyling verifies that HTML output includes the
// left translation style when unwrapped and scrolled, and removes it when wrapped.
func TestHScrollHTMLTemplateStyling(t *testing.T) {
	app := newTestApp(t)
	app.openFile("backend/probe.go")
	ctx := context.Background()

	// Verify the template structure contains the horizontal translation rule.
	tmpl := app.Page().HTML()
	if !strings.Contains(tmpl, `style="left: -{{.ScrollX}}px"`) {
		t.Fatalf("Template does not contain 'style=\"left: -{{.ScrollX}}px\"'")
	}

	app.wrap = false
	app.scrollHorizontal(60)
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	// Unwrapped rows paint only their visible window: the template offset
	// carries the sub-column remainder, and the full scrollX stays in state.
	cellW := app.zf(codeCellW)
	col := int(float64(60) / cellW)
	rem := 60 - int(float64(col)*cellW)
	rendered := renderAppHTML(t, app)
	if !strings.Contains(rendered, "left: -"+strconv.Itoa(rem)+"px") {
		t.Fatalf("Rendered HTML output does not contain 'left: -%dpx'", rem)
	}
	if strings.Contains(rendered, "left: -60px") {
		t.Fatal("Rendered HTML paints the full scroll offset; rows must slice to their window")
	}
	cols := app.codeColsForGroup(0)
	for _, row := range app.View().Rows {
		if got := segmentRuneLen(row.Segments); got > cols {
			t.Fatalf("row %d lays %d runes, budget %d", row.Line, got, cols)
		}
	}

	// Toggle wrap ON: horizontal translation style is removed.
	app.toggleWrap()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw: %v", err)
	}

	renderedWrapped := renderAppHTML(t, app)
	if strings.Contains(renderedWrapped, `clinetext" style="left: -`) {
		t.Fatal("Rendered HTML output still contains 'clinetext\" style=\"left: -' when wrap is ON")
	}
}

// TestHScrollDiffRows verifies horizontal scrolling for unwrapped diff rows.
func TestHScrollDiffRows(t *testing.T) {
	app := newTestApp(t)
	ctx := context.Background()

	app.setMode(ctx, "changes")
	app.wrap = false
	app.scrollX = 0
	app.refresh()

	maxW := app.maxLineWidth()
	if maxW <= 0 {
		t.Fatalf("diff maxLineWidth = %d, want > 0", maxW)
	}

	app.scrollHorizontal(40)
	if app.ScrollX() != 40 {
		t.Fatalf("diff scrollX = %d, want 40", app.ScrollX())
	}
	if app.View().Diff.ScrollX != 40 {
		t.Fatalf("View().Diff.ScrollX = %d, want 40", app.View().Diff.ScrollX)
	}

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatalf("redraw diff: %v", err)
	}

	rendered := renderAppHTML(t, app)
	cellW := app.zf(diffCellW)
	col := int(float64(40) / cellW)
	rem := 40 - int(float64(col)*cellW)
	if len(app.View().Diff.Rows) > 0 && !strings.Contains(rendered, "left: -"+strconv.Itoa(rem)+"px") {
		t.Fatalf("Rendered HTML output for diff does not contain 'left: -%dpx'", rem)
	}
	unified, _ := app.diffColumns()
	for _, row := range app.View().Diff.Rows {
		if len(row.Left) == 0 {
			continue
		}
		if got := segmentRuneLen(row.Left); got > unified {
			t.Fatalf("diff row kind %s lays %d runes, budget %d", row.Kind, got, unified)
		}
	}

	// Toggle wrap ON resets diff scroll.
	app.toggleWrap()
	if app.ScrollX() != 0 {
		t.Fatalf("diff scrollX after wrap toggle = %d, want 0", app.ScrollX())
	}
	if app.View().Diff.ScrollX != 0 {
		t.Fatalf("View().Diff.ScrollX = %d, want 0", app.View().Diff.ScrollX)
	}
}

// TestHScrollSecondaryPane verifies horizontal scrolling in the secondary editor pane.
func TestHScrollSecondaryPane(t *testing.T) {
	app := newTestApp(t)
	app.toggleSplit()
	app.openFile("backend/probe.go")

	app.wrap = false
	app.scrollHorizontal(45)
	app.refresh()

	if app.View().Second.ScrollX != 45 {
		t.Fatalf("Second.ScrollX = %d, want 45", app.View().Second.ScrollX)
	}
}
