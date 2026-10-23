package form

import "strings"

// Mods strips the shift+, ctrl+, and alt+ prefixes from a key name.
func Mods(key string) (string, bool, bool) {
	extend, word := false, false
	for {
		switch {
		case strings.HasPrefix(key, "shift+"):
			extend = true
			key = key[len("shift+"):]
		case strings.HasPrefix(key, "ctrl+"):
			word = true
			key = key[len("ctrl+"):]
		case strings.HasPrefix(key, "alt+"):
			word = true
			key = key[len("alt+"):]
		default:
			return key, extend, word
		}
	}
}

// MoveCaret applies one caret navigation key to a field value.
// ok is false when key is not a caret move.
func MoveCaret(key, value string, caret, start, end int) (int, bool, bool) {
	key, extend, word := Mods(key)
	n := RuneLen(value)
	caret = ClampPos(caret, n)
	var pos int
	switch key {
	case "arrowleft":
		pos = moveLeft(value, caret, start, end, extend, word)
	case "arrowright":
		pos = moveRight(value, caret, start, end, extend, word)
	case "home":
		pos = 0
	case "end":
		pos = n
	default:
		return 0, false, false
	}

	return ClampPos(pos, n), extend, true
}

// CaretMoved reports whether storing pos would change the caret.
// extend false also treats a moved anchor or a select-all as a change.
func CaretMoved(caret, anchor, pos int, all, extend bool) bool {
	changed := pos != caret
	if !extend {
		changed = changed || anchor != pos || all
	}

	return changed
}
