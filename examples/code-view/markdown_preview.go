package main

import "strings"

const (
	previewCols      = 72
	previewRuleRunes = 30
	previewParts     = 4
)

// previewRow is one rendered Markdown row. It is exactly one editor row high,
// so the existing windowing math still applies.
type previewRow struct {
	Kind     string
	Marker   string
	Segments []segment
}

// ensurePreview builds and caches the rendered rows for one Markdown file.
func (doc *fileDoc) ensurePreview() {
	if doc.Preview != nil {
		return
	}

	rows, anchors := previewRows(doc.Lines)
	doc.Preview = rows
	doc.PreviewLine = anchors
}

// previewRows parses the Markdown source into rows and maps each source line
// to the preview row that starts its block.
func previewRows(lines []string) ([]previewRow, []int) {
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

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if inFence {
			if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
				inFence = false
				i++

				continue
			}

			rows = append(rows, previewRow{Kind: "code", Segments: []segment{{Text: lineOrSpace(line), Class: classCode}}})
			anchors[i] = len(rows) - 1
			i++

			continue
		}

		marker, itemText := listMarker(trimmed)

		switch {
		case strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~"):
			inFence = true
			i++
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
			rows = append(rows, previewRow{Kind: headingKind(level), Segments: previewInline(text)})
			anchors[i] = len(rows) - 1
			i++
		case isMarkdownHR(trimmed):
			rows = append(rows, previewRow{Kind: "rule", Segments: []segment{{Text: strings.Repeat("─", previewRuleRunes), Class: classHR}}})
			anchors[i] = len(rows) - 1
			i++
		case strings.HasPrefix(trimmed, ">"):
			text := strings.TrimSpace(strings.TrimLeft(trimmed, ">"))
			start := len(rows)
			rows, _ = appendWrapped(rows, "quote", "", text)
			anchors[i] = start
			i++
		case strings.HasPrefix(trimmed, "|"):
			rows = append(rows, previewRow{Kind: "table", Segments: previewInline(trimmed)})
			anchors[i] = len(rows) - 1
			i++
		case marker != "":
			start := len(rows)
			rows, _ = appendWrapped(rows, "li", marker, itemText)
			anchors[i] = start
			i++
		default:
			start := i
			parts := make([]string, 0, previewParts)

			for i < len(lines) && !specialStart(lines[i]) {
				parts = append(parts, strings.TrimSpace(lines[i]))
				i++
			}

			if len(parts) == 0 {
				i++
				continue
			}

			startRow := len(rows)
			rows, _ = appendWrapped(rows, "p", "", strings.Join(parts, " "))
			anchors[start] = startRow
		}
	}

	return rows, anchors
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

	if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") ||
		strings.HasPrefix(trimmed, ">") || strings.HasPrefix(trimmed, "|") {
		return true
	}

	if headingStart(line) || isMarkdownHR(trimmed) {
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

// appendWrapped wraps text and appends one row per visual line.
func appendWrapped(rows []previewRow, kind, marker, text string) ([]previewRow, string) {
	lines := wrapText(text, previewCols)

	for i, line := range lines {
		rowMarker := ""
		if i == 0 {
			rowMarker = marker
		}

		rows = append(rows, previewRow{Kind: kind, Marker: rowMarker, Segments: previewInline(line)})
	}

	return rows, marker
}

// previewInline strips the Markdown source markers and keeps the token classes.
func previewInline(text string) []segment {
	segs := markdownInline(text)
	out := make([]segment, 0, len(segs))

	for _, seg := range segs {
		class := seg.Class
		if class == classPlain {
			class = ""
		}

		out = addSeg(out, stripMarkers(seg), class)
	}

	return finishLine(out)
}

// stripMarkers removes the markers a rendered view should not print.
func stripMarkers(seg segment) string {
	text := seg.Text

	switch seg.Class {
	case classBold:
		text = strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(text, "**"), "**"), "__"), "__")
	case classItal:
		text = strings.TrimPrefix(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSuffix(text, "*"), "*"), "_"), "_")
	case classCode:
		text = strings.Trim(text, "`")
	case classLink:
		text = linkLabel(text)
	}

	return text
}

// linkLabel returns the visible part of a Markdown link or image.
func linkLabel(text string) string {
	open := strings.Index(text, "[")
	if open < 0 {
		return strings.TrimSuffix(strings.TrimPrefix(text, "<"), ">")
	}

	close := strings.Index(text[open:], "]")
	if close < 0 {
		return text
	}

	label := text[open+1 : open+close]
	if strings.HasPrefix(text, "!") {
		return label
	}

	return label
}

// wrapText breaks text at spaces and hard-breaks a word longer than width.
func wrapText(text string, width int) []string {
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

// lineOrSpace keeps an empty code row visible.
func lineOrSpace(line string) string {
	if line == "" {
		return " "
	}

	return line
}
