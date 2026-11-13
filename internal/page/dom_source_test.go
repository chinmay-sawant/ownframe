package page

import (
	"context"
	"testing"
)

func TestSelectControls(t *testing.T) {
	t.Parallel()

	p, err := New(Config{
		HTML:   `<select id="c"><option value="a">A</option><option value="b" selected>B</option></select>`,
		Width:  120,
		Height: 80,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	got := p.SelectControls()
	opts, ok := got["c"]
	if !ok || len(opts) != 2 {
		t.Fatalf("controls = %v", got)
	}
	if opts[0].Value != "a" || opts[0].Label != "A" || opts[0].Selected {
		t.Fatalf("opts = %+v", opts)
	}
	if opts[1].Value != "b" || !opts[1].Selected {
		t.Fatalf("opts = %+v", opts)
	}
}

func TestPlaceCaretInsertsAtOffset(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<input id="t" type="text">`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	p.SetFormValue("t", "ab")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if err := p.Focus(ctx, "t"); err != nil {
		t.Fatal(err)
	}
	// A caret for another field is ignored.
	p.PlaceCaret("missing", 0)
	p.PlaceCaret("t", 0)
	if err := p.Type(ctx, "X"); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("t"); got != "Xab" {
		t.Fatalf("value = %q", got)
	}
	// Out-of-range offsets clamp to the end.
	p.PlaceCaret("t", 999)
	if err := p.Type(ctx, "Y"); err != nil {
		t.Fatal(err)
	}
	if got := p.FormValue("t"); got != "XabY" {
		t.Fatalf("value = %q", got)
	}
}

func TestSelectControlsEmpty(t *testing.T) {
	t.Parallel()

	p, err := New(Config{HTML: `<p>hi</p>`, Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	if got := p.SelectControls(); len(got) != 0 {
		t.Fatalf("controls = %v", got)
	}
	if p.ExecutedSource() == "" || p.ThemeSource() != "" {
		t.Fatalf("source empty or theme set")
	}
}
