package main

import (
	"strings"
	"testing"
)

func previewText(rows []previewRow) string {
	parts := make([]string, len(rows))

	for i, row := range rows {
		parts[i] = lineText(row.Segments)
	}

	return strings.Join(parts, "\n")
}

func TestPreviewRows(t *testing.T) {
	t.Parallel()

	lines := []string{
		"# Title",
		"",
		"Text with **bold** and `code`.",
		"",
		"- one",
		"- two",
		"",
		"> quote",
		"",
		"```go",
		"fmt.Println(\"x\")",
		"```",
		"",
		"| a | b |",
	}

	rows, anchors := previewRows(lines, previewDefaultCols)
	kinds := map[string]bool{}
	var text []string

	for _, row := range rows {
		kinds[row.Kind] = true
		text = append(text, lineText(row.Segments))
	}

	for _, want := range []string{"h1", "p", "li", "quote", "code", "table", "gap", "rule"} {
		if want == "rule" {
			continue
		}

		if !kinds[want] {
			t.Errorf("missing row kind %q", want)
		}
	}

	all := strings.Join(text, "\n")
	for _, want := range []string{"Title", "bold", "code", "one", "quote", "fmt.Println"} {
		if !strings.Contains(all, want) {
			t.Errorf("preview missing %q: %q", want, all)
		}
	}

	for _, unwanted := range []string{"# Title", "**bold**", "`code`"} {
		if strings.Contains(all, unwanted) {
			t.Errorf("preview kept markers %q", unwanted)
		}
	}

	if anchors[0] != 0 {
		t.Fatalf("heading anchor = %d", anchors[0])
	}
}

func TestPreviewTableWrap(t *testing.T) {
	t.Parallel()

	lines := []string{
		"| Name | Description |",
		"| --- | --- |",
		"| alpha | A long description that should wrap over several preview rows when the column is narrow. |",
	}

	rows, anchors := previewRows(lines, 40)
	all := previewText(rows)

	for _, want := range []string{"Name", "Description", "alpha", "description", "preview rows"} {
		if !strings.Contains(all, want) {
			t.Errorf("table missing %q: %q", want, all)
		}
	}

	if strings.Contains(all, "|") {
		t.Fatalf("table printed a literal pipe: %q", all)
	}

	for _, want := range []string{"│", "─"} {
		if !strings.Contains(all, want) {
			t.Errorf("table missing box character %q", want)
		}
	}

	bold := false

	for _, row := range rows {
		if !strings.HasPrefix(lineText(row.Segments), "│") {
			continue
		}

		for _, seg := range row.Segments {
			if seg.Class == classBold {
				bold = true
			}
		}
	}

	if !bold {
		t.Error("table header is not emphasised")
	}

	if len(rows) < 5 {
		t.Errorf("table rows = %d, want a wrapped body", len(rows))
	}

	for i := range lines {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Errorf("table anchor %d = %d", i, anchors[i])
		}
	}

	for _, row := range rows {
		if width := len([]rune(lineText(row.Segments))); width > 40 {
			t.Errorf("table row is %d columns wide: %q", width, lineText(row.Segments))
		}
	}
}

func TestPreviewTableDashboard(t *testing.T) {
	t.Parallel()

	lines := []string{
		"# Dashboard",
		"",
		"| Service | Status | Uptime |",
		"|:---|---:|:---:|",
		"| api | **ok** | 99.9% |",
		"| worker | down | 42% |",
	}

	rows, anchors := previewRows(lines, previewDefaultCols)
	all := previewText(rows)

	for _, want := range []string{"Service", "Status", "Uptime", "api", "ok", "99.9%", "worker", "down", "42%"} {
		if !strings.Contains(all, want) {
			t.Errorf("dashboard table missing %q: %q", want, all)
		}
	}

	if strings.Contains(all, "|") {
		t.Fatalf("dashboard table printed a literal pipe: %q", all)
	}

	if !strings.Contains(all, "┼") {
		t.Error("dashboard table has no header rule")
	}

	for i := range lines {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Errorf("dashboard anchor %d = %d", i, anchors[i])
		}
	}
}

