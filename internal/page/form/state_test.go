package form

import (
	"testing"
	"time"
)

func TestNextID(t *testing.T) {
	t.Parallel()

	if _, ok := NextID(nil, "a", 1); ok {
		t.Fatal("empty")
	}

	order := []string{"a", "b", "c"}
	if got, ok := NextID(order, "", 1); !ok || got != "a" {
		t.Fatalf("missing %q %v", got, ok)
	}

	if got, _ := NextID(order, "", -1); got != "c" {
		t.Fatal(got)
	}

	if got, _ := NextID(order, "c", 1); got != "a" {
		t.Fatal(got)
	}

	if got, _ := NextID(order, "a", -1); got != "c" {
		t.Fatal(got)
	}
}

func TestToggleAndRadios(t *testing.T) {
	t.Parallel()

	check := Control{Type: "checkbox"}
	check, changed, clear := ToggleClick(check)
	if !changed || clear || !check.Checked {
		t.Fatalf("check %+v %v %v", check, changed, clear)
	}

	radio := Control{Type: "radio", Checked: true}
	radio, changed, clear = ToggleClick(radio)
	if changed || !clear || !radio.Checked {
		t.Fatalf("radio %+v %v %v", radio, changed, clear)
	}

	sel := Control{Tag: "select", Options: []Option{{Value: "a", Selected: true}, {Value: "b"}}}
	sel, changed, clear = ToggleClick(sel)
	if !changed || clear || sel.Value != "b" || sel.Options[0].Selected {
		t.Fatalf("select %+v", sel)
	}

	text := Control{Tag: "input", Type: "text", Value: "a"}
	next, changed, clear := ToggleClick(text)
	if changed || clear || next.Value != "a" {
		t.Fatal(next)
	}

	byID := map[string]Control{
		"a": {Type: "radio", Name: "g", Checked: true},
		"b": {Type: "radio", Name: "g", Checked: true},
		"c": {Type: "radio", Name: "h", Checked: true},
	}
	UncheckRadios(byID, "g", "a")
	if !byID["a"].Checked || byID["b"].Checked || !byID["c"].Checked {
		t.Fatal(byID)
	}

	UncheckRadios(byID, "", "a")
	if !byID["a"].Checked {
		t.Fatal("empty name")
	}
}

func TestSetValueClamp(t *testing.T) {
	t.Parallel()

	c := SetValue(Control{Tag: "input", Type: "text", MaxLength: 2}, "abcd")
	if c.Value != "ab" {
		t.Fatal(c.Value)
	}

	sel := SetValue(Control{
		Tag:     "select",
		Options: []Option{{Value: "a", Selected: true}, {Value: "b"}},
	}, "b")
	if sel.Value != "b" || sel.Options[0].Selected || !sel.Options[1].Selected {
		t.Fatal(sel)
	}
}

func TestClampAndWhole(t *testing.T) {
	t.Parallel()

	if _, _, action := ClampRange(FocusView{}); action != RangeClear {
		t.Fatal(action)
	}

	_, _, action := ClampRange(FocusView{FocusID: "t"})
	if action != RangeZero {
		t.Fatal(action)
	}

	caret, anchor, action := ClampRange(FocusView{
		FocusID: "t",
		Caret:   9,
		Anchor:  -1,
		Present: true,
		Control: Control{Tag: "input", Type: "text", Value: "ab"},
	})
	if action != RangeClamp || caret != 2 || anchor != 0 {
		t.Fatalf("%d %d %v", caret, anchor, action)
	}

	view := FocusView{
		FocusID: "t",
		All:     true,
		Present: true,
		Control: Control{Tag: "input", Type: "text", Value: "ab"},
	}
	if !WholeSelected(view, "t") || WholeSelected(view, "other") {
		t.Fatal("all")
	}

	view.All = false
	view.Caret = 2
	if !WholeSelected(view, "t") {
		t.Fatal("range")
	}

	view.Caret = 1
	if WholeSelected(view, "t") {
		t.Fatal("partial")
	}
}

func TestCaretOfAndText(t *testing.T) {
	t.Parallel()

	plain := CaretOf(FocusView{})
	if plain.Caret != -1 || plain.Focus != "" {
		t.Fatal(plain)
	}

	field := Control{ID: "t", Tag: "input", Type: "text", Value: "abcd"}
	got := CaretOf(FocusView{
		FocusID:      "t",
		Caret:        9,
		Anchor:       1,
		All:          true,
		Present:      true,
		Control:      field,
		PreeditID:    "t",
		Preedit:      "Z",
		PreeditStart: 1,
	})
	if got.Focus != "t" || got.Caret != 4 || got.Start != 1 || got.End != 4 || !got.All || got.Preedit != "Z" {
		t.Fatal(got)
	}

	sel := CaretOf(FocusView{FocusID: "s", Present: true, Control: Control{ID: "s", Tag: "select"}})
	if sel.Focus != "s" || sel.Caret != -1 {
		t.Fatal(sel)
	}

	text, off, ok := CaretText(FocusView{Caret: 2, Control: field})
	if !ok || text != "abcd" || off != 2 {
		t.Fatalf("%q %d %v", text, off, ok)
	}

	_, _, ok = CaretText(FocusView{Control: field, PreeditID: "t", PreeditStart: 0, PreeditEnd: 1})
	if ok {
		t.Fatal("open preedit")
	}

	text, off, ok = CaretText(FocusView{Control: field, PreeditID: "t", Preedit: "Z", Caret: 1, Anchor: 1})
	if !ok || text != "aZbcd" || off != 1 {
		t.Fatalf("preedit %q %d %v", text, off, ok)
	}
}

func TestAdvanceBlink(t *testing.T) {
	t.Parallel()

	on, at := ResetBlink()
	if !on || !at.IsZero() {
		t.Fatal(on, at)
	}

	now := time.Unix(10, 0)
	on, at, draw := AdvanceBlink(false, time.Time{}, now)
	if !on || draw || !at.Equal(now) {
		t.Fatal(on, at, draw)
	}

	on, at, draw = AdvanceBlink(true, now, now.Add(BlinkInterval-time.Millisecond))
	if !on || draw || !at.Equal(now) {
		t.Fatal("early", on, draw)
	}

	on, at, draw = AdvanceBlink(true, now, now.Add(BlinkInterval))
	if on || !draw || at.Equal(now) {
		t.Fatal("toggle", on, draw, at)
	}
}
