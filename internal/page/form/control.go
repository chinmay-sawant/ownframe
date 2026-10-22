package form

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
	// MaxLength is the author's maxlength, or zero when absent. Edits clamp
	// to it (or the default cap), so pastes cannot grow state unbounded.
	MaxLength int
	Options   []Option
}

// Option is one choice inside a select.
type Option struct {
	Value    string
	Label    string
	Selected bool
}

// TextLike reports whether typing edits this control.
// An empty type is a text input.
func TextLike(kind string) bool {
	switch kind {
	case "", "text", "password", "email", "search", "tel", "url", "number", "file":
		return true
	default:
		return false
	}
}
