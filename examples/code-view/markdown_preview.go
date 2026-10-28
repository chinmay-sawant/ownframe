package main

import (
	"strings"
	"sync"
	"unicode"
)

const (
	previewDefaultCols = 72
	previewMinCols     = 24
	previewRuleRunes   = 30
	previewParts       = 4

	previewTableBorders = 3
	previewBorderPad    = 2
)

// previewWidths remembers the column count each rendered preview used. fileDoc
// lives in workspace.go, so the width is kept here rather than on the struct.
var previewWidths sync.Map //nolint:gochecknoglobals // per-document preview width cache, one entry per open file

// previewRow is one rendered Markdown row. It is exactly one editor row high,
// so the existing windowing math still applies.
type previewRow struct {
	Kind     string
	Marker   string
	Segments []segment
	Image    string
	ImageW   int
	Alt      string
	png      []byte
}

// ensurePreview builds and caches the rendered rows at the default width.
// A caller that knows the editor width passes it and gets a reflowed preview.
func (doc *fileDoc) ensurePreview(cols ...int) {
	width := previewDefaultCols
	if len(cols) > 0 {
		width = cols[0]
	}

	doc.ensurePreviewWidth(width)
}

// ensurePreviewWidth builds and caches the rendered rows for one Markdown file
// at the given width. The rows are rebuilt when the width changes.
func (doc *fileDoc) ensurePreviewWidth(cols int) {
	cols = normalizePreviewCols(cols)

	if doc.Preview != nil {
		if width, ok := previewWidths.Load(doc); ok {
			if cached, ok := width.(int); ok && cached == cols {
				return
			}
		}
	}

	rows, anchors := previewRows(doc.Lines, cols)
	doc.Preview = rows
	doc.PreviewLine = anchors
	previewWidths.Store(doc, cols)
}

