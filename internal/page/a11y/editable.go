package a11y

import (
	"context"
	"unicode/utf8"

	"github.com/chinmay-sawant/ownframe/internal/clipboard"
	"github.com/chinmay-sawant/ownframe/internal/host"
)

func editText(ctx context.Context, actor Actor, current, insert string, edit host.AccessibilityTextEdit) error {
	start, end, ok := textByteRange(current, edit.Start, edit.End)
	if !ok {
		return host.ErrAccessibilityRange
	}
	switch edit.Kind {
	case host.AccessibilityCopyText, host.AccessibilityCutText:
		if err := clipboard.WriteContext(ctx, current[start:end]); err != nil {
			return err
		}
		if edit.Kind == host.AccessibilityCopyText {
			return nil
		}
		insert = ""
	case host.AccessibilityPasteText:
		var err error
		insert, err = clipboard.ReadContext(ctx)
		if err != nil {
			return err
		}
		if !utf8.ValidString(insert) {
			return host.ErrAccessibilityRange
		}
	case host.AccessibilityReplaceText:
	default:
		return host.ErrAccessibilityAction
	}
	return actor.IMEReplace(ctx, start, end, insert, len(insert))
}

func textByteRange(value string, start, end int) (int, int, bool) {
	if start < 0 || end < start {
		return 0, 0, false
	}
	index, first := 0, len(value)
	for offset := range value {
		if index == start {
			first = offset
		}
		if index == end {
			return first, offset, true
		}
		index++
	}
	return first, len(value), index == end
}
