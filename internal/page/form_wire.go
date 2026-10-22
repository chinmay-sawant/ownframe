package page

import (
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

const (
	commentOpenLen  = form.CommentOpenLen
	commentTotalLen = form.CommentTotalLen
	caretFontSize   = form.CaretFontSize
)

func headOpen(source string) int { return form.HeadOpen(source) }

func findHead(source string) int { return form.FindHead(source) }

func readOpen(s string, i int) (string, string, int, bool) {
	return form.ReadOpen(s, i)
}

func closeSpan(s string, i int, tag string) (int, int) {
	return form.CloseSpan(s, i, tag)
}

func attrValue(raw, name string) string {
	return form.AttrValue(raw, name)
}

func openTagDrop(tag, raw string, extra []string, drop func(string) bool) string {
	return form.OpenTagDrop(tag, raw, extra, drop)
}

func rewriteControls(source string, spans []controlSpan, live map[string]Control, focusID string, selected bool) string {
	return form.Rewrite(source, spans, live, focusID, selected)
}

func caretBox(d *layout.Display, id string) (layout.Box, bool) {
	return form.CaretBox(d, id)
}

func emptyCaret(d *layout.Display, box layout.Box) (layout.DisplayOp, float64) {
	return form.EmptyCaret(d, box)
}