// previewRows parses the Markdown source into rows and maps each source line
// to the preview row that starts its block. The preview is reflowed to cols.
func previewRows(lines []string, cols int) ([]previewRow, []int) {
	cols = normalizePreviewCols(cols)
	rows := make([]previewRow, 0, len(lines))
	anchors := make([]int, len(lines))

	for i := range anchors {
		anchors[i] = -1
	}

	i := 0
	if len(lines) > 0 && strings.TrimSpace(lines[0]) == "---" {
		i = 1
		for i < len(lines) {
			if strings.TrimSpace(lines[i]) == "---" {
				i++

				break
			}

			i++
		}
	}

	inFence := false
	fenceInfo := ""
	var fenceLines []string
	var fenceNos []int

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if inFence {
			if isFenceLine(trimmed) {
				if isMermaidInfo(fenceInfo) {
					base := len(rows)
					var lineRows []int

					rows, lineRows = appendMermaid(rows, fenceLines, cols)
					for k, no := range fenceNos {
						anchors[no] = base
						if k < len(lineRows) {
							anchors[no] = base + lineRows[k]
						}
					}
				}

				inFence = false
				fenceInfo = ""
				fenceLines = nil
				fenceNos = nil
				i++

				continue
			}

			if isMermaidInfo(fenceInfo) {
				fenceLines = append(fenceLines, line)
				fenceNos = append(fenceNos, i)
			} else {
				rows = append(rows, previewRow{Kind: "code", Segments: []segment{{Text: lineOrSpace(line), Class: classCode}}})
				anchors[i] = len(rows) - 1
			}

			i++

			continue
		}

		if isFenceLine(trimmed) {
			inFence = true
			fenceInfo = fenceInfoOf(trimmed)
			i++

			continue
		}

		marker, itemText := listMarker(trimmed)

		switch {
		case trimmed == "":
			if len(rows) == 0 || rows[len(rows)-1].Kind != "gap" {
				rows = append(rows, previewRow{Kind: "gap"})
			}
			anchors[i] = len(rows) - 1
			i++
		case headingStart(line):
			r := []rune(strings.TrimLeft(line, " "))
			level, marker := headingLevel(r)
			text := strings.TrimSpace(string(r[marker:]))
			start := len(rows)
			rows = appendPreviewWrapped(rows, headingKind(level), "", text, cols)
			anchors[i] = start
			i++
		case isMarkdownHR(trimmed):
			width := previewRuleRunes
			if cols < width {
				width = cols
			}

			rows = append(rows, previewRow{Kind: "rule", Segments: []segment{{Text: strings.Repeat("─", width), Class: classHR}}})
			anchors[i] = len(rows) - 1
			i++
		case isTableStart(lines, i):
			firstLine := i
			start := len(rows)
			var lineRows []int

			rows, lineRows, i = appendTable(rows, lines, i, cols)
			for k, rel := range lineRows {
				if firstLine+k < len(anchors) {
					anchors[firstLine+k] = start + rel
				}
			}
		case strings.HasPrefix(trimmed, "|"):
			start := len(rows)
			rows, _ = appendTableRow(rows, splitTableRow(trimmed), nil, nil, 0, cols, false)
			anchors[i] = start
			i++
		case isHTMLBlockStart(trimmed):
			start := len(rows)
			startLine := i

			var parts []string

			for i < len(lines) && strings.TrimSpace(lines[i]) != "" {
				parts = append(parts, lines[i])
				i++
			}

			rows = appendHTMLPreview(rows, strings.Join(parts, "\n"), cols)

			if len(rows) == start {
				rows = append(rows, previewRow{Kind: "gap"})
				anchors[startLine] = start
			} else {
				for no := startLine; no < i; no++ {
					anchors[no] = start
				}
			}
		case strings.HasPrefix(trimmed, ">"):
			text := strings.TrimSpace(strings.TrimLeft(trimmed, ">"))
			start := len(rows)
			rows = appendPreviewWrapped(rows, "quote", "", text, cols)
			anchors[i] = start
			i++
		case marker != "":
			start := len(rows)
			rows = appendPreviewWrapped(rows, "li", marker, itemText, cols)
			anchors[i] = start
			i++
		default:
			start := i
			parts := make([]string, 0, previewParts)

			for i < len(lines) && !specialStart(lines[i]) && !isTableStart(lines, i) {
				parts = append(parts, strings.TrimSpace(lines[i]))
				i++
			}

			if len(parts) == 0 {
				i++
				continue
			}

			startRow := len(rows)
			rows = appendPreviewWrapped(rows, "p", "", strings.Join(parts, " "), cols)
			anchors[start] = startRow
		}
	}

	if inFence && isMermaidInfo(fenceInfo) {
		base := len(rows)
		var lineRows []int

		rows, lineRows = appendMermaid(rows, fenceLines, cols)
		for k, no := range fenceNos {
			anchors[no] = base
			if k < len(lineRows) {
				anchors[no] = base + lineRows[k]
			}
		}
	}

	rows, starts := sanitizePreviewRows(rows, cols)

	for i, anchor := range anchors {
		if anchor >= 0 && anchor < len(starts) {
			anchors[i] = starts[anchor]
		}
	}

	// One blank grid row top and bottom so the first and last block never
	// touch the pane edge. Gap rows are exactly rowH, so the window's
	// row slicing and anchor math are unaffected.
	rows = append([]previewRow{{Kind: "gap"}}, rows...)

	for i, anchor := range anchors {
		if anchor >= 0 {
			anchors[i] = anchor + 1
		}
	}

	rows = append(rows, previewRow{Kind: "gap"})

	return rows, anchors
}

// normalizePreviewCols keeps a preview readable at the narrowest editor width.
func normalizePreviewCols(cols int) int {
	if cols < previewMinCols {
		return previewMinCols
	}

	return cols
}

// headingStart reports whether a line can start an ATX heading.
func headingStart(line string) bool {
	r := []rune(strings.TrimLeft(line, " "))
	level, _ := headingLevel(r)

	return level > 0
}

// specialStart reports whether a line ends a paragraph block.
func specialStart(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}

	if isFenceLine(trimmed) ||
		strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "|") {
		return true
	}

	if headingStart(line) || isMarkdownHR(trimmed) || isHTMLBlockStart(trimmed) {
		return true
	}

	marker, _ := listMarker(trimmed)

	return marker != ""
}

