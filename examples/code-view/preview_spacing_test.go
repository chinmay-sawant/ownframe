package main

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"strings"
	"testing"
)

func TestLongMarkdownLineKeepsItsEndInSourceAndPreview(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	line := strings.Repeat("ordinary markdown words ", 80) + "FINAL-CONTENT"
	writeFile(t, dir, "README.md", line+"\n")
	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}
	app.openFile("README.md")
	doc := app.activeDoc()
	if doc == nil || len(doc.Lines) != 1 || doc.Lines[0] != line {
		t.Fatal("source line lost content")
	}
	app.togglePreview()
	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if got := previewText(doc.Preview); !strings.Contains(got, "FINAL-CONTENT") {
		t.Fatal("preview lost the end of the line")
	}
}

// previewSpacingSource is a narrow-width document that mixes the three blocks
// the preview reflows differently: a Markdown table with long cells, a prose
// paragraph past the column budget, and an HTML block holding a heading, a
// list, and a table.
const previewSpacingSource = `# Preview spacing at forty columns

| Component | What happens when the editor gets narrow |
| --- | --- |
| alpha | A table cell that is much longer than the forty column budget, so it wraps onto several preview rows. |
| beta | Another long table cell with enough words that its wrapped rows need a gap before the next logical row. |

A paragraph that runs past the forty column budget so the preview has to wrap it over several visual rows without letting one row escape the width. It keeps going with the long token previewspacingregressiontoken to force a hard break inside a word.

<div class="notice">
<h2>HTML block heading that is long enough to wrap at a narrow width</h2>
<p>A paragraph inside the HTML block that is also longer than forty columns.</p>
<ul>
<li>A list item inside the HTML block long enough to wrap over more than one row.</li>
<li>A second list item that wraps as well, with enough words to fill two rows.</li>
</ul>
<table>
<tr><th>Key</th><th>Detail column heading that is definitely too long for a fitted column</th></tr>
<tr><td>gamma</td><td>A raw HTML table cell long enough to wrap over several rows inside the block.</td></tr>
</table>
</div>
`

// TestPreviewSpacingNarrow checks that a narrow preview keeps one line per row
// and never lets a table or an HTML block spill past the column budget.
func TestPreviewSpacingNarrow(t *testing.T) {
	const cols = 40

	rows, _ := previewRows(strings.Split(previewSpacingSource, "\n"), cols)

	longest := 0

	for i, row := range rows {
		for _, seg := range row.Segments {
			if strings.ContainsAny(seg.Text, "\n\r") {
				t.Fatalf("row %d segment has a newline: %q", i, seg.Text)
			}
		}

		text := lineText(row.Segments)
		width := len([]rune(text))
		longest = max(longest, width)

		if width > cols {
			t.Fatalf("row %d (%s) is %d columns wide: %q", i, row.Kind, width, text)
		}
	}

	if longest == 0 {
		t.Fatal("preview produced no rows")
	}

	t.Logf("longest preview row is %d of %d columns", longest, cols)

	kinds := map[string]bool{}

	for _, row := range rows {
		kinds[row.Kind] = true
	}

	for _, want := range []string{"h1", "h2", "li", "p", "table", "gap"} {
		if !kinds[want] {
			t.Fatalf("preview has no %q row: %s", want, previewText(rows))
		}
	}

	// The Markdown table and the raw HTML table are the only table blocks in
	// the source, so two spans also prove the HTML table was structured.
	spans := previewTableSpans(rows)
	if len(spans) != 2 {
		t.Fatalf("table spans = %v, want the Markdown table and the HTML table", spans)
	}

	for _, span := range spans {
		assertPreviewTableSpans(t, rows, span)
	}

	// Render the same source in the app at its narrowest width, with the
	// details drawer taking another 280 px off the editor.
	dir := t.TempDir()
	writeFile(t, dir, "README.md", previewSpacingSource)

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.Page().SetSize(minWidth, minHeight)
	app.toggleDetails()

	if got := app.previewCols(); got > cols {
		t.Fatalf("app preview renders at %d columns, want %d or fewer", got, cols)
	}

	app.togglePreview()

	if err := app.Page().Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if !view.Preview || len(view.PreviewRows) == 0 {
		t.Fatalf("app preview = %v rows = %d", view.Preview, len(view.PreviewRows))
	}

	if view.TrackTop != 0 {
		t.Fatalf("app preview starts at track %d, want 0", view.TrackTop)
	}

	for i, row := range view.PreviewRows {
		if width := len([]rune(lineText(row.Segments))); width > cols {
			t.Fatalf("app preview row %d is %d columns wide: %q", i, width, lineText(row.Segments))
		}
	}

	t.Logf("app preview rows = %d at track %d", len(view.PreviewRows), view.TrackTop)
}

