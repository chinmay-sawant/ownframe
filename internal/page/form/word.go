package form

import "unicode"

// IsWordRune reports whether r is part of a word for a double-click.
func IsWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}

// WordBounds returns the word around pos as [start, end).
// A point past the last rune measures the word before it. A non-word rune
// returns an empty range at pos.
func WordBounds(s string, pos int) (int, int) {
	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return 0, 0
	}

	if pos >= n {
		pos = n - 1
	}

	if pos < 0 {
		pos = 0
	}

	if !IsWordRune(runes[pos]) {
		return pos, pos
	}

	start := pos
	for start > 0 && IsWordRune(runes[start-1]) {
		start--
	}

	end := pos + 1
	for end < n && IsWordRune(runes[end]) {
		end++
	}

	return start, end
}

// LineBounds returns the line around pos as [start, end).
func LineBounds(s string, pos int) (int, int) {
	runes := []rune(s)
	n := len(runes)
	pos = ClampPos(pos, n)

	start := 0
	for i := pos; i > 0; i-- {
		if runes[i-1] == '\n' {
			start = i

			break
		}
	}

	end := n
	for i := pos; i < n; i++ {
		if runes[i] == '\n' {
			end = i

			break
		}
	}

	return start, end
}