// listMarker returns the rendered marker and the item text.
func listMarker(trimmed string) (string, string) {
	if len(trimmed) > 1 && strings.ContainsRune("-*+", rune(trimmed[0])) && trimmed[1] == ' ' {
		return "•", strings.TrimSpace(trimmed[2:])
	}

	j := 0
	for j < len(trimmed) && trimmed[j] >= '0' && trimmed[j] <= '9' {
		j++
	}

	if j > 0 && j+1 < len(trimmed) && trimmed[j] == '.' && trimmed[j+1] == ' ' {
		return trimmed[:j+1], strings.TrimSpace(trimmed[j+2:])
	}

	return "", ""
}

// appendPreviewWrapped wraps prose and appends one row per visual line. Inline
// HTML and Markdown become styled runs before wrapping, so a tag never prints.
func appendPreviewWrapped(rows []previewRow, kind, marker, text string, cols int) []previewRow {
	return appendPreviewStyled(rows, kind, marker, previewInlineRuns(text), cols)
}

// appendPreviewStyled wraps styled runs and appends one row per visual line. A
// hard newline from <br> starts a new row, and the marker only lands on the
// first row of the block.
func appendPreviewStyled(rows []previewRow, kind, marker string, runs []segment, cols int) []previewRow {
	cols = normalizePreviewCols(cols)

	hard := splitRowSegments(runs)

	for len(hard) > 0 && !segmentsHaveText(hard[0]) {
		hard = hard[1:]
	}

	for len(hard) > 0 && !segmentsHaveText(hard[len(hard)-1]) {
		hard = hard[:len(hard)-1]
	}

	if len(hard) == 0 {
		return append(rows, previewRow{Kind: kind, Marker: marker, Segments: previewInline("")})
	}

	markerSet := false

	for _, hardLine := range hard {
		visual := wrapStyledRuns(hardLine, cols)
		if len(visual) == 0 {
			rows = append(rows, previewRow{Kind: kind, Segments: []segment{{Text: " ", Class: classPlain}}})

			continue
		}

		for _, segs := range visual {
			rowMarker := ""
			if !markerSet {
				rowMarker = marker
				markerSet = true
			}

			rows = append(rows, previewRow{Kind: kind, Marker: rowMarker, Segments: finishLine(segs)})
		}
	}

	return rows
}

// appendHTMLPreview turns one raw HTML block into structured preview rows. It
// falls back to the plain text lines when nothing structured came out.
func appendHTMLPreview(rows []previewRow, raw string, cols int) []previewRow {
	blocks := htmlPreviewBlocks(raw)
	start := len(rows)

	for _, block := range blocks {
		if len(rows) > start && (block.Kind != "li" || rows[len(rows)-1].Kind != "li") {
			rows = append(rows, previewRow{Kind: "gap"})
		}

		switch {
		case block.Table != nil:
			rows, _ = appendTableBlock(rows, block.Table.Header, block.Table.Rows, nil, cols)
		case block.Kind == "rule":
			width := previewRuleRunes
			if cols < width {
				width = cols
			}

			rows = append(rows, previewRow{Kind: "rule", Segments: []segment{{Text: strings.Repeat("─", width), Class: classHR}}})
		default:
			rows = appendPreviewStyled(rows, block.Kind, block.Marker, block.Runs, cols)
		}
	}

	if len(rows) == start {
		for _, text := range htmlTextLines(raw) {
			rows = appendPreviewWrapped(rows, "p", "", text, cols)
		}
	}

	return rows
}

// previewInline strips the Markdown source markers and keeps the token classes.
// Boundary spaces become plain runs: markdownInline moves a trailing space into
// the next styled run, so a styled run may start or end with one.
func previewInline(text string) []segment {
	text = strings.ReplaceAll(stripHTML(text), "\n", " ")
	text = stripLinkedImages(text)
	segs := markdownInline(text)
	out := make([]segment, 0, len(segs))

	for _, seg := range segs {
		class := seg.Class
		if class == classPlain {
			class = ""
		}

		out = appendPreviewRun(out, decodeHTMLEntities(stripMarkers(seg)), class)
	}

	return finishLine(out)
}

// appendPreviewRun adds one preview run without moving its trailing space into
// the next run, so a word gap never leaves the plain run that carries it. A
// styled run contributes its boundary spaces as plain runs on either side.
func appendPreviewRun(dst []segment, text, class string) []segment {
	if text == "" {
		return dst
	}

	if class == "" {
		return appendPlainRun(dst, text)
	}

	lead, core, trail := splitBoundarySpace(text)
	if core == "" {
		return appendPlainRun(dst, text)
	}

	dst = appendPlainRun(dst, lead)
	dst = appendClassRun(dst, core, class)

	return appendPlainRun(dst, trail)
}

