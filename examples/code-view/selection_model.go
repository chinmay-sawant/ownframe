package main

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const crlfLen = 2

// textSelection is the anchor and caret of an editor selection. The anchor is
// where the drag started and the caret is where it now ends, so a drag up or
// left has its anchor after its caret. Both are rune offsets into the buffer.
type textSelection struct {
	Anchor int
	Caret  int
}

// Range returns the selection as ordered, clamped rune offsets.
func (s textSelection) Range(n int) (start, end int) {
	start = clampEditPos(s.Anchor, n)
	end = clampEditPos(s.Caret, n)

	if start > end {
		start, end = end, start
	}

	return start, end
}

// selectionLine is one line's start offset and its text without the break.
type selectionLine struct {
	Start int
	Text  string
}

// selectionLines splits text into lines and the rune offset where each line
// starts. A CRLF pair ends a line and belongs to neither side, while a lone CR
// stays in its line as text. The result always holds at least one line, so the
// empty buffer has a line 0.
func selectionLines(text string) []selectionLine {
	runes := []rune(text)
	lines := make([]selectionLine, 0, strings.Count(text, "\n")+1)
	start := 0

	for i := 0; i < len(runes); i++ {
		switch {
		case runes[i] == '\r' && i+1 < len(runes) && runes[i+1] == '\n':
			lines = append(lines, selectionLine{Start: start, Text: string(runes[start:i])})
			i++
			start = i + 1
		case runes[i] == '\n':
			lines = append(lines, selectionLine{Start: start, Text: string(runes[start:i])})
			start = i + 1
		}
	}

	return append(lines, selectionLine{Start: start, Text: string(runes[start:])})
}

// selectionFromPoints converts a drag's start and end line and column pairs to
// a selection. The start pair becomes the anchor and the end pair the caret, so
// a drag up or left keeps its direction. Columns count runes, not display
// cells, and a column past the line end clamps to that line's end.
func selectionFromPoints(text string, startLine, startCol, endLine, endCol int) textSelection {
	lines := selectionLines(text)

	return textSelection{
		Anchor: offsetAtPoint(lines, startLine, startCol),
		Caret:  offsetAtPoint(lines, endLine, endCol),
	}
}

// offsetAtPoint returns the rune offset of one line and column. Both clamp to
// the nearest line and rune in the buffer, so a drag past either edge still
// lands on a real offset.
func offsetAtPoint(lines []selectionLine, line, col int) int {
	row := lines[clampEditPos(line, len(lines)-1)]

	return row.Start + clampEditPos(col, utf8.RuneCountInString(row.Text))
}

// pointForOffset turns a rune offset into a zero-based line and column. The
// column counts runes in the line. A CRLF pair is one break, so an offset
// inside the pair reports the start of the next line; a lone CR is a column on
// its line. An offset past the buffer clamps to its last point.
func pointForOffset(text string, offset int) (line, col int) {
	runes := []rune(text)
	offset = clampEditPos(offset, len(runes))

	for i := 0; i < offset; i++ {
		switch {
		case runes[i] == '\r' && i+1 < len(runes) && runes[i+1] == '\n':
			line++
			col = 0

			if i+1 < offset {
				i++
			}
		case runes[i] == '\n':
			line++
			col = 0
		default:
			col++
		}
	}

	return line, col
}

// selectionLineRange returns the rune columns of one line that a selection
// covers. It reports ok false when the line is outside the selection or the
// selection only touches the break there, so a caller can skip an empty mark.
// A selection that covers a whole line includes its break and marks every rune
// on that line, but never the break itself.
func selectionLineRange(text string, sel textSelection, line int) (from, to int, ok bool) {
	row, ok := selectionLineAt(text, line)
	if !ok {
		return 0, 0, false
	}

	lineEnd := row.Start + utf8.RuneCountInString(row.Text)
	if sel.Anchor <= row.Start && sel.Caret <= row.Start {
		return 0, 0, false
	}

	if sel.Anchor >= lineEnd && sel.Caret >= lineEnd {
		return 0, 0, false
	}

	start, end := sel.Range(utf8.RuneCountInString(text))
	from = max(start, row.Start) - row.Start
	to = min(end, lineEnd) - row.Start

	if from >= to {
		return 0, 0, false
	}

	return from, to, true
}

