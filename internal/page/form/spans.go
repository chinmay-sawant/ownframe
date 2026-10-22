package form

import "strings"

// hasFieldSpans reports whether any control needs a paintable field span.
// Plain buttons are left in the source, so a page with only buttons does not
// pay for a copy, the form style block, or attribute rewriting.
func hasFieldSpans(spans []Span) bool {
	for _, sp := range spans {
		tag := strings.ToLower(sp.Control.Tag)
		kind := strings.ToLower(sp.Control.Type)
		if tag == "input" || tag == "textarea" || tag == "select" {
			return true
		}
		if TextLike(kind) || kind == "checkbox" || kind == "radio" {
			return true
		}
	}

	return false
}
