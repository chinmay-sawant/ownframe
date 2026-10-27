package main

import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"strings"
	"testing"
)

func TestScrollRailsPaint(t *testing.T) {
	if html := buildHTML(); !strings.Contains(html, ".list-scroll") || !strings.Contains(html, ".hist-graph") {
		t.Fatal("scrollrail stylesheet is not in the built page")
	}

	app := newTestApp(t)
	app.mode = "changes"
	app.changes = changeSet{Branch: "test", Note: "test"}

	for i := 0; i < 30; i++ {
		app.changes.Files = append(app.changes.Files, changedFile{Path: fmt.Sprintf("file-%02d.go", i), Status: "M"})
	}

	app.statuses = statusMap(app.changes)
	app.selectFirstDiff()

	app.commits.Commits = nil
	for i := 0; i < 30; i++ {
		app.commits.Commits = append(app.commits.Commits, gitCommit{Hash: fmt.Sprintf("%040d", i), Short: fmt.Sprintf("%07d", i)})
	}

	app.changesScroll.SetOffset(0, len(app.changes.Files), app.changeVisibleRows())
	app.historyScroll.SetOffset(0, len(app.commits.Commits), app.historyVisibleRows())
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if _, ok := boxByID(app.Page().Boxes(), "changes-scroll"); !ok {
		t.Fatalf("no changes scroll rail: scroll=%v files=%d visible=%d", app.View().Diff.Scroll, len(app.changes.Files), app.changeVisibleRows())
	}

	if _, ok := boxByID(app.Page().Boxes(), "changes-thumb"); !ok {
		t.Fatal("no changes scroll thumb")
	}

	if _, ok := boxByID(app.Page().Boxes(), "history-scroll"); !ok {
		t.Fatalf("no history scroll rail: scroll=%v commits=%d visible=%d", app.View().History.Scroll, len(app.commits.Commits), app.historyVisibleRows())
	}

	if _, ok := boxByID(app.Page().Boxes(), "history-thumb"); !ok {
		t.Fatal("no history scroll thumb")
	}
}

// TestPaneScrollThumbsPaint checks the working-tree and history thumbs are
// the light gray bar, and a long history subject stays on two whole words
// clear of that bar. The hover card's lines fill the card.
func TestPaneScrollThumbsPaint(t *testing.T) {
	app := newScrollAccessibilityApp(t, 48, 48)
	app.commits.Commits[0].Short = "c000001"
	app.commits.Commits[0].Subject = "Fix the history subject wrapping so a long commit message uses the full row width"
	app.commits.Commits[0].Body = "The hover card should wrap on word boundaries and sit against the text instead of leaving a wide empty gutter on the right of the card."
	app.historyScroll.SetOffset(0, len(app.commits.Commits), app.historyVisibleRows())
	app.hoverCommit = "c000001"
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	img, err := png.Decode(bytes.NewReader(app.Page().Probe().PNG()))
	if err != nil {
		t.Fatal(err)
	}

	boxes := app.Page().Boxes()
	for _, id := range []string{"changes-thumb", "history-thumb"} {
		thumb, ok := boxByID(boxes, id)
		if !ok {
			t.Fatalf("missing %s", id)
		}
		r, g, b, _ := img.At(int(thumb.X+thumb.W/2), int(thumb.Y+thumb.H/2)).RGBA()
		if r>>8 < 140 || g>>8 < 140 || b>>8 < 140 {
			t.Fatalf("%s pixel = %d,%d,%d, want the light thumb", id, r>>8, g>>8, b>>8)
		}
	}

	thumb, _ := boxByID(boxes, "history-thumb")
	display := app.Page().Probe().Display()
	if display == nil {
		t.Fatal("no display")
	}
	scale := display.PixelPerPoint
	if scale == 0 {
		scale = 96.0 / 72.0
	}

	var subjectLines int
	var widestTip float64
	for _, op := range display.Ops {
		if op.Text == "" {
			continue
		}
		x := op.X * scale
		right := (op.X + op.W) * scale
		if x < thumb.X && (strings.HasPrefix(op.Text, "Fix the history") || strings.HasPrefix(op.Text, "so a long")) {
			subjectLines++
			if right > thumb.X-1 {
				t.Fatalf("subject %q ends at %.1f, thumb starts at %.1f", op.Text, right, thumb.X)
			}
		}
		if strings.Contains(op.Text, "word boundaries") {
			if w := op.W * scale; w > widestTip {
				widestTip = w
			}
		}
	}
	if subjectLines != 2 {
		t.Fatalf("subject lines = %d, want 2", subjectLines)
	}
	if widestTip < 250 {
		t.Fatalf("hover line width = %.1f, want it to fill the card", widestTip)
	}
}
