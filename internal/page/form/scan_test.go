package form

import (
	"strings"
	"testing"
)

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
