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
	inlineCardHint     = "Uncommitted change"
)

// The card's kind labels. The change map's kinds are lowercase, so the card
// maps them through inlineCardKind.
const (
	cardKindAdded    = "Added"
	cardKindModified = "Modified"
	cardKindDeleted  = "Deleted"
)

// inlineCard is the popover above one changed line. Left, Top, and Width are
// page pixels, and the text fields are already flattened onto one line.
type inlineCard struct {
	Show    bool
	Line    int
	Left    int
	Top     int
	Width   int
	Title   string
	Path    string
	Kind    string
	OldText string
	NewText string
	Hint    string
	HasPrev bool
	HasNext bool
}

// inlineCardFor builds the card for one changed line. rowTop is the row's top
// in page pixels, pageW is the page width, pageH is the page height, line is
// the 1-based source line, change is its entry in the change map, path names
// the file, and newText is the text the open document shows at that line. An
// addition has no old side and a deletion has no new side. The card is
// left-anchored and 360 px wide, so pageW does not move it.
func inlineCardFor(rowTop, pageW, pageH, line int, change inlineLine, path, newText string) inlineCard {
	oldText, freshText := inlineCardTexts(change, newText)

	return inlineCard{
		Show:    true,
		Left:    inlineCardLeft,
		Top:     inlineCardTop(rowTop, pageH),
		Width:   inlineCardWidth,
		Title:   "Line " + strconv.Itoa(line),
		Path:    inlineCardText(strings.TrimSpace(path), inlineCardTextCols),
		Kind:    inlineCardKind(change.Kind),
		OldText: oldText,
		NewText: freshText,
		Hint:    inlineCardHint,
	}
}

// inlineCardTop lifts the card above its row and keeps it on the page. A card
// near the bottom sits at pageH-120 so its autosized body stays in view, and a
// card near the top stops at 8 px.
func inlineCardTop(rowTop, pageH int) int {
	return max(inlineCardGap, min(rowTop-inlineCardGap, pageH-inlineCardReserve))
}

// inlineCardKind maps a change map kind to the card's label. A line the map
// does not mark reads as Modified; the app only builds a card for a changed
// line.
func inlineCardKind(kind inlineKind) string {
	switch kind {
	case inlineAdded:
		return cardKindAdded
	case inlineDeleted:
		return cardKindDeleted
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