// TestLongLineLastRuneWrapsPastTheClip checks a line longer than the editor
// wraps in source and in preview, and that no visual row is wider than the
// column budget that fits in the pane.
func TestLongLineLastRuneWrapsPastTheClip(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	const marker = "FINAL-CONTENT"
	line := strings.Repeat("ordinary markdown words ", 80) + marker
	writeFile(t, dir, "README.md", line+"\n")

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("README.md")
	app.Page().SetSize(1160, 900)

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	doc := app.activeDoc()
	cols := app.codeColsForGroup(0)
	plan := doc.wrapped(cols)
	assertWrappedMarker(t, "source", plan.Rows, cols, marker)
	assertBodyInk(t, app, "source")

	app.toggleEdit()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	assertBodyInk(t, app, "edit")
	app.toggleEdit()
	app.togglePreview()

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	pcols := app.previewCols()
	rows := doc.Preview
	assertPreviewMarker(t, rows, pcols, marker)
	assertBodyInk(t, app, "preview")
}

// assertBodyInk checks the first painted rows end a full glyph inside the
// code body. A scrollbar or the line-number gutter is not the right edge.
func assertBodyInk(t *testing.T, app *App, label string) {
	t.Helper()

	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("missing code body")
	}

	img := mustDecodePNG(t, app.Page().Probe().PNG())
	right := int(body.X + body.W)

	for row := 0; row < inkSampleRows; row++ {
		y := int(body.Y) + row*rowH + rowH/2
		ink := rightInk(img, int(body.X), right, y)
		if ink < 0 {
			continue
		}

		if gap := right - ink; gap < codeClipSlack {
			t.Fatalf("%s row %d ink ends %d px from the body edge", label, row, gap)
		}
	}
}

// TestEditorRowsStopAboveTheFileNav checks the last line of a long file can
// scroll fully above the changed-file pill.
func TestEditorRowsStopAboveTheFileNav(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()
	writeFile(t, dir, "README.md", strings.Repeat("editor line\n", 80))

	app, err := New(dir)
	if err != nil {
		t.Fatal(err)
	}

	app.openFile("README.md")
	app.changes = changeSet{Files: []changedFile{{Path: "README.md", Status: "M", Add: 1}}}
	app.fileNav = newModifiedNav(app.changes)
	app.Page().SetSize(1160, 900)
	app.scrollTo(app.contentHeight())

	if err := app.Page().Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	view := app.View()
	if len(view.Rows) == 0 {
		t.Fatal("no visible rows")
	}

	last := view.Rows[len(view.Rows)-1]
	if last.No != 80 {
		t.Fatalf("last visible row is line %d, want 80", last.No)
	}

	if used := len(view.Rows) * rowH; used > app.editorClearHeight() {
		t.Fatalf("visible rows use %d px, clear height %d", used, app.editorClearHeight())
	}

	body, ok := boxByID(app.Page().Boxes(), "code-body")
	if !ok {
		t.Fatal("missing code body")
	}

	nav, ok := boxByID(app.Page().Boxes(), "file-nav")
	if !ok {
		t.Fatal("missing file navigator")
	}

	rowBottom := body.Y + float64(len(view.Rows)*rowH)
	if rowBottom > nav.Y {
		t.Fatalf("text ends at %.0f, pill starts at %.0f", rowBottom, nav.Y)
	}
}

