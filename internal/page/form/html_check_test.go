package form

import (
	"strings"
	"testing"
)

func TestRewriteCheckbox(t *testing.T) {
	t.Parallel()

	off := `<input class="c" type="checkbox" id="ok" checked disabled>`
	ctrl := Control{ID: "ok", Tag: "input", Type: "checkbox"}
	got := Rewrite(off, []Span{whole(off, ctrl)}, nil, "", false)
	tag := openOf(got, "input")
	if strings.Contains(tag, "checked") || strings.Contains(tag, "disabled") || !strings.Contains(tag, `class="c"`) {
		t.Fatalf("%s", got)
	}

	on := `<input class="c" type="checkbox" id="ok" value="yes">`
	live := map[string]Control{
		"ok": {Tag: "input", Type: "checkbox", Checked: true, Disabled: true},
	}
	got = Rewrite(on, []Span{whole(on, ctrl)}, live, "ok", false)
	tag = openOf(got, "input")
	if !strings.Contains(tag, "checked") || !strings.Contains(tag, "disabled") {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(tag, `type="checkbox"`) || !strings.Contains(tag, `value="yes"`) {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(tag, `data-ownframe-focus="1"`) {
		t.Fatalf("%s", got)
	}
}
