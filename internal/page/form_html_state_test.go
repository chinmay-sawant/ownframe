package page

import (
	"strings"
	"testing"
)

func TestRewriteSelectedNoCaret(t *testing.T) {
	t.Parallel()

	src := `<input id="e" type="text" value="x">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: "x"}
	got := rewriteControls(src, []controlSpan{whole(src, ctrl)}, nil, "e", true)
	tag := openOf(got, "span")
	if !strings.Contains(tag, selectedAttr) || !strings.Contains(tag, focusAttr) {
		t.Fatalf("%s", got)
	}
	if strings.Contains(spanPart(got), "data-ownframe-caret") {
		t.Fatalf("%s", got)
	}
}

func TestRewriteFocusedValueAsOneTextRun(t *testing.T) {
	t.Parallel()

	src := `<input id="e" type="text" value="x">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: "x"}
	got := rewriteControls(src, []controlSpan{whole(src, ctrl)}, nil, "e", false)
	if !strings.Contains(got, `>x</span>`) || strings.Contains(got, "data-ownframe-caret") {
		t.Fatalf("%s", got)
	}
	got = rewriteControls(src, []controlSpan{whole(src, ctrl)}, nil, "", false)
	if strings.Contains(spanPart(got), "data-ownframe-caret") || strings.Contains(spanPart(got), focusAttr) {
		t.Fatalf("%s", got)
	}
}

func TestRewriteCaretMidValue(t *testing.T) {
	t.Parallel()

	src := `<input id="e" type="text" value="abcd">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: "abcd"}
	st := caretState{focus: "e", caret: 2, start: 2, end: 2}
	got := rewriteControlsState(src, []controlSpan{whole(src, ctrl)}, nil, st)
	if !strings.Contains(got, `>abcd</span>`) {
		t.Fatalf("%s", got)
	}
}

func TestRewriteRange(t *testing.T) {
	t.Parallel()

	src := `<input id="e" type="text" value="abcd">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: "abcd"}
	st := caretState{focus: "e", caret: 3, start: 1, end: 3}
	got := rewriteControlsState(src, []controlSpan{whole(src, ctrl)}, nil, st)
	if !strings.Contains(got, `>a<span data-ownframe-selection="1">bc</span>d</span>`) {
		t.Fatalf("%s", got)
	}
}