// appendPlainRun appends plain text, merging with a plain run already at the end.
func appendPlainRun(dst []segment, text string) []segment {
	return appendClassRun(dst, text, "")
}

// appendClassRun appends text with a class, merging with the previous run when
// the class matches.
func appendClassRun(dst []segment, text, class string) []segment {
	if text == "" {
		return dst
	}

	if n := len(dst); n > 0 && dst[n-1].Class == class {
		dst[n-1].Text += text

		return dst
	}

	return append(dst, segment{Text: text, Class: class})
}

// splitBoundarySpace separates the leading and trailing spaces of one run from
// its content, so a styled run never owns a word gap.
func splitBoundarySpace(text string) (lead, core, trail string) {
	core = strings.TrimSpace(text)
	if core == "" {
		return text, "", ""
	}

	start := strings.Index(text, core)
	end := start + len(core)

	return text[:start], core, text[end:]
}

// plainInline returns the text a rendered inline run will show.
func plainInline(text string) string {
	segs := previewInline(text)

	var b strings.Builder

	for _, seg := range segs {
		b.WriteString(seg.Text)
	}

	return b.String()
}

// stripMarkers removes the markers a rendered view should not print. Spaces
// outside a marker pair stay in the string for appendPreviewRun to separate.
func stripMarkers(seg segment) string {
	if seg.Class == "" || seg.Class == classPlain {
		return seg.Text
	}

	lead, core, trail := splitBoundarySpace(seg.Text)
	if core == "" {
		return seg.Text
	}

	switch seg.Class {
	case classBold:
		core = trimAffixes(core, "**", "__")
	case classItal:
		core = trimAffixes(core, "*", "_")
	case classCode:
		core = strings.Trim(core, "`")
	case classLink:
		core = linkLabel(core)
	}

	return lead + core + trail
}

// trimAffixes removes one marker from each end of text.
func trimAffixes(text string, markers ...string) string {
	for _, marker := range markers {
		text = strings.TrimSuffix(strings.TrimPrefix(text, marker), marker)
	}

	return text
}

// stripLinkedImages replaces a linked image, [![alt](src)](href), with its alt
// text before the inline scanner runs. The scanner stops at the inner ](, so
// without this the outer URL leaks into the rendered prose.
func stripLinkedImages(text string) string {
	for {
		start := strings.Index(text, "[![")
		if start < 0 {
			return text
		}

		inner := strings.Index(text[start+3:], "](")
		if inner < 0 {
			return text
		}

		innerEnd := strings.Index(text[start+3+inner+2:], ")")
		if innerEnd < 0 {
			return text
		}

		after := start + 3 + inner + 2 + innerEnd + 1
		if !strings.HasPrefix(text[after:], "](") {
			return text
		}

		outerEnd := strings.Index(text[after+2:], ")")
		if outerEnd < 0 {
			return text
		}

		alt := text[start+3 : start+3+inner]
		end := after + 2 + outerEnd + 1
		text = text[:start] + alt + text[end:]
	}
}

// linkLabel returns the visible part of a Markdown link or image. Spaces around
// the link stay in the string so a word gap is never dropped.
func linkLabel(text string) string {
	lead, core, trail := splitBoundarySpace(text)
	if core == "" {
		return text
	}

	open := strings.Index(core, "[")
	if open < 0 {
		return lead + strings.TrimSuffix(strings.TrimPrefix(core, "<"), ">") + trail
	}

	close := strings.Index(core[open:], "]")
	if close < 0 {
		return text
	}

	return lead + core[open+1:open+close] + trail
}

// wrapText breaks text at spaces and hard-breaks a word longer than width.
func wrapText(text string, width int) []string {
	if width < 1 {
		width = 1
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return []string{""}
	}

	lines := make([]string, 0, len(text)/width+1)
	current := ""

	for _, word := range words {
		r := []rune(word)

		if len(r) > width {
			if current != "" {
				lines = append(lines, current)
			}

			for len(r) > width {
				lines = append(lines, string(r[:width]))
				r = r[width:]
			}

			current = string(r)

			continue
		}

		if current == "" {
			current = word

			continue
		}

		if len([]rune(current))+1+len(r) <= width {
			current += " " + word
		} else {
			lines = append(lines, current)
			current = word
		}
	}

	if current != "" {
		lines = append(lines, current)
	}

	if len(lines) == 0 {
		return []string{""}
	}

	return lines
}

