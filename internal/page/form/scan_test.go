package form

import (
	"strings"
	"testing"
)

func TestHasControls(t *testing.T) {
	t.Parallel()

	for _, src := range []string{
		`<input id="t">`,
		`<INPUT ID="p" TYPE="password">`,
		`<textarea id="m">x</textarea>`,
		`<TEXTAREA ID="m">x</TEXTAREA>`,
		`<select id="s"><option>A</option></select>`,
		`<SELECT ID="s"><option>A</option></SELECT>`,
		`<button id="b">Go</button>`,
		`<BUTTON ID="b">Go</BUTTON>`,
		`<!-- <input id="z"> -->`,
	} {
		if !HasControls(src) {
			t.Fatalf("HasControls(%q) = false, want true", src)
		}
	}

	for _, src := range []string{
		``,
		`plain text`,
		`<div id="a">hi</div>`,
		`<inpu id="t">`,
	} {
		if HasControls(src) {
			t.Fatalf("HasControls(%q) = true, want false", src)
		}
	}
}

func TestScanControls(t *testing.T) {
	t.Parallel()

	want := []string{
		`<input id="t" value="Hi&amp;">`,
		`<INPUT ID="p" TYPE="PassWord" VALUE="s">`,
		`<input id="c" type="checkbox" name="ok" checked="">`,
		`<input type="radio" id="r" name="g" value="1">`,
		`<input id="f" type='file' value=a.txt>`,
		`<TEXTAREA ID="m">a&amp;b</TEXTAREA>`,
		`<select id="s" name="c"><option value="one">One</option><option value="two" selected>Two</option></select>`,
		`<select id="g"><optgroup><option>A&amp;</option><option value="b">B</option></optgroup></select>`,
	}
	src := strings.Join(want, "") +
		`<STYLE><input id="no"></STYLE><script><input id="zz"></script>` +
		`<input value="missing"><input id="b" type="button">` +
		`<!-- <input id="z"> -->`
	got := Scan(src)
	if len(got) != len(want) {
		t.Fatal(len(got))
	}
	for i, sp := range got {
		if src[sp.Start:sp.End] != want[i] {
			t.Fatalf("%d %q", i, src[sp.Start:sp.End])
		}
	}

	chk := func(i int, tag, kind, name, val string, on bool) {
		t.Helper()
		c := got[i].Control
		if c.Tag != tag || c.Type != kind || c.Name != name || c.Value != val || c.Checked != on {
			t.Fatalf("%d %#v", i, c)
		}
	}
	chk(0, "input", "", "", "Hi&", false)
	chk(1, "input", "password", "", "s", false)
	chk(2, "input", "checkbox", "ok", "", true)
	chk(3, "input", "radio", "g", "1", false)
	chk(4, "input", "file", "", "a.txt", false)
	chk(5, "textarea", "", "", "a&b", false)
	chk(6, "select", "", "c", "two", false)
	chk(7, "select", "", "", "A&", false)

	opt := got[6].Control.Options
	group := got[7].Control.Options
	if len(opt) != 2 || opt[0].Value != "one" || opt[0].Label != "One" || opt[0].Selected ||
		opt[1].Value != "two" || opt[1].Label != "Two" || !opt[1].Selected {
		t.Fatalf("opt %#v", opt)
	}
	if len(group) != 2 || group[0].Label != "A&" || group[0].Value != "A&" || !group[0].Selected ||
		group[1].Value != "b" || group[1].Selected {
		t.Fatalf("group %#v", group)
	}
}

func TestGenericTabindexTargets(t *testing.T) {
	const src = `<div id="slider" role="slider" tabindex="0"><button id="child">Go</button></div><section id="panel" tabindex="-1"><input id="field" value="live"></section><div id="not-focusable"></div><div id="bad" tabindex="invalid"></div><script><div id="script" tabindex="0"></div></script>`
	for _, scan := range []func(string) []Span{Scan, func(s string) []Span { sp, _ := ScanSingle(s); return sp }} {
		spans := scan(src)
		if len(spans) != 4 {
			t.Fatalf("controls %+v", spans)
		}
		if spans[0].Control.ID != "slider" || CanEdit(spans[0].Control) || spans[2].Control.TabIndex != -1 {
			t.Fatal("generic tabindex treated as editable or lost negative index")
		}
		out := RewriteState(src, spans, nil, Caret{Focus: "field", Caret: 4})
		if !strings.Contains(out, `id="slider" role="slider" tabindex="0"`) || !strings.Contains(out, `id="child"`) || !strings.Contains(out, `data-ownframe-field`) {
			t.Fatal("generic focus wrapper damaged nested controls")
		}
	}
}
