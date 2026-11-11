package page

import (
	"context"
	"testing"
)

func TestCaretSurvivesRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcdef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	p.setCaret(4, false)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	st := p.formState()
	if st.Caret != 4 || st.Anchor != 4 || p.FocusID() != "t" {
		t.Fatalf("caret %d anchor %d focus %q", st.Caret, st.Anchor, p.FocusID())
	}
}

func TestCaretAndSelectionSurviveResize(t *testing.T) {
	t.Parallel()
	p := bftPage(t, `<input id="t" type="text" value="abcdef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	p.setCaret(2, false)
	p.setCaret(4, true)
	p.SetSize(600, 400)
	if err := p.Redraw(t.Context()); err != nil {
		t.Fatal(err)
	}
	st := p.formState()
	if st.Caret != 4 || st.Anchor != 2 || p.FocusID() != "t" || p.FormValue("t") != "abcdef" {
		t.Fatal("resize changed the live field, caret, selection, or focus")
	}
}
