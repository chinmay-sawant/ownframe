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

func TestPreviewRowsNeverContainNewlines(t *testing.T) {
	t.Parallel()

	const cols = 32

	lines := []string{
		"# A heading that is long enough to wrap over several preview rows",
		"",
		"A paragraph with <b>inline bold</b> and a verylongunbrokenwordthatmusthardbreakacrossrows and more words.",
		"",
		"- a list item that also wraps because it is long enough to need more than one row",
		"",
		"> a quote that wraps onto another row because the width is narrow",
		"",
		"| Name | Description |",
		"| --- | --- |",
		"| alpha | a cell that wraps over several rows to force the sanitizer to work |",
		"",
		"```go",
		`fmt.Println("a code line that is much longer than the preview width allows")`,
		"```",
		"",
		"<details><summary>A summary</summary><p>Body with <em>emphasis</em> and a long tail that wraps.</p></details>",
	}

	rows, _ := previewRows(lines, cols)

	for i, row := range rows {
		if strings.ContainsAny(row.Marker, "\n\r") {
			t.Fatalf("row %d marker has a newline: %q", i, row.Marker)
		}

		if text := lineText(row.Segments); len([]rune(text)) > cols {
			t.Fatalf("row %d is %d columns wide: %q", i, len([]rune(text)), text)
		}

		for _, seg := range row.Segments {
			if strings.ContainsAny(seg.Text, "\n\r") {
				t.Fatalf("row %d segment has a newline: %q", i, seg.Text)
			}
		}
	}

	long := "one\ntwo three four five six seven eight nine ten eleven twelve thirteen fourteen fifteen"
	clean, starts := sanitizePreviewRows([]previewRow{{
		Kind:     "p",
		Segments: []segment{{Text: long, Class: classPlain}},
	}}, previewMinCols)

	if len(starts) != 1 || starts[0] != 0 {
		t.Fatalf("sanitizer starts = %v", starts)
	}

	if len(clean) < 2 {
		t.Fatalf("sanitizer rows = %d, want a split", len(clean))
	}

	for _, row := range clean {
		text := lineText(row.Segments)
		if strings.ContainsAny(text, "\n\r") {
			t.Fatalf("sanitized row has a newline: %q", text)
		}

		if width := len([]rune(text)); width > previewMinCols {
			t.Fatalf("sanitized row is %d columns wide: %q", width, text)
		}
	}
}

func TestPreviewTableWrappedRowsDoNotOverlap(t *testing.T) {
	t.Parallel()

	const cols = 44

	lines := []string{
		"| Name | Description |",
		"| --- | --- |",
		"| alpha | A long description that wraps over several preview rows. |",
		"| beta | Another long description that also wraps over several rows. |",
	}

	rows, anchors := previewRows(lines, cols)

	for i := range lines {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Fatalf("table anchor %d = %d", i, anchors[i])
		}
	}

	tableRows := 0
	hasRule := false
	var pipePos []int

	for i, row := range rows {
		text := lineText(row.Segments)
		if width := len([]rune(text)); width > cols {
			t.Fatalf("row %d is %d columns wide: %q", i, width, text)
		}

		if row.Kind == "table" && strings.Contains(text, "┼") {
			hasRule = true
		}

		if row.Kind != "table" || !strings.HasPrefix(text, "│") {
			continue
		}

		tableRows++

		var pos []int

		for j, r := range []rune(text) {
			if r == '│' {
				pos = append(pos, j)
			}
		}

		if len(pos) == 0 {
			t.Fatalf("table row %d has no separators: %q", i, text)
		}

		if pipePos == nil {
			pipePos = pos

			continue
		}

		if len(pos) != len(pipePos) {
			t.Fatalf("table row %d separators = %v, want %v", i, pos, pipePos)
		}

		for k := range pos {
			if pos[k] != pipePos[k] {
				t.Fatalf("table row %d separators = %v, want %v: %q", i, pos, pipePos, text)
			}
		}
	}

	if tableRows < 5 {
		t.Fatalf("table rows = %d, want a wrapped header and body", tableRows)
	}

	if !hasRule {
		t.Fatal("table has no header rule")
	}

	gap := false

	for i := anchors[2]; i < anchors[3]; i++ {
		if rows[i].Kind == "gap" {
			gap = true
		}
	}

	if !gap {
		t.Fatal("wrapped table rows have no gap row between them")
	}
}

