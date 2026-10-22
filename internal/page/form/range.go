package form

import "unicode/utf8"

// RangeOf normalizes two offsets into [start, end) clamped to n runes.
func RangeOf(a, b, n int) (int, int) {
	start, end := a, b
	if start > end {
		start, end = end, start
	}

	start = ClampPos(start, n)
	end = ClampPos(end, n)
	if end < start {
		end = start
	}

	return start, end
}

// ClampPos clamps pos into [0, n].
func ClampPos(pos, n int) int {
	switch {
	case pos < 0:
		return 0
	case pos > n:
		return n
	default:
		return pos
	}
}

// RuneLen is utf8.RuneCountInString.
func RuneLen(s string) int {
	return utf8.RuneCountInString(s)
}

// RuneAtByte converts a byte offset into value to a rune offset, clamped
// to the value and snapped to a rune boundary.
func RuneAtByte(value string, off int) int {
	if off <= 0 {
		return 0
	}

	if off > len(value) {
		off = len(value)
	}

	for off > 0 && off < len(value) && !utf8.RuneStart(value[off]) {
		off--
	}

	return utf8.RuneCountInString(value[:off])
}
