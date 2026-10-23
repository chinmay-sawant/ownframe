package form

import "time"

// State is the value of each control on the current document.
// Caret and Anchor are rune offsets into the focused control's value; the
// selection is the range between them and the caret is the moving end.
// All marks a select-all, which keeps a whole-value selection for an empty
// value. BlinkOn shows the caret; BlinkAt is the last toggle or reset.
// PreeditID is the control showing an IME composition. Preedit is that text.
// PreeditStart and PreeditEnd are byte offsets into Preedit.
type State struct {
	ByID         map[string]Control
	Order        []string
	FocusID      string
	LastFocusID  string
	Caret        int
	Anchor       int
	All          bool
	Doc          int
	BlinkOn      bool
	BlinkAt      time.Time
	PreeditID    string
	Preedit      string
	PreeditStart int
	PreeditEnd   int
}

// NewDoc is an empty document for doc, with the caret shown.
func NewDoc(doc int) State {
	return State{
		ByID:    map[string]Control{},
		Doc:     doc,
		BlinkOn: true,
	}
}

// Bounds normalizes the caret range to [start, end) clamped to n runes.
func (s *State) Bounds(n int) (int, int) {
	return RangeOf(s.Anchor, s.Caret, n)
}

// ClearRange drops the caret, the anchor, and the select-all flag.
func (s *State) ClearRange() {
	s.Caret, s.Anchor, s.All = 0, 0, false
}

// RestartBlink shows the caret and restarts the blink clock.
func (s *State) RestartBlink() {
	on, at := ResetBlink()
	s.BlinkOn, s.BlinkAt = on, at
}

// CaretEnd puts the caret and the anchor at n.
func (s *State) CaretEnd(n int) {
	s.Caret, s.Anchor = n, n
	s.All = false
	s.RestartBlink()
}

// SetCaret moves the caret to pos within n runes.
// extend keeps the anchor.
func (s *State) SetCaret(pos, n int, extend bool) {
	pos = ClampPos(pos, n)
	if !extend {
		s.Anchor = pos
	}

	s.Caret = pos
	s.All = false
	s.RestartBlink()
}

// FocusCaret focuses id and puts both ends at off.
// changed is false when that was already the state.
func (s *State) FocusCaret(id string, off int) bool {
	changed := s.FocusID != id || s.Caret != off || s.Anchor != off || s.All
	s.FocusID = id
	s.Caret, s.Anchor = off, off
	s.All = false
	s.RestartBlink()

	return changed
}

// DragTo moves the caret to off and clears a select-all.
// changed is false when the caret was already there.
func (s *State) DragTo(off int) bool {
	if off == s.Caret {
		return false
	}

	s.Caret = off
	s.All = false

	return true
}

// SelectSpan focuses id and selects [start, end).
func (s *State) SelectSpan(id string, start, end int) {
	s.FocusID = id
	s.Anchor, s.Caret = start, end
	s.All = false
}

// Blur clears focus and the caret range.
// The returned id is the focus that was cleared, or "" when none was focused.
// LastFocusID records a non-empty cleared id.
func (s *State) Blur() string {
	old := s.FocusID
	if old != "" {
		s.LastFocusID = old
	}

	s.FocusID = ""
	s.ClearRange()

	return old
}

// ClampToFocus clamps the caret and the anchor to the focused value.
// A missing focus clears them. A missing or non-editable control zeroes
// the offsets and leaves the select-all flag.
func (s *State) ClampToFocus() {
	if s.FocusID == "" {
		s.ClearRange()

		return
	}

	c, ok := s.Lookup(s.FocusID)
	caret, anchor, action := ClampRange(FocusView{
		FocusID: s.FocusID,
		Caret:   s.Caret,
		Anchor:  s.Anchor,
		Control: c,
		Present: ok,
	})
	if action == RangeClear {
		s.ClearRange()

		return
	}

	s.Caret, s.Anchor = caret, anchor
}

// Lookup returns the control id, or false when id is absent.
func (s *State) Lookup(id string) (Control, bool) {
	if s == nil || s.ByID == nil {
		return Control{}, false
	}

	c, ok := s.ByID[id]

	return c, ok
}

