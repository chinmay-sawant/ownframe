package gesture

import "github.com/chinmay-sawant/ownframe/internal/page/form"

// Effect is what a key-down does after the page handler.
type Effect int

const (
	// EffectNone leaves focus and the caret where they are.
	EffectNone Effect = iota
	// EffectBlur clears focus. Escape does this, focused or not.
	EffectBlur
	// EffectCaret asks the page to run its caret key.
	EffectCaret
)

// KeyDown classifies key for the focused field.
// text is the field value. An unfocused call passes a zero field and
// empty text. Escape blurs. Arrow, home, and end keys, including a word
// or extend prefix, move the caret of an editable field. Any other key
// does nothing here. The page still owns the caret offsets.
func KeyDown(field form.Control, text, key string) Effect {
	if key == "escape" {
		return EffectBlur
	}

	// A zero control is not focused. An empty id is not a field the page edits.
	if field.ID == "" || field.Disabled || !form.CanEdit(field) {
		return EffectNone
	}

	if _, _, ok := form.MoveCaret(key, text, 0, 0, 0); !ok {
		return EffectNone
	}

	return EffectCaret
}
