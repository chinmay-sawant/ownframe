package main

import (
	"strconv"
	"strings"
)

// The inline card is the popover above a changed line in the open editor. It
// is a pure model: the app fills it from the working-tree change map and the
// open document, and the template paints it over the gutter marker at the
// card's box. The card has no stored height, so the template autosizes it.
const (
	inlineCardWidth    = 360
	inlineCardLeft     = 12
	inlineCardGap      = 8   // the card sits this far above its row and page top
	inlineCardReserve  = 120 // page room the card leaves below its own top
	inlineCardTextCols = 44  // monospace cells that fit inside the card
	inlineCardMaxLines = 8   // selected lines shown before the card counts the rest
	inlineCardHint     = "Uncommitted change"
	inlineCardLineHint = "Current line"
)

// The card's kind labels. The change map's kinds are lowercase, so the card
// maps them through inlineCardKind.
const (
	cardKindAdded    = "Added"
	cardKindModified = "Modified"
	cardKindDeleted  = "Deleted"
	cardKindLine     = "Line"
	cardKindSelected = "Selection"
)

// inlineCard is the popover above one changed line. Left, Top, and Width are
// page pixels, and the text fields are already flattened onto one line.
// A multi-line selection fills Snippet (up to inlineCardMaxLines source lines
// joined with newlines, with the rest counted in MoreLines) and EndLine;
// single-line cards keep OldText and NewText. OffX and OffY are the user's
// drag offset from the anchored position, kept across scrolls.
type inlineCard struct {
	Show      bool
	Line      int
	EndLine   int
	Left      int
	Top       int
	Width     int
	Title     string
	Path      string
	Kind      string
	OldText   string
	NewText   string
	Snippet   string
	MoreLines int
	OffX      int
	OffY      int
	Hint      string
	HasPrev   bool
	HasNext   bool
}

// inlineCardFor builds the card for one line. rowTop is the row's top
// in page pixels, pageW is the page width, pageH is the page height, line is
// the 1-based source line, change is its entry in the change map, path names
// the file, and newText is the text the open document shows at that line. An
// addition has no old side and a deletion has no new side. A line without a
// change entry keeps inlineNone and shows the current text under the Line
// kind. The card is left-anchored and 360 px wide, so pageW does not move it.
func inlineCardFor(rowTop, pageW, pageH, line int, change inlineLine, path, newText string) inlineCard {
	oldText, freshText := inlineCardTexts(change, newText)
	kind := inlineCardKind(change.Kind)
	hint := inlineCardHint
	if change.Kind == inlineNone {
		hint = inlineCardLineHint
	}

	return inlineCard{
		Show:    true,
		Left:    inlineCardLeft,
		Top:     inlineCardTop(rowTop, pageH),
		Width:   inlineCardWidth,
		Title:   "Line " + strconv.Itoa(line),
		Path:    inlineCardText(strings.TrimSpace(path), inlineCardTextCols),
		Kind:    kind,
		OldText: oldText,
		NewText: freshText,
		Hint:    hint,
	}
}

// inlineCardForSelection builds the card for a multi-line selection. start
// and end are 1-based inclusive source lines, lines holds the document text
// for that range, rowTop positions the card above the selection start, and
// pageW does not move the left-anchored card. The shown lines render as one
// snippet with line breaks kept; only the first inlineCardMaxLines render and
// the rest count toward MoreLines.
func inlineCardForSelection(rowTop, pageW, pageH, start, end int, path string, lines []string) inlineCard {
	shown := min(len(lines), inlineCardMaxLines)
	clip := make([]string, 0, shown)
	for _, text := range lines[:shown] {
		clip = append(clip, inlineCardSnippetLine(text, inlineCardTextCols))
	}
	more := len(lines) - shown
	if more < 0 {
		more = 0
	}
	n := end - start + 1
	if n < 1 {
		n = len(lines)
	}
	return inlineCard{
		Show:      true,
		Line:      start,
		EndLine:   end,
		Left:      inlineCardLeft,
		Top:       inlineCardTop(rowTop, pageH),
		Width:     inlineCardWidth,
		Title:     "Lines " + strconv.Itoa(start) + "-" + strconv.Itoa(end),
		Path:      inlineCardText(strings.TrimSpace(path), inlineCardTextCols),
		Kind:      cardKindSelected,
		Snippet:   strings.Join(clip, "\n"),
		MoreLines: more,
		Hint:      strconv.Itoa(n) + " lines selected",
	}
}

// inlineCardSnippetLine clips one snippet line to width runes, keeping the
// leading indentation so the card reads as one code block. Tabs expand to two
// spaces and trailing whitespace is trimmed; overlong lines end in an
// ellipsis.
func inlineCardSnippetLine(text string, width int) string {
	text = strings.ReplaceAll(text, "\t", "  ")
	text = strings.TrimRight(text, " ")
	if width < 1 {
		return ""
	}
	runes := []rune(text)
	if len(runes) <= width {
		return text
	}
	return string(runes[:width]) + "…"
}

// inlineCardTop lifts the card above its row and keeps it on the page. A card
// near the bottom sits at pageH-120 so its autosized body stays in view, and a
// card near the top stops at 8 px.
func inlineCardTop(rowTop, pageH int) int {
	return max(inlineCardGap, min(rowTop-inlineCardGap, pageH-inlineCardReserve))
}

// inlineCardKind maps a change map kind to the card's label. A line without a
// change entry reads as Line; the app builds a card for any line the gutter
// or Ctrl+I names.
func inlineCardKind(kind inlineKind) string {
	switch kind {
	case inlineAdded:
		return cardKindAdded
	case inlineDeleted:
		return cardKindDeleted
	case inlineNone:
		return cardKindLine
	default:
		return cardKindModified
	}
}

// inlineCardTexts clips the two sides the card shows. An addition keeps only
// the document line and a deletion only the replaced text; a modified line
// shows both.
func inlineCardTexts(change inlineLine, newText string) (string, string) {
	oldText := change.OldText

	switch change.Kind {
	case inlineAdded:
		oldText = ""
	case inlineDeleted:
		newText = ""
	}

	return inlineCardText(oldText, inlineCardTextCols), inlineCardText(newText, inlineCardTextCols)
}

// inlineCardText flattens text onto one line and clips it to width runes,
// marking the cut with an ellipsis. A line break or a tab becomes a space so
// two tokens never join.
func inlineCardText(text string, width int) string {
	text = strings.NewReplacer("\r\n", " ", "\r", " ", "\n", " ", "\t", " ").Replace(text)
	text = strings.TrimSpace(text)

	if width < 1 {
		return ""
	}

	runes := []rune(text)
	if len(runes) <= width {
		return text
	}

	return string(runes[:width]) + "…"
}
