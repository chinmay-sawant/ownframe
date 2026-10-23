package form

import "unicode"

// PrevWordStart returns the start of the word before pos.
func PrevWordStart(s string, pos int) int {
	runes := []rune(s)
	i := ClampPos(pos, len(runes))

	for i > 0 && unicode.IsSpace(runes[i-1]) {
		i--
	}

	for i > 0 && !unicode.IsSpace(runes[i-1]) {
		i--
	}

	return i
}

// NextWordStart returns the start of the word after pos.
func NextWordStart(s string, pos int) int {
	runes := []rune(s)
	i := ClampPos(pos, len(runes))

	for i < len(runes) && unicode.IsSpace(runes[i]) {
		i++
	}

	for i < len(runes) && !unicode.IsSpace(runes[i]) {
		i++
	}

	for i < len(runes) && unicode.IsSpace(runes[i]) {
		i++
	}

	return i
}
