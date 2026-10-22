package page

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

// Control is one form field taken from the executed HTML.
type Control = form.Control

// Option is one choice inside a select.
type Option = form.Option

// controlSpan is one control and where it sits in the executed HTML.
// Start is the index of '<'. End is the index after the element.
type controlSpan = form.Span

// formState is the value of each control on the current document.
// caret and anchor are rune offsets into the focused control's value; the
// selection is the range [min, max) and the caret is the moving end.
// all marks a select-all, which keeps FormSelected true for an empty value.
// blinkOn shows the caret; blinkAt is the last toggle or reset.
type formState struct {
	byID         map[string]Control
	order        []string
	focusID      string
	lastFocusID  string
	caret        int
	anchor       int
	all          bool
	doc          int
	blinkOn      bool
	blinkAt      time.Time
	preeditID    string
	preedit      string
	preeditStart int
	preeditEnd   int
}
