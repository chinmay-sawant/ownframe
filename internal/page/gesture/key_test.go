package gesture

import (
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

func textField(value string) form.Control {
	return form.Control{ID: "t", Tag: "input", Type: "text", Value: value}
}

func TestKeyDownEscapeBlursWithOrWithoutAField(t *testing.T) {
	t.Parallel()

	if got := KeyDown(form.Control{}, "", "escape"); got != EffectBlur {
		t.Fatalf("unfocused escape = %v", got)
	}

	if got := KeyDown(textField("ab"), "ab", "escape"); got != EffectBlur {
		t.Fatalf("focused escape = %v", got)
	}

	if got := KeyDown(textField("ab"), "ab", "shift+escape"); got != EffectNone {
		t.Fatalf("shift+escape = %v", got)
	}
}

func TestKeyDownCaretKeys(t *testing.T) {
	t.Parallel()

	field := textField("abcd ef")
	keys := []string{
		"arrowleft", "arrowright", "home", "end",
		"ctrl+arrowright", "alt+arrowleft", "shift+arrowright", "ctrl+shift+end",
	}

	for _, key := range keys {
		if got := KeyDown(field, field.Value, key); got != EffectCaret {
			t.Fatalf("%s = %v, want caret", key, got)
		}
	}

	for _, key := range []string{"a", "backspace", "space", "enter"} {
		if got := KeyDown(field, field.Value, key); got != EffectNone {
			t.Fatalf("%s = %v, want none", key, got)
		}
	}
}

func TestKeyDownSkipsFieldsThatCannotMove(t *testing.T) {
	t.Parallel()

	disabled := textField("ab")
	disabled.Disabled = true
	if got := KeyDown(disabled, disabled.Value, "arrowleft"); got != EffectNone {
		t.Fatalf("disabled = %v", got)
	}

	button := form.Control{ID: "b", Tag: "button", Type: "button", Value: "Go"}
	if got := KeyDown(button, button.Value, "end"); got != EffectNone {
		t.Fatalf("button = %v", got)
	}

	if got := KeyDown(form.Control{}, "", "home"); got != EffectNone {
		t.Fatalf("unfocused home = %v", got)
	}
}
