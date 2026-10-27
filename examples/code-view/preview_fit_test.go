package main

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

// previewFitParagraph is long enough to fill the visible preview at the
// narrowest window, so every checked row is a real wrapped line.
const previewFitParagraph = "Write screens in HTML and CSS, and application logic in Go. " +
	"ownframe renders the page with a small engine and opens a window with Ebiten. " +
	"It ships no Chromium, WebKit, or JavaScript runtime. " +
	"The preview wraps this paragraph across several rows so a full line reaches " +
	"close to the right padding before the next word starts on the following row. "

// previewFitHeading is longer than the heading budget at every test width, so
// the document has to narrow to the heading face before it renders.
const previewFitHeading = "# A heading that is deliberately long enough to wrap across more than one row " +
	"because it keeps going with many words and never stops until it is well past " +
	"the width of the pane. It still has more words to force a second visual row here."

// TestPreviewRowsFitPaneAtCommonWidths renders the app at each width a window
// may take and checks every visible preview row ends inside the preview text
// area. The budget keeps the marker padding and one preview cell clear of the
// right edge, so the painted ink and the display text ops must both stop short
// of that line.
func TestPreviewRowsFitPaneAtCommonWidths(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Code review\n\n"+strings.Repeat(previewFitParagraph, 8)+"\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("README.md")
	app.togglePreview()

	for _, width := range []int{840, 1160, 1280, 1500, 1920} {
		t.Run(fmt.Sprintf("width-%d", width), func(t *testing.T) {
			assertPreviewRowsFit(t, app, width)
		})
	}
}

// TestPreviewLongHeadingFitsPane checks the same right edge for a document
// whose heading uses the larger face. The heading budget is narrower than the
// body budget, so the preview has to switch to it for the whole document.
func TestPreviewLongHeadingFitsPane(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "README.md", previewFitHeading+"\n\n"+strings.Repeat(previewFitParagraph, 4)+"\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("README.md")
	app.togglePreview()

	for _, width := range []int{840, 1160, 1280, 1500, 1920} {
		t.Run(fmt.Sprintf("width-%d", width), func(t *testing.T) {
			assertPreviewRowsFit(t, app, width)
		})
	}
}

// assertPreviewRowsFit renders one width and checks every preview row against
// the text area right edge with both the display list and the painted pixels.
func assertPreviewRowsFit(t *testing.T, app *App, width int) {
	t.Helper()

	app.Page().SetSize(width, 900)
	app.refresh()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if !view.Preview || len(view.PreviewRows) == 0 {
		t.Fatalf("preview = %v rows = %d", view.Preview, len(view.PreviewRows))
	}

	if view.TrackTop != 0 {
		t.Fatalf("preview starts at track %d, want 0", view.TrackTop)
	}

	box, ok := boxByID(app.Page().Boxes(), "markdown-preview")
	if !ok {
		t.Fatal("missing markdown preview")
	}

	display := app.Page().Probe().Display()
	if display == nil {
		t.Fatal("page fell back to a bitmap")
	}

	edges := previewTextEdges(display, box, view.TrackTop, len(view.PreviewRows))
	textAreaRight := box.X + box.W - previewPadX
	limit := textAreaRight - previewMarkPad - previewCellW
	widest := 0.0

	for _, edge := range edges {
		if edge > widest {
			widest = edge
		}
	}

	t.Logf("preview cols=%d widest row=%.1f limit=%.1f", app.previewCols(), widest, limit)

	for i, row := range view.PreviewRows {
		text := lineText(row.Segments)
		if strings.TrimSpace(text) == "" {
			continue
		}

		if edges[i] == 0 {
			t.Fatalf("preview row %d has no text op: %q", i, text)
		}

		if edges[i] > limit {
			t.Fatalf("preview row %d ends at %.1f, text area limit %.1f: %q", i, edges[i], limit, text)
		}
	}

	img := mustDecodePNG(t, app.Page().Probe().PNG())

	for i, row := range view.PreviewRows {
		text := lineText(row.Segments)
		if strings.TrimSpace(text) == "" {
			continue
		}

		y := int(box.Y) + view.TrackTop + i*rowH + rowH/2
		ink := rightInk(img, int(box.X), int(box.X+box.W)-paneScrollW, y)
		if ink < 0 {
			continue
		}

		if float64(ink) > limit {
			t.Fatalf("preview row %d paints ink to %d, text area limit %.1f: %q", i, ink, limit, text)
		}
	}
}

// TestPreviewCellWCoversMeasuredAdvance pins previewCellW above the advance the
// engine actually lays out for a long monospace run. The budget divides by this
// constant, so a face or font-size change that widens the advance fails here
// before it can clip a row.
func TestPreviewCellWCoversMeasuredAdvance(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFile(t, dir, "README.md", "# Code review\n\n"+strings.Repeat("The quick brown fox jumps over the lazy dog. ", 40)+"\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("README.md")
	app.Page().SetSize(1500, 900)
	app.togglePreview()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	box, ok := boxByID(app.Page().Boxes(), "markdown-preview")
	if !ok {
		t.Fatal("missing markdown preview")
	}

	display := app.Page().Probe().Display()
	if display == nil {
		t.Fatal("page fell back to a bitmap")
	}

	pp := display.PointsPerPixel
	if pp <= 0 {
		pp = 1
	}

	textLeft := box.X + previewPadX
	widest := 0.0

	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind != ownframe.DisplayOpText || op.Text == "" {
			continue
		}

		x := op.X / pp
		runes := len([]rune(op.Text))
		if x < textLeft-1 || runes < 20 {
			continue
		}

		if advance := op.W / pp / float64(runes); advance > widest {
			widest = advance
		}
	}

	if widest == 0 {
		t.Fatal("preview painted no long text run")
	}

	if widest > previewCellW {
		t.Fatalf("measured preview advance %.3f px exceeds budget cell %.3f px", widest, previewCellW)
	}

	t.Logf("measured preview advance %.3f px, budget cell %.3f px", widest, previewCellW)
}

// previewTextEdges returns the right edge of the rightmost text operation in
// each visible preview row. Operations are matched to a row by their baseline.
func previewTextEdges(display *ownframe.Display, box ownframe.Box, trackTop, rows int) []float64 {
	edges := make([]float64, rows)

	pp := display.PointsPerPixel
	if pp <= 0 {
		pp = 1
	}

	textLeft := box.X + previewPadX
	top := box.Y + float64(trackTop)
	bottom := top + float64(rows*rowH)

	for i := range display.Ops {
		op := &display.Ops[i]
		if op.Kind != ownframe.DisplayOpText || op.Text == "" {
			continue
		}

		x, y := op.X/pp, op.Y/pp
		if x < textLeft-1 || y < top || y >= bottom {
			continue
		}

		row := int((y - top) / rowH)
		right := (op.X + op.W) / pp
		if right > edges[row] {
			edges[row] = right
		}
	}

	return edges
}
