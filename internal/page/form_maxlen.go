package page

// DefaultMaxFieldLen caps one field value when the template sets no
// maxlength (GEN-28). Larger author values clamp to it, so pasted
// megabytes cannot grow page state without bound.
const DefaultMaxFieldLen = 4096

// MaxPasteRunes caps one paste or type call before it merges.
const MaxPasteRunes = 16384

// fieldCap is the rune cap for c: the author maxlength when set and
// sane, else the default.
func fieldCap(c Control) int {
	if c.MaxLength > 0 && c.MaxLength <= DefaultMaxFieldLen {
		return c.MaxLength
	}

	return DefaultMaxFieldLen
}

// clampField truncates s to n runes.
func clampField(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}

	return string(r[:n])
}
