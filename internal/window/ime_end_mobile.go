//go:build android || ios

package window

import "time"

func (s *shell) imeEndByUser() {
	if s.ime.rotation.endByUser(time.Now()) {
		return
	}
	s.ime.last = ""
	s.ime.pendingEnd = true
}
