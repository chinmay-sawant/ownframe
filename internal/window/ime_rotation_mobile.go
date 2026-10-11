//go:build android || ios

package window

import "time"

func (s *shell) imeLayoutChanged(w, h int) {
	s.ime.rotation.layout(w, h, time.Now())
}