func TestPreviewLongParagraphUsesOneRowPerLine(t *testing.T) {
	t.Parallel()

	const cols = 40

	paragraph := strings.TrimSpace(strings.Repeat("alpha ", 30))
	lines := []string{paragraph, "", "Beta paragraph."}

	rows, anchors := previewRows(lines, cols)
	want := wrapText(paragraph, cols)

	var rowsForParagraph []string

	gaps := 0

	for i, row := range rows {
		text := lineText(row.Segments)
		if width := len([]rune(text)); width > cols {
			t.Fatalf("row %d is %d columns wide: %q", i, width, text)
		}

		switch row.Kind {
		case "gap":
			gaps++
		case "p":
			if gaps == 0 {
				rowsForParagraph = append(rowsForParagraph, text)
			}
		}
	}

	if len(rowsForParagraph) != len(want) {
		t.Fatalf("paragraph rows = %d, want %d: %v", len(rowsForParagraph), len(want), rowsForParagraph)
	}

	if strings.Join(rowsForParagraph, " ") != paragraph {
		t.Fatalf("paragraph rows do not rejoin: %q", rowsForParagraph)
	}

	if gaps != 1 {
		t.Fatalf("gap rows = %d, want exactly one for the blank source line", gaps)
	}

	if anchors[2] <= anchors[0] {
		t.Fatalf("second paragraph anchor %d is not after the first %d", anchors[2], anchors[0])
	}
}

func TestPreviewRawHTMLHeadingsAndTable(t *testing.T) {
	t.Parallel()

	lines := []string{
		"<h2>Overview</h2>",
		"<p>Some <strong>bold</strong> text and <code>code()</code>.</p>",
		"<table>",
		"<tr><th>Name</th><th>Role</th></tr>",
		"<tr><td>Ada</td><td>Engineer</td></tr>",
		"<tr><td>Grace</td><td>Admiral and compiler author</td></tr>",
		"<tr><td>R&amp;D</td><td>Research</td></tr>",
		"</table>",
	}

	rows, anchors := previewRows(lines, previewDefaultCols)
	all := previewText(rows)

	hasHeading := false

	for _, row := range rows {
		if row.Kind == "h2" && lineText(row.Segments) == "Overview" {
			hasHeading = true
		}
	}

	if !hasHeading {
		t.Fatalf("raw HTML heading missing: %q", all)
	}

	for _, want := range []string{"Overview", "bold", "code()", "Name", "Role", "Ada", "Engineer", "Grace", "Admiral", "R&D", "Research"} {
		if !strings.Contains(all, want) {
			t.Errorf("raw HTML preview missing %q: %q", want, all)
		}
	}

	for _, unwanted := range []string{"<h2>", "<table>", "<tr>", "<td>", "<th>", "<strong>", "<p>"} {
		if strings.Contains(all, unwanted) {
			t.Errorf("raw HTML preview kept tag %q: %q", unwanted, all)
		}
	}

	for i := range lines {
		if anchors[i] < 0 || anchors[i] >= len(rows) {
			t.Errorf("HTML anchor %d = %d", i, anchors[i])
		}
	}

	bold := false

	for _, row := range rows {
		if row.Kind != "table" || !strings.HasPrefix(lineText(row.Segments), "│") {
			continue
		}

		for _, seg := range row.Segments {
			if seg.Class == classBold {
				bold = true
			}
		}

		break
	}

	if !bold {
		t.Error("raw HTML table header is not emphasised")
	}
}

func TestPreviewHTMLScriptStripped(t *testing.T) {
	t.Parallel()

	lines := []string{
		"<div>",
		`<script>const leak = "hidden";</script>`,
		`<style>.leak { color: red; }</style>`,
		`<p>Visible <strong>text</strong> before <script>alert("x")</script> after.</p>`,
		"<ul><li>First</li><li>Second</li></ul>",
		"<details><summary>More</summary><p>Details body</p></details>",
		"</div>",
	}

	rows, _ := previewRows(lines, 60)
	all := previewText(rows)

	for _, want := range []string{"Visible", "text", "before", "after", "First", "Second", "More", "Details body"} {
		if !strings.Contains(all, want) {
			t.Errorf("script preview missing %q: %q", want, all)
		}
	}

	for _, unwanted := range []string{"leak", "hidden", "color: red", "alert", "<script", "<style", "<div", "<p", "</"} {
		if strings.Contains(all, unwanted) {
			t.Errorf("script preview kept %q: %q", unwanted, all)
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