// previewGlyph is one visible rune and the token class it carries.
type previewGlyph struct {
	r     rune
	class string
}

// wrapStyledRuns wraps styled runs at spaces, hard-breaking a word longer than
// width. It returns one segment slice per visual line.
func wrapStyledRuns(segs []segment, width int) [][]segment {
	return wrapGlyphs(glyphsFromSegments(segs), width)
}

// glyphsFromSegments flattens styled runs and collapses whitespace so a row
// never carries a newline or a run of spaces into the renderer. A space where
// two classes meet is plain, so a link or code pill never swallows the gap
// around it.
func glyphsFromSegments(segs []segment) []previewGlyph {
	out := make([]previewGlyph, 0, len(segs))
	pendingSpace := false

	for _, seg := range segs {
		class := seg.Class
		if class == classPlain {
			class = ""
		}

		for _, r := range seg.Text {
			if unicode.IsSpace(r) {
				if len(out) > 0 {
					pendingSpace = true
				}

				continue
			}

			if pendingSpace {
				spaceClass := class
				if out[len(out)-1].class != class {
					spaceClass = ""
				}

				out = append(out, previewGlyph{r: ' ', class: spaceClass})
				pendingSpace = false
			}

			out = append(out, previewGlyph{r: r, class: class})
		}
	}

	return out
}

// wrapGlyphs breaks glyphs into visual lines no wider than width.
func wrapGlyphs(glyphs []previewGlyph, width int) [][]segment {
	if width < 1 {
		width = 1
	}

	if len(glyphs) == 0 {
		return nil
	}

	lines := make([][]segment, 0, len(glyphs)/width+1)
	start := 0

	for start < len(glyphs) {
		for start < len(glyphs) && glyphs[start].r == ' ' {
			start++
		}

		if start >= len(glyphs) {
			break
		}

		end := start
		lastSpace := -1

		for end < len(glyphs) && end-start < width {
			if glyphs[end].r == ' ' {
				lastSpace = end
			}

			end++
		}

		if end < len(glyphs) && glyphs[end].r != ' ' && lastSpace > start {
			end = lastSpace
		}

		line := glyphs[start:end]

		for len(line) > 0 && line[len(line)-1].r == ' ' {
			line = line[:len(line)-1]
		}

		lines = append(lines, segmentsFromGlyphs(line))
		start = end
	}

	return lines
}

// segmentsFromGlyphs groups one visual line back into styled runs.
func segmentsFromGlyphs(glyphs []previewGlyph) []segment {
	if len(glyphs) == 0 {
		return nil
	}

	out := make([]segment, 0, previewParts)

	var b strings.Builder

	class := glyphs[0].class

	for _, glyph := range glyphs {
		if glyph.class != class {
			if b.Len() > 0 {
				out = append(out, segment{Text: b.String(), Class: class})
				b.Reset()
			}

			class = glyph.class
		}

		b.WriteRune(glyph.r)
	}

	if b.Len() > 0 {
		out = append(out, segment{Text: b.String(), Class: class})
	}

	return out
}

// splitRowSegments splits styled runs at hard newlines, keeping each class.
func splitRowSegments(segs []segment) [][]segment {
	lines := [][]segment{nil}

	for _, seg := range segs {
		text := strings.ReplaceAll(seg.Text, "\r\n", "\n")
		text = strings.ReplaceAll(text, "\r", "\n")

		for k, part := range strings.Split(text, "\n") {
			if k > 0 {
				lines = append(lines, nil)
			}

			if part != "" {
				lines[len(lines)-1] = append(lines[len(lines)-1], segment{Text: part, Class: seg.Class})
			}
		}
	}

	return lines
}

// segmentsHaveText reports whether the runs print anything but whitespace.
func segmentsHaveText(segs []segment) bool {
	for _, seg := range segs {
		for _, r := range seg.Text {
			if !unicode.IsSpace(r) {
				return true
			}
		}
	}

	return false
}

