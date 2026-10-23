package gesture

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// Op names one typing or clipboard edit of a focused field.
type Op int

const (
	// OpInsert inserts text at the caret.
	OpInsert Op = iota
	// OpBackspace removes the selection or the rune before the caret.
	OpBackspace
	// OpDeleteWord removes the selection or the word before the caret.
	OpDeleteWord
	// OpPaste inserts text after the paste cap.
	OpPaste
)

// ClampPaste caps one paste to MaxPasteRunes before it merges (GEN-28).
func ClampPaste(text string) string {
	return form.ClampField(text, form.MaxPasteRunes)
}

// Edit applies op to field. text is the inserted runes for OpInsert and
// OpPaste. caret and anchor are rune offsets. The bool is false when
// nothing changed. OpPaste caps text before it inserts.
func Edit(field form.Control, text string, op Op, caret, anchor int) (form.Control, int, int, bool) {
	switch op {
	case OpBackspace:
		return form.BackspaceValue(field, caret, anchor)
	case OpDeleteWord:
		return form.DeleteWordValue(field, caret, anchor)
	case OpPaste:
		text = ClampPaste(text)

		return form.InsertValue(field, caret, anchor, text)
	case OpInsert:
		return form.InsertValue(field, caret, anchor, text)
	default:
		return field, caret, anchor, false
	}
}

// Selection is a caret range. All marks a select-all, including an empty value.
type Selection struct {
	Anchor int
	Caret  int
	All    bool
}

// SelectAll selects the whole text. The page stores the offsets.
func SelectAll(text string) Selection {
	return Selection{Caret: form.RuneLen(text), All: true}
}

// Selected returns the selected range of text, or the whole value when
// the range is empty.
func Selected(text string, anchor, caret int) string {
	runes := []rune(text)
	start, end := form.RangeOf(anchor, caret, len(runes))
	if start == end {
		return text
	}

	return string(runes[start:end])
}

// Cut removes the selected range, or the whole value when the range is empty.
// The int is the caret after the cut. The page clears the select-all flag.
func Cut(field form.Control, anchor, caret int) (form.Control, string, int) {
	runes := []rune(field.Value)
	start, end := form.RangeOf(anchor, caret, len(runes))
	clipped := Selected(field.Value, anchor, caret)
	if start == end {
		start, end = 0, len(runes)
	}

	out := make([]rune, 0, len(runes)-(end-start))
	out = append(out, runes[:start]...)
	out = append(out, runes[end:]...)
	field.Value = string(out)

	return field, clipped, start
}
