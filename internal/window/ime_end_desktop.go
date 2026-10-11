//go:build !android && !ios && !js

package window

// imeEndByUser leaves field focus intact after a desktop input method closes.
func (s *shell) imeEndByUser() {}
