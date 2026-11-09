package form

import (
	"strconv"
	"strings"
)

const (
	slotNext = 1
	slotPrev = -1
)

// InsertSlot inserts text, or, for a one-digit numeric slot, stores one
// digit and asks the page to focus the next slot. A non-digit is refused.
func InsertSlot(c Control, caret, anchor int, text string) (Control, int, int, bool) {
	if !isSlot(c) {
		return InsertValue(c, caret, anchor, text)
	}

	digit, ok := oneDigit(text)
	if !ok {
		return c, caret, anchor, false
	}

	changed := c.Value != digit
	if changed {
		c.Value = digit
	}
	c.SlotStep = slotNext

	return c, len([]rune(digit)), len([]rune(digit)), changed
}

// BackspaceSlot removes text, or, when a one-digit slot is already empty,
// asks the page to focus the previous slot.
func BackspaceSlot(c Control, caret, anchor int) (Control, int, int, bool) {
	if !isSlot(c) {
		return BackspaceValue(c, caret, anchor)
	}

	next, caret, anchor, changed := BackspaceValue(c, caret, anchor)
	if !changed && c.Value == "" {
		next.SlotStep = slotPrev
	}

	return next, caret, anchor, changed
}

// SlotNeighbor returns the id step slots away from id. step is 1 or -1.
// A missing, disabled, or non-slot neighbor returns "".
func SlotNeighbor(byID map[string]Control, id string, step int) string {
	if step == 0 || byID == nil {
		return ""
	}

	prefix, index, ok := slotID(id)
	if !ok {
		return ""
	}

	next := index + step
	if next < 0 {
		return ""
	}

	nid := prefix + "-" + strconv.Itoa(next)
	c, found := byID[nid]
	if !found || c.Disabled || !isSlot(c) {
		return ""
	}

	return nid
}

func isSlot(c Control) bool {
	if c.MaxLength != 1 || c.InputMode != "numeric" {
		return false
	}

	_, _, ok := slotID(c.ID)

	return ok
}

func slotID(id string) (string, int, bool) {
	dash := strings.LastIndex(id, "-")
	if dash <= 0 || dash == len(id)-1 {
		return "", 0, false
	}

	suffix := id[dash+1:]
	n, err := strconv.Atoi(suffix)
	if err != nil || n < 0 || strconv.Itoa(n) != suffix {
		return "", 0, false
	}

	return id[:dash], n, true
}

func oneDigit(text string) (string, bool) {
	runes := []rune(text)
	if len(runes) == 0 {
		return "", false
	}

	for _, r := range runes {
		if r < '0' || r > '9' {
			return "", false
		}
	}

	return string(runes[:1]), true
}
