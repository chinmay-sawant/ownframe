package form

import (
	"strings"
	"testing"
)

func TestRewriteOverlap(t *testing.T) {
	t.Parallel()

	src := "0123456789"
	spans := []Span{
		{Control: Control{ID: "c", Type: "text", Value: "CC"}, Start: 6, End: 10},
		{Control: Control{ID: "z", Type: "text", Value: "ZZ"}, Start: -1, End: 2},
		{Control: Control{ID: "a", Type: "text", Value: "AA"}, Start: 0, End: 4},
		{Control: Control{ID: "b", Type: "text", Value: "BB"}, Start: 2, End: 6},
		{Control: Control{ID: "y", Type: "text", Value: "YY"}, Start: 8, End: 20},
	}
	got := Rewrite(src, spans, nil, "a", false)
	if strings.Contains(got, "BB") || strings.Contains(got, "ZZ") || strings.Contains(got, "YY") {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(got, `>AA</span>45`) ||
		!strings.Contains(got, ">CC</span>") {
		t.Fatalf("%s", got)
	}
	if !strings.Contains(openOf(got, "span"), focusAttr) {
		t.Fatalf("%s", got)
	}
}
