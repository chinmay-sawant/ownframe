package gesture

import "testing"

func TestIMESplitCaretCountsBytes(t *testing.T) {
	t.Parallel()

	before, after, n := SplitCaret("héllo", 2)
	if before != "hé" || after != "llo" || n != 3 {
		t.Fatalf("before %q after %q bytes %d", before, after, n)
	}

	before, after, n = SplitCaret("hi", 9)
	if before != "hi" || after != "" || n != 2 {
		t.Fatalf("clamp before %q after %q bytes %d", before, after, n)
	}
}

func TestIMEReplaceSnapsToRunes(t *testing.T) {
	t.Parallel()

	field := textField("héllo")
	if _, _, ok := Replace(field, "", 3, 3, 0); ok {
		t.Fatal("empty range changed")
	}

	got, pos, ok := Replace(field, "", 2, 3, 0)
	if !ok || got != "hllo" || pos != 1 {
		t.Fatalf("snap %q %d %v", got, pos, ok)
	}

	got, pos, ok = Replace(textField(""), "ab", 0, 0, 2)
	if !ok || got != "ab" || pos != 2 {
		t.Fatalf("insert %q %d %v", got, pos, ok)
	}
}

func TestIMECompositionKeepsClearsAndSets(t *testing.T) {
	t.Parallel()

	cur := Preedit{}
	if _, redraw := Compose(cur, "e", "", 0, 0, true); redraw {
		t.Fatal("empty clear redrew")
	}

	cur = Preedit{ID: "e", Text: "か"}
	next, redraw := Compose(cur, "e", "", 0, 0, true)
	if !redraw || next != (Preedit{}) {
		t.Fatalf("clear %+v %v", next, redraw)
	}

	if _, redraw = Compose(cur, "e", "か", 0, 3, false); redraw {
		t.Fatal("unfocused composition redrew")
	}

	next, redraw = Compose(Preedit{}, "e", "かな", 0, 6, true)
	want := Preedit{ID: "e", Text: "かな", Start: 0, End: 6}
	if !redraw || next != want {
		t.Fatalf("set %+v %v", next, redraw)
	}

	if _, redraw = Compose(want, "e", "かな", 0, 6, true); redraw {
		t.Fatal("same composition redrew")
	}
}

func TestIMECaretSampleUsesPreedit(t *testing.T) {
	t.Parallel()

	field := textField("hello")
	got, off := CaretSample(field, field.Value, 2, 2, Preedit{})
	if got != "hello" || off != 2 {
		t.Fatalf("plain %q %d", got, off)
	}

	pre := Preedit{ID: "t", Text: "か", Start: len("か"), End: len("か")}
	got, off = CaretSample(field, field.Value, 1, 4, pre)
	if got != "hかo" || off != 2 {
		t.Fatalf("preedit %q %d", got, off)
	}

	secret := textField("ab")
	secret.Type = "password"
	got, off = CaretSample(secret, secret.Value, 1, 1, Preedit{ID: "other", Text: "x"})
	if got != "••" || off != 1 {
		t.Fatalf("password %q %d", got, off)
	}
}

func TestIMEAnchorNeedsADisplay(t *testing.T) {
	t.Parallel()

	if _, ok := Anchor(nil, textField("a"), "a", 0, 0, Preedit{}, 0, 0); ok {
		t.Fatal("nil display produced an anchor")
	}
}
