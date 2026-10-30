package main

import (
	"math"
	"strings"
)

const (
	commitTipWidth        = 340
	commitTipGap          = 10
	commitTipInset        = 6
	commitTipLift         = 4
	commitTipEdge         = 8
	commitTipPadH         = 14
	commitTipPadX         = 20
	commitTipLineH        = 14
	commitTipMinH         = 44
	commitTipMaxH         = 160
	commitTipSubjectLines = 2
	commitTipBodyLines    = 6
	commitTipMetaLines    = 1
	commitTipBodyMaxRunes = 280
	// commitTipCell is the measured advance of the 10px hover font, in pixels.
	commitTipCell = 4.45
	// historySubjectCols is how many runes fit the subject next to the graph
	// and the hash, in the 9.5px sidebar font, clear of the scrollbar.
	historySubjectCols = 33
)

// commitTipCols is the rune budget of one hover line. cardWidth includes the
// card's horizontal padding, which the text does not use.
func commitTipCols(cardWidth int, cell float64, pad int) int {
	text := cardWidth - pad
	if text < 1 {
		return 1
	}

	if cell <= 0 {
		cell = commitTipCell
	}

	cols := int(float64(text) / cell)
	if cols < 1 {
		return 1
	}

	return cols
}

// historySubjectLines wraps a commit subject onto the two lines the rail
// row can show. Breaks stay on word boundaries. A longer subject ends the
// second line with an ellipsis after a whole word.
func historySubjectLines(subject string) []string {
	words := strings.Fields(strings.TrimSpace(subject))
	if len(words) == 0 {
		return nil
	}

	first, rest := packWords(words, historySubjectCols)
	if first == "" {
		return nil
	}
	if len(rest) == 0 {
		return []string{first}
	}

	second, leftover := packWords(rest, historySubjectCols-1)
	if second == "" {
		return []string{first}
	}
	if len(leftover) > 0 {
		second += "…"
	}

	return []string{first, second}
}

// packWords joins words that fit in width runes. An overlong first word is
// cut so a single token cannot run past the column.
func packWords(words []string, width int) (line string, rest []string) {
	if width < 1 || len(words) == 0 {
		return "", words
	}

	used := 0
	taken := 0
	var b strings.Builder

	for _, word := range words {
		runes := []rune(word)
		need := len(runes)
		if used > 0 {
			need++
		}
		if used > 0 && used+need > width {
			break
		}
		if used == 0 && len(runes) > width {
			b.WriteString(string(runes[:width]))
			rest = append([]string{string(runes[width:])}, words[1:]...)

			return b.String(), rest
		}
		if used > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(word)
		used += need
		taken++
	}

	return b.String(), words[taken:]
}

// commitTipLayout is the page-pixel box of the commit hover card. Show is
// false when there is no page area or no text to draw, so the caller can skip
// the element entirely. The primary sets left, top, and width inline from
// these fields.
type commitTipLayout struct {
	Show   bool
	Left   int
	Top    int
	Width  int
	Height int
}

// layoutCommitTip places the hover card to the right of the sidebar, clear of
// the history rail's overflow. When the window leaves no room there, the card
// falls back inside the sidebar's left edge and covers the row band. rowTop is
// the hovered row's page top and rowH its height; the card rides the row's top
// edge and clamps inside the page so it never runs off the window.
func layoutCommitTip(pageW, pageH, sideLeft, sideW, rowTop, rowH, bodyLines int, z float64) commitTipLayout {
	if z <= 0 {
		z = 1
	}

	sc := func(v int) int { return max(1, int(math.Round(float64(v)*z))) }
	if pageW < 1 || pageH < 1 || bodyLines < 1 {
		return commitTipLayout{}
	}

	height := sc(commitTipPadH) + bodyLines*sc(commitTipLineH)
	if minH := sc(commitTipMinH); height < minH {
		height = minH
	}

	if maxH := sc(commitTipMaxH); height > maxH {
		height = maxH
	}

	left := sideLeft + sideW + sc(commitTipGap)
	if left+sc(commitTipWidth)+sc(commitTipGap) > pageW {
		left = sideLeft + sc(commitTipInset)
	}

	width := sc(commitTipWidth)
	if room := pageW - left - sc(commitTipGap); width > room {
		width = room
	}

	if width < 1 {
		width = 1
	}

	top := rowTop - sc(commitTipLift)
	if edge := sc(commitTipEdge); top < edge {
		top = edge
	}

	if bottom := pageH - height - sc(commitTipEdge); top > bottom {
		top = bottom
	}

	if edge := sc(commitTipEdge); top < edge {
		top = edge
	}

	return commitTipLayout{
		Show:   true,
		Left:   left,
		Top:    top,
		Width:  width,
		Height: height,
	}
}

// commitTipText returns the wrapped lines of the hover card: the subject
// first, then the body, then the meta line. width is the usable text width in
// runes, so no returned line is longer than width. The subject keeps two
// lines, the body six lines and 280 runes, and the meta one line; dropped text
// ends in an ellipsis.
func commitTipText(subject, body, meta string, width int) []string {
	if width < 1 {
		return nil
	}

	lines := tipWrap(subject, width, commitTipSubjectLines)
	lines = append(lines, tipWrap(tipCapRunes(tipCollapse(body), commitTipBodyMaxRunes), width, commitTipBodyLines)...)

	if tail := tipWrap(meta, width, commitTipMetaLines); len(tail) > 0 {
		lines = append(lines, tail...)
	}

	return lines
}

// tipCollapse folds every run of whitespace into one space so a body that
// spans several paragraphs still wraps as one block.
func tipCollapse(text string) string {
	return strings.Join(strings.Fields(text), " ")
}

// tipCapRunes shortens text to limit runes, ending the kept part in an
// ellipsis. limit below one yields an empty string.
func tipCapRunes(text string, limit int) string {
	if limit < 1 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}

	return string(runes[:limit-1]) + "…"
}

// tipWrap breaks text into at most maxLines lines of width runes, at a space
// where one fits and inside a word only when the word alone is wider than the
// line. The last line ends in an ellipsis when words were left over.
func tipWrap(text string, width, maxLines int) []string {
	if width < 1 || maxLines < 1 {
		return nil
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	lines := make([]string, 0, maxLines)

	for len(words) > 0 && len(lines) < maxLines {
		word := words[0]
		words = words[1:]

		if runes := []rune(word); len(runes) > width {
			lines = append(lines, string(runes[:width]))
			words = append([]string{string(runes[width:])}, words...)

			continue
		}

		line := word
		for len(words) > 0 {
			next := words[0]
			if len([]rune(line))+1+len([]rune(next)) > width {
				break
			}

			line += " " + next
			words = words[1:]
		}

		lines = append(lines, line)
	}

	if len(words) > 0 && len(lines) > 0 {
		lines[len(lines)-1] = tipCut(lines[len(lines)-1], width)
	}

	return lines
}

// tipCut ends a line in an ellipsis without growing it past width runes.
func tipCut(line string, width int) string {
	runes := []rune(line)
	if width < 1 {
		return ""
	}

	if len(runes) >= width {
		runes = runes[:width-1]
	}

	return string(runes) + "…"
}
