package form

// CanEdit reports whether typing edits c.
func CanEdit(c Control) bool {
	if c.Tag == "select" || c.Type == "checkbox" || c.Type == "radio" {
		return false
	}

	return c.Tag == "textarea" || TextLike(c.Type)
}

// IsFileInput reports whether c is a file picker. It stays editable so a
// typed name works without a desktop dialog, but it never shows a caret.
func IsFileInput(c Control) bool {
	return c.Type == "file"
}

// WillChange reports whether a click can change the control's value.
func WillChange(c Control) bool {
	switch {
	case c.Type == "checkbox":
		return true
	case c.Type == "radio":
		return !c.Checked
	case c.Tag == "select":
		return len(c.Options) > 1
	default:
		return false
	}
}
