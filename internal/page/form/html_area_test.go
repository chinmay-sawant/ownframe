package form

import (
	"strings"
	"testing"
)

func TestRewriteTextareaFocus(t *testing.T) {
	t.Parallel()

	src := `<textarea id="note" name="n">old</textarea>`
	ctrl := Control{ID: "note", Tag: "textarea", Name: "n", Value: `a<b&c"`}
	got := Rewrite(src, []Span{whole(src, ctrl)}, nil, "note", false)
	tag := openOf(got, "span")
	if !strings.Contains(tag, `data-ownframe-field="textarea"`) ||
		!strings.Contains(tag, `data-ownframe-focus="1"`) || !strings.Contains(tag, `name="n"`) {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(got, `>a&lt;b&amp;c&#34;</span>`) {
		t.Fatalf("%s", got)
	}
}
