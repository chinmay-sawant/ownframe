package form

// RangeAction says how a redraw should treat the caret range.
type RangeAction int

const (
	// RangeClear drops the caret, the anchor, and the select-all flag.
	RangeClear RangeAction = iota
	// RangeZero zeroes the caret and the anchor and leaves the select-all flag.
	RangeZero
	// RangeClamp keeps the range inside the focused value.
	RangeClamp
)

// ClampRange decides how to clamp the caret after a merge.
// An empty focus clears the range. A missing or non-editable control
// zeroes the offsets and leaves the select-all flag alone.
func ClampRange(v FocusView) (int, int, RangeAction) {
	if v.FocusID == "" {
		return 0, 0, RangeClear
	}

	if !v.Present || !CanEdit(v.Control) {
		return 0, 0, RangeZero
	}

	n := RuneLen(v.Control.Value)

	return ClampPos(v.Caret, n), ClampPos(v.Anchor, n), RangeClamp
}

// WholeSelected reports whether id is focused and its whole value is selected.
func WholeSelected(v FocusView, id string) bool {
	if v.FocusID != id || !v.Present || !CanEdit(v.Control) {
		return false
	}

	if v.All {
		return true
	}

	n := RuneLen(v.Control.Value)
	start, end := RangeOf(v.Anchor, v.Caret, n)

	return start != end && start == 0 && end == n
}

// NextID is the control id step places away in order, wrapping.
// ok is false when order is empty.
func NextID(order []string, current string, step int) (string, bool) {
	if len(order) == 0 {
		return "", false
	}

	at := -1
	for i, id := range order {
		if id == current {
			at = i

			break
		}
	}

	next := 0
	if at < 0 {
		if step < 0 {
			next = len(order) - 1
		}
	} else {
		next = (at + step + len(order)) % len(order)
	}

	return order[next], true
}

// ToggleClick applies one click to a checkbox, radio, or select.
// clearRadios is true when other radios of the same name should turn off.
// Any other control is unchanged.
func ToggleClick(c Control) (Control, bool, bool) {
	switch {
	case c.Type == "checkbox":
		c.Checked = !c.Checked

		return c, true, false
	case c.Type == "radio":
		changed := !c.Checked
		c.Checked = true

		return c, changed, true
	case c.Tag == "select":
		before := SelectedIndex(c.Options)
		old := c.Value
		c = NextOption(c)

		return c, old != c.Value || before != SelectedIndex(c.Options), false
	default:
		return c, false, false
	}
}

// UncheckRadios clears every radio named name except keep.
// An empty name or a nil map does nothing.
func UncheckRadios(byID map[string]Control, name, keep string) {
	if name == "" || byID == nil {
		return
	}

	for id, other := range byID {
		if id == keep || other.Type != "radio" || other.Name != name {
			continue
		}

		other.Checked = false
		byID[id] = other
	}
}

// SetValue stores value on c, clamped to the field cap.
// A select marks each option whose value matches.
func SetValue(c Control, value string) Control {
	value = ClampField(value, FieldCap(c))
	c.Value = value
	if c.Tag == "select" {
		for i := range c.Options {
			c.Options[i].Selected = c.Options[i].Value == value
		}
	}

	return c
}
