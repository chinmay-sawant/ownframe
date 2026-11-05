package click

import (
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestLabelForMatchesIDAndText(t *testing.T) {
	t.Parallel()

	source := `<div><label id="lab" for="email">Email<span>*</span></label><input id="email"></div>`
	if got := LabelFor(source, "lab", ""); got != "email" {
		t.Fatalf("id match %q", got)
	}
	if got := LabelFor(source, "", "Email*"); got != "email" {
		t.Fatalf("text match %q", got)
	}
	if got := LabelFor(source, "", "Other"); got != "" {
		t.Fatalf("miss %q", got)
	}
	if got := LabelFor(`<label>Email</label>`, "", "Email"); got != "" {
		t.Fatalf("missing for %q", got)
	}
}

func TestLabelControlWalksToLabel(t *testing.T) {
	t.Parallel()

	boxes := []layout.Box{
		{Tag: "div", ParentIndex: -1},
		{Tag: "label", Text: "Email*", ParentIndex: 0},
		{Tag: "span", Text: "*", ParentIndex: 1},
	}
	source := `<label for="email">Email<span>*</span></label>`
	if got := LabelControl(source, boxes, boxes[2]); got != "email" {
		t.Fatalf("walk %q", got)
	}
}
