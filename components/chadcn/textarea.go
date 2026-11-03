package chadcn

import (
	"strconv"

	"github.com/chinmay-sawant/ownframe/ui"
)

// TextareaOptions describes appearance and state for one multi-line field. The
// zero value is an enabled field about three rows tall.
type TextareaOptions struct {
	// Placeholder is shown while the value is empty.
	Placeholder string
	// Rows is the visible row count. Zero and negative values use three rows.
	Rows int
	// Disabled prevents edits and focus.
	Disabled bool
	// Invalid marks the field with aria-invalid so CSS can paint the error
	// border and assistive technology reports the state.
	Invalid bool
}

// Textarea builds a multi-line field with an initial value. The host owns the
// live value through Page.Form and the page keeps newlines while editing.
func Textarea(id, value string) *ui.Element {
	return TextareaWithOptions(id, value, TextareaOptions{})
}

// TextareaWithOptions builds a Textarea with a row count, placeholder, and
// state. Rows sets both the rows attribute and the laid-out height.
func TextareaWithOptions(id, value string, options TextareaOptions) *ui.Element {
	rows := options.Rows
	if rows <= 0 {
		rows = defaultTextareaRows
	}
	area := ui.Tag("textarea").ID(id).Class(classes("of-scope", "of-textarea", "of-focus")).
		Attr("rows", strconv.Itoa(rows)).
		MinHeight(ui.Px(float64(rows)*textareaLineHeight + textareaBlockHeight)).
		Text(value)
	if options.Placeholder != "" {
		area.Attr("placeholder", options.Placeholder)
	}
	if options.Invalid {
		area.Attr("aria-invalid", ariaBool(true))
	}
	return setDisabled(area, options.Disabled)
}

const (
	defaultTextareaRows = 3
	textareaLineHeight  = 20.0
	textareaBlockHeight = 16.0
)