// sanitizePreviewRows enforces the one-row-one-line rule. It returns the new
// rows and the first new row of every old row, so source anchors can be
// remapped after wrapping or splitting.
func sanitizePreviewRows(rows []previewRow, cols int) ([]previewRow, []int) {
	cols = normalizePreviewCols(cols)
	starts := make([]int, len(rows))
	out := make([]previewRow, 0, len(rows))

	for i, row := range rows {
		starts[i] = len(out)
		out = append(out, sanitizePreviewRow(row, cols)...)
	}

	return out, starts
}

// sanitizePreviewRow splits embedded newlines and keeps every row inside cols.
// Table, diagram, rule, and code rows clip so their alignment survives; prose
// rows wrap.
func sanitizePreviewRow(row previewRow, cols int) []previewRow {
	if row.Image != "" {
		return []previewRow{row}
	}

	if row.Kind == "gap" {
		return []previewRow{{Kind: row.Kind}}
	}

	hard := splitRowSegments(row.Segments)
	markerSet := false

	switch row.Kind {
	case "table", "diagram", "rule", "code":
		out := make([]previewRow, 0, len(hard))

		for _, line := range hard {
			segs := clipPreviewSegments(line, cols)
			if len(segs) == 0 {
				segs = []segment{{Text: " ", Class: classPlain}}
			}

			marker := ""
			if !markerSet {
				marker = row.Marker
				markerSet = true
			}

			out = append(out, previewRow{Kind: row.Kind, Marker: marker, Segments: segs})
		}

		return out
	default:
		out := make([]previewRow, 0, len(hard))

		for _, line := range hard {
			visual := wrapStyledRuns(line, cols)
			if len(visual) == 0 {
				out = append(out, previewRow{Kind: row.Kind, Marker: nextPreviewMarker(&markerSet, row.Marker), Segments: []segment{{Text: " ", Class: classPlain}}})

				continue
			}

			for _, segs := range visual {
				out = append(out, previewRow{Kind: row.Kind, Marker: nextPreviewMarker(&markerSet, row.Marker), Segments: finishLine(segs)})
			}
		}

		if len(out) == 0 {
			out = append(out, previewRow{Kind: row.Kind, Segments: []segment{{Text: " ", Class: classPlain}}})
		}

		return out
	}
}

// nextPreviewMarker returns the marker once, on the first row only.
func nextPreviewMarker(set *bool, marker string) string {
	if *set {
		return ""
	}

	*set = true

	return marker
}

// clipPreviewSegments caps styled runs at cols runes and marks a cut with an
// ellipsis that counts toward the width.
func clipPreviewSegments(segs []segment, cols int) []segment {
	out := make([]segment, 0, len(segs))
	left := cols

	for _, seg := range segs {
		if left <= 0 {
			break
		}

		r := []rune(seg.Text)
		if len(r) > left {
			if left > 1 {
				out = append(out, segment{Text: string(r[:left-1]) + "…", Class: seg.Class})
			} else {
				out = append(out, segment{Text: "…", Class: seg.Class})
			}

			break
		}
		out = append(out, seg)
		left -= len(r)
	}

	return out
}

// lineOrSpace keeps an empty code row visible.
func lineOrSpace(line string) string {
	if line == "" {
		return " "
	}

	return line
}

// isFenceLine reports whether a trimmed line opens or closes a code fence.
func isFenceLine(trimmed string) bool {
	return strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~")
}

// fenceInfoOf returns the info string after the opening fence markers.
func fenceInfoOf(trimmed string) string {
	return strings.TrimSpace(strings.TrimLeft(trimmed, "`~"))
}

// tableAlign is the Markdown column alignment.
type tableAlign int

const (
	tableLeft tableAlign = iota
	tableRight
	tableCenter
)

// isTableStart reports whether line i starts a Markdown table.
func isTableStart(lines []string, i int) bool {
	if i < 0 || i+1 >= len(lines) || !strings.Contains(lines[i], "|") {
		return false
	}

	_, ok := tableAlignments(lines[i+1])

	return ok
}