func TestPreviewTableNarrowCells(t *testing.T) {
	t.Parallel()

	lines := []string{
		"| Name | Detail | Note |",
		"| --- | --- | --- |",
		"| alpha | beta gamma delta epsilon | zeta eta theta |",
	}

	rows, _ := previewRows(lines, 48)
	all := previewText(rows)

	for _, want := range []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta", "eta", "theta"} {
		if !strings.Contains(all, want) {
			t.Errorf("narrow table lost %q: %q", want, all)
		}
	}

	if strings.Contains(all, "|") {
		t.Fatalf("narrow table printed a literal pipe: %q", all)
	}
}

func TestPreviewHTMLTags(t *testing.T) {
	t.Parallel()

	lines := []string{
		`Text with <b>bold</b>, <i>italic</i>, <code>code()</code>, <a href="https://example.com">a link</a>, <kbd>Ctrl</kbd>, <mark>marked</mark> &amp; more &lt;ok&gt;.`,
		"",
		"Visible <!-- hidden --> text.",
	}

	rows, _ := previewRows(lines, previewDefaultCols)
	all := previewText(rows)

	for _, want := range []string{"bold", "italic", "code()", "a link", "Ctrl", "marked", "&", "more", "<ok>.", "Visible", "text"} {
		if !strings.Contains(all, want) {
			t.Errorf("HTML preview missing %q: %q", want, all)
		}
	}

	for _, unwanted := range []string{"<b>", "</b>", "<i>", "<code>", "<a ", "<kbd>", "<mark>", "<!--", "hidden"} {
		if strings.Contains(all, unwanted) {
			t.Errorf("HTML preview kept %q: %q", unwanted, all)
		}
	}
}

func TestPreviewHTMLBlock(t *testing.T) {
	t.Parallel()

	lines := []string{
		`<div class="note">`,
		"  <p>First <strong>paragraph</strong>.</p>",
		"  <p>Second line<br>Third line</p>",
		`  <img src="map.png" alt="a map">`,
		"</div>",
	}

	rows, anchors := previewRows(lines, previewDefaultCols)
	all := previewText(rows)

	for _, want := range []string{"First paragraph.", "Second line", "Third line", "a map"} {
		if !strings.Contains(all, want) {
			t.Errorf("HTML block missing %q: %q", want, all)
		}
	}

	for _, unwanted := range []string{"<div", "<p>", "<br>", "<img", "</div>"} {
		if strings.Contains(all, unwanted) {
			t.Errorf("HTML block kept tag %q: %q", unwanted, all)
		}
	}

	var second, third = -1, -1

	for i, row := range rows {
		switch lineText(row.Segments) {
		case "Second line":
			second = i
		case "Third line":
			third = i
		}
	}

	if second < 0 || third != second+1 {
		t.Fatalf("<br> rows = %d, %d", second, third)
	}

	for i := range lines {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Errorf("HTML block anchor %d = %d", i, anchors[i])
		}
	}
}

func TestPreviewMermaidFlowchart(t *testing.T) {
	t.Parallel()

	lines := []string{
		"```mermaid",
		"flowchart TD",
		"  A[Start] --> B{Ready?}",
		"  B -->|yes| C[Done]",
		"  C --- A",
		"```",
	}

	rows, anchors := previewRows(lines, 80)
	all := previewText(rows)
	diagrams := 0
	marked := false

	for _, row := range rows {
		if row.Kind == "diagram" {
			diagrams++
		}

		if row.Marker == "◇" {
			marked = true
		}

		if row.Kind == "code" {
			t.Fatalf("mermaid source stayed code: %q", lineText(row.Segments))
		}
	}

	for _, want := range []string{"Start", "Ready?", "Done", "yes", "▼", "┌", "│", "─"} {
		if !strings.Contains(all, want) {
			t.Errorf("flowchart missing %q: %q", want, all)
		}
	}

	if diagrams == 0 {
		t.Fatal("flowchart produced no diagram rows")
	}

	if !marked {
		t.Error("flowchart rows have no diagram marker")
	}

	for i := 1; i <= 4; i++ {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Errorf("mermaid source anchor %d = %d", i, anchors[i])
		}
	}
}

