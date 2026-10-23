package page

import (
	"context"
	"testing"
)

func TestCaretReplacesSelection(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	st := p.formState().snap()
	st.Anchor, st.Caret = 1, 3
	p.formState().store(&st)

	if err := p.Type(ctx, "Z"); err != nil {
		t.Fatal(err)
	}

	got := p.formState().snap()
	if p.FormValue("t") != "aZd" || got.Caret != 2 || got.Anchor != 2 {
		t.Fatalf("value %q caret %d anchor %d", p.FormValue("t"), got.Caret, got.Anchor)
	}
}

func TestSelectAllSetsFullRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcd"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")

	if err := p.SelectAll(ctx); err != nil {
		t.Fatal(err)
	}

	st := p.formState().snap()
	start, end := st.Bounds(4)
	if start != 0 || end != 4 || st.Caret != 4 || !st.All {
		t.Fatalf("range [%d,%d) caret %d all %v", start, end, st.Caret, st.All)
	}

	if !p.FormSelected("t") {
		t.Fatal("not selected")
	}
}