// tableAlignments parses a separator row and reports whether it is valid.
func tableAlignments(line string) ([]tableAlign, bool) {
	cells := splitTableRow(line)
	if len(cells) == 0 {
		return nil, false
	}

	out := make([]tableAlign, len(cells))

	for i, cell := range cells {
		cell = strings.TrimSpace(cell)
		if cell == "" || strings.Trim(cell, "-:") != "" || !strings.Contains(cell, "-") {
			return nil, false
		}

		left := strings.HasPrefix(cell, ":")
		right := strings.HasSuffix(cell, ":")

		switch {
		case left && right:
			out[i] = tableCenter
		case right:
			out[i] = tableRight
		default:
			out[i] = tableLeft
		}
	}

	return out, true
}

// splitTableRow removes the outer pipes and splits one row into cells. A
// backslash escapes a pipe, so the pipe stays in the cell text.
func splitTableRow(line string) []string {
	line = strings.TrimSpace(line)

	var (
		cells []string
		cell  strings.Builder
	)

	for i := 0; i < len(line); i++ {
		if line[i] == '\\' && i+1 < len(line) {
			if line[i+1] == '|' || line[i+1] == '\\' {
				cell.WriteByte(line[i+1])
				i++

				continue
			}
		}

		if line[i] == '|' {
			cells = append(cells, cell.String())
			cell.Reset()

			continue
		}

		cell.WriteByte(line[i])
	}

	cells = append(cells, cell.String())

	if len(cells) > 0 && strings.TrimSpace(cells[0]) == "" {
		cells = cells[1:]
	}

	if len(cells) > 0 && strings.TrimSpace(cells[len(cells)-1]) == "" {
		cells = cells[:len(cells)-1]
	}

	for i := range cells {
		cells[i] = strings.TrimSpace(cells[i])
	}

	return cells
}

// appendTable renders one table block and maps each source line to a row.
func appendTable(rows []previewRow, lines []string, start, cols int) ([]previewRow, []int, int) {
	header := splitTableRow(lines[start])
	align, hasSeparator := tableAlignments(lines[start+1])

	bodyStart := start + 1
	if hasSeparator {
		bodyStart = start + previewBorderPad
	}

	var body [][]string

	j := bodyStart
	if hasSeparator {
		for j < len(lines) && isTableBodyLine(lines[j]) {
			body = append(body, splitTableRow(lines[j]))
			j++
		}
	}

	lineRows := make([]int, j-start)
	if !hasSeparator {
		rows, _ = appendTableRow(rows, header, align, nil, 0, cols, false)

		return rows, lineRows, j
	}

	base := len(rows)
	rows, starts := appendTableBlock(rows, header, body, align, cols)

	for k, rel := range starts {
		index := 0
		if k > 0 {
			index = bodyStart - start + k - 1
		}

		if index >= 0 && index < len(lineRows) {
			lineRows[index] = rel - base
		}
	}
	return rows, lineRows, j
}

// appendTableBlock renders a header and body into table rows. It returns the
// row offset of every logical row: index 0 is the header, then one per body
// row. A gap row follows a wrapped row so adjacent multiline rows stay apart.
func appendTableBlock(rows []previewRow, header []string, body [][]string, align []tableAlign, cols int) ([]previewRow, []int) {
	widths, columns := tableWidths(header, body, cols)
	starts := make([]int, len(body)+1)
	starts[0] = len(rows)

	var depth int

	rows, depth = appendTableRow(rows, header, align, widths, columns, cols, true)
	rows = append(rows, previewRow{Kind: "table", Segments: tableRule(widths)})

	wrapped := depth > 1

	for k, cells := range body {
		if wrapped {
			rows = append(rows, previewRow{Kind: "gap"})
		}

		starts[k+1] = len(rows)
		rows, depth = appendTableRow(rows, cells, align, widths, columns, cols, false)
		wrapped = depth > 1
	}

	return rows, starts
}

// isTableBodyLine reports whether a line belongs to the body of a table.
func isTableBodyLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || !strings.Contains(trimmed, "|") {
		return false
	}

	if isFenceLine(trimmed) || headingStart(line) || isMarkdownHR(trimmed) ||
		isHTMLBlockStart(trimmed) || strings.HasPrefix(trimmed, ">") {
		return false
	}

	marker, _ := listMarker(trimmed)

	return marker == ""
}