func TestPreviewMermaidSequence(t *testing.T) {
	t.Parallel()

	lines := []string{
		"```mermaid",
		"sequenceDiagram",
		"  participant A as Alice",
		"  participant B as Bob",
		"  A->>B: hello",
		"  B-->>A: hi",
		"```",
	}

	rows, anchors := previewRows(lines, 80)
	all := previewText(rows)

	for _, want := range []string{"Alice", "Bob", "hello", "hi", "▶", "◀", "╌"} {
		if !strings.Contains(all, want) {
			t.Errorf("sequence diagram missing %q: %q", want, all)
		}
	}

	for _, row := range rows {
		if row.Kind != "diagram" {
			t.Errorf("sequence row kind = %q", row.Kind)
		}
	}

	for i := 1; i <= 5; i++ {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Errorf("sequence anchor %d = %d", i, anchors[i])
		}
	}
}

func TestPreviewReflowWidth(t *testing.T) {
	t.Parallel()

	lines := []string{
		"# Heading",
		"",
		"A paragraph that is long enough to need several rows at a narrow width and only one row at a wide width.",
		"",
		"| Name | Detail |",
		"| --- | --- |",
		"| alpha | a long table cell that wraps differently |",
	}

	narrow, _ := previewRows(lines, 24)
	wide, _ := previewRows(lines, previewDefaultCols)

	if len(narrow) <= len(wide) {
		t.Fatalf("narrow rows = %d, wide rows = %d", len(narrow), len(wide))
	}

	clamped, _ := previewRows(lines, 5)
	if len(clamped) != len(narrow) {
		t.Errorf("clamped rows = %d, want %d", len(clamped), len(narrow))
	}

	doc := &fileDoc{Lines: lines}

	doc.ensurePreviewWidth(24)
	narrowRows := len(doc.Preview)

	doc.ensurePreviewWidth(previewDefaultCols)
	wideRows := len(doc.Preview)

	if narrowRows <= wideRows {
		t.Fatalf("cached narrow rows = %d, wide rows = %d", narrowRows, wideRows)
	}

	doc.ensurePreviewWidth(24)
	if len(doc.Preview) != narrowRows {
		t.Errorf("reflow back rows = %d, want %d", len(doc.Preview), narrowRows)
	}

	short := &fileDoc{Lines: []string{"One short line."}}
	short.ensurePreview()

	if len(short.Preview) != 1 {
		t.Errorf("default preview rows = %d", len(short.Preview))
	}
}

func TestPreviewMermaidUnsupported(t *testing.T) {
	t.Parallel()

	lines := []string{
		"```mermaid",
		"pie",
		`  "Dogs": 42`,
		`  "Cats": 58`,
		"```",
	}

	rows, anchors := previewRows(lines, previewDefaultCols)
	all := previewText(rows)
	codes := 0

	for _, row := range rows {
		if row.Kind == "code" {
			codes++
		}
	}

	if !strings.Contains(all, "not rendered") {
		t.Errorf("unsupported note missing: %q", all)
	}

	for _, want := range []string{"Dogs", "Cats"} {
		if !strings.Contains(all, want) {
			t.Errorf("unsupported source missing %q: %q", want, all)
		}
	}

	if codes != 3 {
		t.Errorf("unsupported source code rows = %d, want 3", codes)
	}

	for i := 1; i <= 3; i++ {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Errorf("unsupported anchor %d = %d", i, anchors[i])
		}
	}
}
