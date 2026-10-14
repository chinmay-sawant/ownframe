package page

import "time"

// Control is one form field taken from the executed HTML.
// A control with an empty id is ignored.
type Control struct {
	ID       string
	Tag      string
	Type     string
	Name     string
	Bind     string
	Value    string
	Checked  bool
	Disabled bool
	// TabIndex is the author's tabindex. A negative value keeps the control
	// out of the Tab order. Zero means the attribute was absent.
	TabIndex int
	Options  []Option
}

// Option is one choice inside a select.
type Option struct {
	Value    string
	Label    string
	Selected bool
}

// controlSpan is one control and where it sits in the executed HTML.
// Start is the index of '<'. End is the index after the element.
type controlSpan struct {
	Control Control
	Start   int
	End     int
}

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

// textLike reports whether typing edits this control.
// An empty type is a text input.
func textLike(kind string) bool {
	switch kind {
	case "", "text", "password", "email", "search", "tel", "url", "number", "file":
		return true
	default:
		return false
	}
}