func assertWrappedMarker(t *testing.T, label string, rows []codeRow, cols int, marker string) {
	t.Helper()

	found := -1

	for i, row := range rows {
		text := lineText(row.Segments)
		if n := len([]rune(text)); n > cols {
			t.Fatalf("%s row %d is %d runes, budget %d: %q", label, i, n, cols, text)
		}

		if strings.Contains(text, marker) {
			found = i
		}
	}

	if found <= 0 {
		t.Fatalf("%s marker %q is on row %d, want a later row", label, marker, found)
	}
}

func assertPreviewMarker(t *testing.T, rows []previewRow, cols int, marker string) {
	t.Helper()

	found := -1

	for i, row := range rows {
		text := lineText(row.Segments)
		if n := len([]rune(text)); n > cols {
			t.Fatalf("preview row %d is %d runes, budget %d: %q", i, n, cols, text)
		}

		if strings.Contains(text, marker) {
			found = i
		}
	}

	if found <= 0 {
		t.Fatalf("preview marker %q is on row %d, want a later row", marker, found)
	}
}

func mustDecodePNG(t *testing.T, raw []byte) image.Image {
	t.Helper()

	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}

	return img
}

// inkDarkR, inkDarkG, and inkDarkB are the code-body background (#0a0d13)
// plus a little headroom. Pixels at or under this stay out of the ink scan.
const (
	inkSampleRows = 4
	inkDarkR      = 30
	inkDarkG      = 34
	inkDarkB      = 40
)

func rightInk(img image.Image, left, right, y int) int {
	b := img.Bounds()
	if y < b.Min.Y || y >= b.Max.Y {
		return -1
	}

	last := -1

	for x := left; x < right && x < b.Max.X; x++ {
		r, g, b, a := img.At(x, y).RGBA()
		if a == 0 {
			continue
		}

		if r>>8 < inkDarkR && g>>8 < inkDarkG && b>>8 < inkDarkB {
			continue
		}

		last = x
	}

	return last
}

// previewTableSpans returns the row range of every table block. A gap row is
// inside the range only when it separates two visual groups of table rows.
func previewTableSpans(rows []previewRow) [][2]int {
	var spans [][2]int

	for i := 0; i < len(rows); i++ {
		if rows[i].Kind != "table" {
			continue
		}

		end := i

		for j := i + 1; j < len(rows); j++ {
			if rows[j].Kind == "table" {
				end = j

				continue
			}

			if rows[j].Kind == "gap" && j+1 < len(rows) && rows[j+1].Kind == "table" {
				continue
			}

			break
		}

		spans = append(spans, [2]int{i, end})
		i = end
	}

	return spans
}

// assertPreviewTableSpans checks that one table has no foreign row inside it,
// that every gap follows a wrapped logical row, and that each logical row took
// more than one visual row.
func assertPreviewTableSpans(t *testing.T, rows []previewRow, span [2]int) {
	t.Helper()

	depth := 0
	gaps := 0

	for k := span[0]; k <= span[1]; k++ {
		switch rows[k].Kind {
		case "table":
			depth++
		case "gap":
			if k == span[0] || k == span[1] {
				t.Fatalf("table %v has a gap at its edge: %d", span, k)
			}

			if depth < 2 {
				t.Fatalf("table %v gap at %d follows %d visual rows, want a wrapped row", span, k, depth)
			}

			gaps++
			depth = 0
		default:
			t.Fatalf("table %v is split by row %d (%s)", span, k, rows[k].Kind)
		}
	}

	if gaps == 0 {
		t.Fatalf("table %v has wrapped logical rows but no gap: %q", span, previewText(rows[span[0]:span[1]+1]))
	}

	if depth < 2 {
		t.Fatalf("table %v last logical row has %d visual rows, want a wrapped row", span, depth)
	}
}