// Activate toggles id and reports whether its value, checked state, or
// selection changed. A missing or disabled control does not change.
// A control that was not focused takes focus. An editable control puts
// the caret at the end of its value. Any other newly focused control
// drops the caret range. Refocusing the same control clears select-all.
func (s *State) Activate(id string) bool {
	c, ok := s.Lookup(id)
	if !ok || c.Disabled {
		return false
	}

	c, changed, clear := ToggleClick(c)
	if clear {
		UncheckRadios(s.ByID, c.Name, id)
	}

	s.ByID[id] = c
	if s.FocusID != id {
		s.FocusID = id
		if CanEdit(c) {
			s.CaretEnd(RuneLen(c.Value))
		} else {
			s.ClearRange()
		}

		return changed
	}

	s.All = false

	return changed
}

// WriteValue stores value for id, clamped like a typed edit.
// It does nothing when id is absent. The focused control puts the caret
// at the end of the new value. end is true when that caret move ran.
func (s *State) WriteValue(id, value string) (Control, bool) {
	c, ok := s.Lookup(id)
	if !ok {
		return Control{}, false
	}

	c = SetValue(c, value)
	s.ByID[id] = c
	if s.FocusID == id {
		s.CaretEnd(RuneLen(c.Value))
	}

	return c, true
}

// WriteChecked sets a checkbox or radio.
// It does nothing when id is missing or not a checkbox or radio.
// A radio set true clears other radios with the same non-empty name.
// A radio set false clears only that radio.
func (s *State) WriteChecked(id string, checked bool) {
	c, ok := s.Lookup(id)
	if !ok {
		return
	}

	switch c.Type {
	case "checkbox":
		c.Checked = checked
		s.ByID[id] = c
	case "radio":
		c.Checked = checked
		s.ByID[id] = c
		if checked {
			UncheckRadios(s.ByID, c.Name, id)
		}
	}
}

// FieldID is the focused field id, excluding buttons.
// A nil state or a focused button returns "".
func (s *State) FieldID() string {
	if s == nil {
		return ""
	}

	if c, ok := s.Lookup(s.FocusID); ok && c.Tag == "button" {
		return ""
	}

	return s.FocusID
}

// TakeFocus moves focus to id.
// moved is false when the control is missing, disabled, or already focused.
// reveal is true when the newly focused control is a button.
// An editable control puts the caret at the end. Any other move clears the range.
func (s *State) TakeFocus(id string, c Control, present bool) (moved, reveal bool) {
	if !present || c.Disabled || s.FocusID == id {
		return false, false
	}

	s.FocusID = id
	if CanEdit(c) {
		s.CaretEnd(RuneLen(c.Value))
	} else {
		s.ClearRange()
	}

	return true, c.Tag == "button"
}

// StepBlink advances the caret blink when the focused value is editable,
// not a file input, and has an empty range. The bool is true when the
// caller should draw. A false result leaves the clock unchanged, except
// the first call after a reset, which starts the clock without drawing.
func (s *State) StepBlink(now time.Time, n int, editable, file bool) bool {
	if !editable || file {
		return false
	}

	start, end := s.Bounds(n)
	if start != end {
		return false
	}

	on, at, draw := AdvanceBlink(s.BlinkOn, s.BlinkAt, now)
	s.BlinkOn, s.BlinkAt = on, at

	return draw
}

// View is the focused control, caret, range, and preedit.
func (s *State) View() FocusView {
	if s == nil || s.FocusID == "" {
		return FocusView{}
	}

	c, ok := s.ByID[s.FocusID]

	return FocusView{
		FocusID:      s.FocusID,
		Caret:        s.Caret,
		Anchor:       s.Anchor,
		All:          s.All,
		Control:      c,
		Present:      ok,
		PreeditID:    s.PreeditID,
		Preedit:      s.Preedit,
		PreeditStart: s.PreeditStart,
		PreeditEnd:   s.PreeditEnd,
	}
}
