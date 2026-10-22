package form

import (
	"strings"
	"testing"
)

func TestRewriteKeepsAuthorAttrs(t *testing.T) {
	t.Parallel()

	src := `<input id="e" class="fancy" style="color:red" placeholder="Email" name="e" value="x" data-ownframe-focus="1">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Name: "e", Value: "x"}
	got := Rewrite(src, []Span{whole(src, ctrl)}, nil, "", false)
	tag := openOf(got, "span")
	if !strings.Contains(tag, `class="fancy"`) || !strings.Contains(tag, `style="color:red"`) {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(tag, `placeholder="Email"`) || !strings.Contains(tag, `name="e"`) {
		t.Fatalf("%s", got)
	}
	if strings.Contains(tag, "value=") || strings.Contains(tag, "type=") || strings.Contains(tag, focusAttr) {
		t.Fatalf("%s", got)
	}
}

func TestRewritePlaceholderEmpty(t *testing.T) {
	t.Parallel()

	src := `<input id="e" type="text" placeholder="Email">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text"}
	got := Rewrite(src, []Span{whole(src, ctrl)}, nil, "", false)
	if !strings.Contains(openOf(got, "span"), placeholderAttr) || !strings.Contains(got, ">Email</span>") {
		t.Fatalf("%s", got)
	}
	if strings.Contains(spanPart(got), focusAttr) || strings.Contains(spanPart(got), "data-ownframe-caret") {
		t.Fatalf("%s", got)
	}
}

func TestRewriteCaretBeforePlaceholder(t *testing.T) {
	t.Parallel()

	src := `<input id="e" type="text" placeholder="Email">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text"}
	got := Rewrite(src, []Span{whole(src, ctrl)}, nil, "e", false)
	if !strings.Contains(got, `>Email</span>`) || strings.Contains(got, "data-ownframe-caret") {
		t.Fatalf("%s", got)
	}
}
