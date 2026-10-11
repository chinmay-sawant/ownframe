package window

import "github.com/chinmay-sawant/ownframe/internal/host"

// Late callbacks from a cancelled session must not edit a newly focused field.
// Preedit already shown in the old field remains part of its draft.
func (s *shell) imeTargetCurrent(field string) bool {
	focus, ok := s.app.(host.Focuser)
	return ok && field != "" && focus.FocusID() == field
}
