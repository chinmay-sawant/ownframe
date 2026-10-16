package page

import (
	"context"
	"strings"
)

// caretKey moves the caret for a focused text field and draws. Arrow keys
// move by rune, Home and End move to the ends, and a ctrl+ or alt+ prefix
// jumps by word. A shift+ prefix extends the range from the anchor.
func (p *Page) caretKey(ctx context.Context, key string) error {
	c, ok := p.typingTarget()
	if !ok {
		return nil
	}

	key, extend, word := caretMods(key)
	n := runeLen(c.Value)
	caret := clampPos(p.formState().caret, n)
	start, end := p.formState().bounds(n)
	pos := caret

	switch key {
	case "arrowleft":
		switch {
		case word:
			pos = prevWordStart(c.Value, caret)
		case start != end && !extend:
			pos = start
		default:
			pos--
		}
	case "arrowright":
		switch {
		case word:
			pos = nextWordStart(c.Value, caret)
		case start != end && !extend:
			pos = end
		default:
			pos++
		}
	case "home":
		pos = 0
	case "end":
		pos = n
	default:
		return nil
	}

	pos = clampPos(pos, n)
	changed := pos != p.formState().caret
	if !extend {
		changed = changed || p.formState().anchor != pos || p.formState().all
	}

	if !changed {
		return nil
	}

	p.setCaret(pos, extend)

	return p.Redraw(ctx)
}

// caretMods strips the shift+, ctrl+, and alt+ prefixes from a key name.
func caretMods(key string) (string, bool, bool) {
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
