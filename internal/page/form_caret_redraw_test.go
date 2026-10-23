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
