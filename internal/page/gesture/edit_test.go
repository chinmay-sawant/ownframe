package gesture

import (
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/page/form"
)

func TestEditInsertBackspaceDeleteAndPaste(t *testing.T) {
	t.Parallel()

	field := textField("abcd")
	next, caret, anchor, ok := Edit(field, "X", OpInsert, 2, 2)
	if !ok || next.Value != "abXcd" || caret != 3 || anchor != 3 {
		t.Fatalf("insert %q %d %d %v", next.Value, caret, anchor, ok)
	}

	next, caret, anchor, ok = Edit(next, "", OpBackspace, 3, 3)
	if !ok || next.Value != "abcd" || caret != 2 || anchor != 2 {
		t.Fatalf("backspace %q %d %d %v", next.Value, caret, anchor, ok)
	}

	next, caret, _, ok = Edit(next, "", OpDeleteWord, 2, 2)
	if !ok || next.Value != "cd" || caret != 0 {
		t.Fatalf("delete word %q %d %v", next.Value, caret, ok)
	}

	big := strings.Repeat("x", form.MaxPasteRunes+1)
	if got := form.RuneLen(ClampPaste(big)); got != form.MaxPasteRunes {
		t.Fatalf("paste cap %d", got)
	}

	next, caret, _, ok = Edit(textField(""), big, OpPaste, 0, 0)
	// The paste cap runs first. The field cap is tighter, so the merged value stops there.
	if !ok || form.RuneLen(next.Value) != form.DefaultMaxFieldLen || caret != form.DefaultMaxFieldLen {
		t.Fatalf("paste len %d caret %d ok %v", form.RuneLen(next.Value), caret, ok)
	}

	if _, _, _, ok = Edit(field, "z", Op(-1), 0, 0); ok {
		t.Fatal("unknown edit changed the field")
	}
}

func TestEditCutCopyAndSelectAll(t *testing.T) {
	t.Parallel()

	field := textField("hello")
	if got := Selected(field.Value, 0, 0); got != "hello" {
		t.Fatalf("empty range %q", got)
	}

	if got := Selected(field.Value, 1, 4); got != "ell" {
		t.Fatalf("range %q", got)
	}

	next, clipped, pos := Cut(field, 1, 4)
	if clipped != "ell" || next.Value != "ho" || pos != 1 {
		t.Fatalf("cut %q %q %d", clipped, next.Value, pos)
	}

	next, clipped, pos = Cut(field, 0, 0)
	if clipped != "hello" || next.Value != "" || pos != 0 {
		t.Fatalf("cut all %q %q %d", clipped, next.Value, pos)
	}

	sel := SelectAll("ab")
	if sel.Anchor != 0 || sel.Caret != 2 || !sel.All {
		t.Fatalf("select all %+v", sel)
	}

	sel = SelectAll("")
	if sel.Caret != 0 || !sel.All {
		t.Fatalf("empty select all %+v", sel)
	}
}