// selectionLineAt returns one line without splitting the whole text, so a
// caller decorating a visible row does not allocate every line above it. It
// reports ok false when line is past the buffer.
func selectionLineAt(text string, line int) (selectionLine, bool) {
	if line < 0 {
		return selectionLine{}, false
	}

	startByte, startRune := 0, 0
	runeCount, lineCount := 0, 0

	for i, r := range text {
		switch {
		case r == '\r' && i+1 < len(text) && text[i+1] == '\n':
			if lineCount == line {
				return selectionLine{Start: startRune, Text: text[startByte:i]}, true
			}

			lineCount++
			runeCount += 2
			startRune = runeCount
			startByte = i + crlfLen
		case r == '\n':
			if i > 0 && text[i-1] == '\r' {
				continue
			}

			if lineCount == line {
				return selectionLine{Start: startRune, Text: text[startByte:i]}, true
			}

			lineCount++
			runeCount++
			startRune = runeCount
			startByte = i + 1
		default:
			runeCount++
		}
	}

	if lineCount == line {
		return selectionLine{Start: startRune, Text: text[startByte:]}, true
	}

	return selectionLine{}, false
}

// selectionText returns the buffer text a selection covers, ordered and clamped
// to the buffer.
func selectionText(text string, sel textSelection) string {
	runes := []rune(text)
	start, end := sel.Range(len(runes))

	return string(runes[start:end])
}

// wordSelection returns the rune range of the word or whitespace run at offset,
// or the single punctuation rune there. An offset at a line break looks back
// one rune, so a double click after the last word of a line selects that word.
// An empty range means the offset sits on a blank line.
func wordSelection(text string, offset int) (start, end int) {
	runes := []rune(text)
	offset = clampEditPos(offset, len(runes))
	if len(runes) == 0 {
		return 0, 0
	}

	i := offset
	if i == len(runes) || selectionBreakRune(runes, i) {
		i--
	}

	if i < 0 || selectionBreakRune(runes, i) {
		return offset, offset
	}

	switch {
	case isWordRune(runes[i]):
		start, end = i, i+1

		for start > 0 && isWordRune(runes[start-1]) {
			start--
		}

		for end < len(runes) && isWordRune(runes[end]) {
			end++
		}
	case selectionSpaceRune(runes, i):
		start, end = i, i+1

		for start > 0 && selectionSpaceRune(runes, start-1) {
			start--
		}

		for end < len(runes) && selectionSpaceRune(runes, end) {
			end++
		}
	default:
		start, end = i, i+1
	}

	return start, end
}

// lineSelection returns the rune range of the whole line that offset sits on,
// without the line break. An offset at the end of a line reports that line, so
// a line select from the caret stays on the visible line.
func lineSelection(text string, offset int) (start, end int) {
	line, _ := pointForOffset(text, offset)
	row, ok := selectionLineAt(text, line)
	if !ok {
		return 0, 0
	}

	return row.Start, row.Start + utf8.RuneCountInString(row.Text)
}

// selectionBreakRune reports whether i starts a line break. A CRLF pair is one
// break, and a lone CR is line text.
func selectionBreakRune(runes []rune, i int) bool {
	if i < 0 || i >= len(runes) {
		return false
	}

	if runes[i] == '\n' {
		return true
	}

	return runes[i] == '\r' && i+1 < len(runes) && runes[i+1] == '\n'
}

// selectionSpaceRune reports whether i is whitespace inside a line, so a
// whitespace run never swallows a line break.
func selectionSpaceRune(runes []rune, i int) bool {
	if i < 0 || i >= len(runes) {
		return false
	}

	return !selectionBreakRune(runes, i) && unicode.IsSpace(runes[i])
}
