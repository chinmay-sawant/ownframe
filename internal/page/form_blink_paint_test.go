package page

import (
	"context"
	"strings"
	"testing"
)

func TestWholeSelectionPaintsASelectionSpan(t *testing.T) {
	src := `<input id="e" type="text" value="x">`
	ctrl := Control{ID: "e", Tag: "input", Type: "text", Value: "x"}
	got := rewriteControls(src, []controlSpan{{Control: ctrl, Start: 0, End: len(src)}}, nil, "e", true)
	part := got
	if i := strings.Index(got, "<span"); i >= 0 {
		part = got[i:]
	}
	if !strings.Contains(part, `<span data-ownframe-selection="1">x</span>`) {
		t.Fatalf("whole selection has no selection span: %s", got)
	}
}

func TestCaretUsesTheTextColor(t *testing.T) {
	ctx := context.Background()
	p, err := New(Config{
		HTML:   `<input id="e" type="text" value="ab" style="color:#ff0000">`,
		Width:  300,
		Height: 100,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if err := p.Focus(ctx, "e"); err != nil {
		t.Fatal(err)
	}

	d := p.Display()
	if d == nil {
		t.Fatal("no display")
	}

	found := false
	for i := range d.Ops {
		op := &d.Ops[i]
		if op.W >= 2 || op.H <= 5 || op.H > 20 {
			continue
		}

		found = true

		if op.R < 0.99 || op.G > 0.01 || op.B > 0.01 {
			t.Fatalf("caret color = %.3f,%.3f,%.3f, want the text red", op.R, op.G, op.B)
		}
	}

	if !found {
		t.Fatal("no caret op")
	}
}
