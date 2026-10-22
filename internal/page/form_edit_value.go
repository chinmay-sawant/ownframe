package page

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// insertValue inserts text at the caret and replaces the selection.
func (p *Page) insertValue(c Control, caret, anchor int, text string) (Control, int, int, bool) {
	return form.InsertValue(c, caret, anchor, text)
}

// backspaceValue removes the selection, or the rune before the caret.
func (p *Page) backspaceValue(c Control, caret, anchor int) (Control, int, int, bool) {
	return form.BackspaceValue(c, caret, anchor)
}

// deleteWordValue removes the selection, or the word before the caret.
func (p *Page) deleteWordValue(c Control, caret, anchor int) (Control, int, int, bool) {
	return form.DeleteWordValue(c, caret, anchor)
}
