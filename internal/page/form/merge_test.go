package form

import "testing"

func TestMergeControls(t *testing.T) {
	t.Parallel()

	prev := map[string]Control{
		"t":   {ID: "t", Tag: "input", Value: "user", Disabled: true},
		"old": {ID: "old", Tag: "input", Value: "x"},
		"s":   {ID: "s", Tag: "select", Value: "two"},
		"m":   {ID: "m", Tag: "textarea", Value: "typed"},
		"p":   {ID: "p", Tag: "input", Type: "text", Value: "user"},
		"c":   {ID: "c", Tag: "input", Type: "checkbox", Checked: true},
		"q":   {ID: "q", Tag: "select", Value: "gone"},
	}
	src := `<input id="t" name="n" value="temp">` +
		`<select id="s" name="fresh"><option value="one" selected>One</option>` +
		`<option value="two">Two</option></select>` +
		`<textarea id="m" name="body">temp</textarea>` +
		`<input id="new" value="n">` +
		`<input id="t" type="password" value="second">` +
		`<input id="p" type="password" value="fresh">` +
		`<input id="c" type="checkbox" name="on">` +
		`<select id="q"><option value="one">One</option>` +
		`<option value="two" selected>Two</option></select>`
	got, order := Merge(prev, Scan(src))
	if got == nil || len(order) != 7 || order[0] != "t" || order[3] != "new" {
		t.Fatal(order)
	}
	if _, ok := got["old"]; ok {
		t.Fatal("dropped")
	}

	text := got["t"]
	if text.Value != "user" || text.Name != "n" || text.Type != "" || text.Disabled {
		t.Fatal(text)
	}
	sel := got["s"]
	if sel.Value != "two" || sel.Name != "fresh" || sel.Options[0].Selected || !sel.Options[1].Selected {
		t.Fatal(sel)
	}
	if got["m"].Value != "typed" || got["m"].Name != "body" || got["new"].Value != "n" {
		t.Fatal(got["m"])
	}
	if got["p"].Type != "password" || got["p"].Value != "fresh" || !got["c"].Checked || got["c"].Name != "on" {
		t.Fatal(got["p"])
	}
	fall := got["q"]
	if fall.Value != "two" || fall.Options[0].Selected || !fall.Options[1].Selected {
		t.Fatal(fall)
	}

	empty, none := Merge(nil, nil)
	if empty == nil || len(empty) != 0 || len(none) != 0 {
		t.Fatal("nil")
	}
}