// tableWidths returns the fitted column widths and the column count.
func tableWidths(header []string, body [][]string, cols int) ([]int, int) {
	columns := len(header)

	for _, row := range body {
		if len(row) > columns {
			columns = len(row)
		}
	}

	if columns < 1 {
		columns = 1
	}

	maxColumns := (cols - 1) / (previewTableBorders + 1)
	if maxColumns < 1 {
		maxColumns = 1
	}

	if columns > maxColumns {
		columns = maxColumns
	}

	natural := make([]int, columns)
	for i := range natural {
		natural[i] = 1
	}

	for i, cell := range header {
		if i < columns {
			natural[i] = max(natural[i], len([]rune(plainInline(cell))))
		}
	}

	for _, row := range body {
		for i, cell := range row {
			if i < columns {
				natural[i] = max(natural[i], len([]rune(plainInline(cell))))
			}
		}
	}

	budget := cols - (previewTableBorders*columns + 1)
	if budget < columns {
		budget = columns
	}

	total := 0
	for _, width := range natural {
		total += width
	}

	for total > budget {
		widest := -1

		for i, width := range natural {
			if width > 1 && (widest < 0 || width > natural[widest]) {
				widest = i
			}
		}

		if widest < 0 {
			break
		}

		natural[widest]--
		total--
	}

	return natural, columns
}

// appendTableRow renders one logical table row, wrapping every cell to its
// column width. A nil widths slice computes the widths from the row itself. It
// returns the number of visual rows the logical row used.
func appendTableRow(rows []previewRow, cells []string, align []tableAlign, widths []int, columns, cols int, header bool) ([]previewRow, int) {
	if widths == nil || columns < 1 {
		widths, columns = tableWidths(cells, nil, cols)
	}

	cells = padTableCells(cells, columns)

	for i := range cells {
		cells[i] = plainInline(cells[i])
	}

	wrapped := make([][]string, columns)

	depth := 1

	for i := range wrapped {
		wrapped[i] = wrapText(cells[i], widths[i])
		depth = max(depth, len(wrapped[i]))
	}

	class := ""
	if header {
		class = classBold
	}

	for line := 0; line < depth; line++ {
		segs := addSeg(nil, "│", classPunc)

		for i := 0; i < columns; i++ {
			text := ""
			if line < len(wrapped[i]) {
				text = wrapped[i][line]
			}

			text = padTableCell(text, widths[i], tableAlignAt(align, i))
			segs = addSeg(segs, " ", classPunc)
			segs = addSeg(segs, text, class)
			segs = addSeg(segs, " ", classPunc)

			if i < columns-1 {
				segs = addSeg(segs, "│", classPunc)
			}
		}

		segs = addSeg(segs, "│", classPunc)
		rows = append(rows, previewRow{Kind: "table", Segments: segs})
	}

	return rows, depth
}

// padTableCells pads or merges cells to the table column count.
func padTableCells(cells []string, columns int) []string {
	if len(cells) > columns {
		merged := append([]string(nil), cells[:columns-1]...)
		merged = append(merged, strings.Join(cells[columns-1:], " "))

		return merged
	}

	out := make([]string, columns)
	copy(out, cells)

	return out
}

// tableAlignAt returns the alignment for a column.
func tableAlignAt(align []tableAlign, i int) tableAlign {
	if i < len(align) {
		return align[i]
	}

	return tableLeft
}

// padTableCell pads a cell to the column width using its alignment.
func padTableCell(text string, width int, align tableAlign) string {
	r := []rune(text)
	if len(r) > width {
		r = r[:width]
	}

	gap := width - len(r)
	if gap <= 0 {
		return string(r)
	}

	switch align {
	case tableRight:
		return strings.Repeat(" ", gap) + string(r)
	case tableCenter:
		left := gap / previewBorderPad

		return strings.Repeat(" ", left) + string(r) + strings.Repeat(" ", gap-left)
	default:
		return string(r) + strings.Repeat(" ", gap)
	}
}

// tableRule renders the separator between the header and the body.
func tableRule(widths []int) []segment {
	segs := []segment{{Text: "├", Class: classPunc}}

	for i, width := range widths {
		segs = append(segs, segment{Text: strings.Repeat("─", width+previewBorderPad), Class: classPunc})
		if i < len(widths)-1 {
			segs = append(segs, segment{Text: "┼", Class: classPunc})
		}
	}

	segs = append(segs, segment{Text: "┤", Class: classPunc})

	return segs
}
