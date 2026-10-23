package page

import (
	"context"
	"testing"
)

func TestSelectAtLandsOnGlyph(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="1234567890"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "1234567890")
	left, width := runSpan(p, op)
	base := runBase(p, op)
	want := []int{2, 5, 8}

	for i, frac := range []float64{0.2, 0.5, 0.8} {
		if err := p.SelectAt(ctx, left+width*frac, base-4); err != nil {
			t.Fatal(err)
		}

		st := p.formState().snap()
		if p.FocusID() != "t" || st.Caret != want[i] || st.Anchor != want[i] {
			t.Fatalf("frac %.1f focus %q caret %d anchor %d", frac, p.FocusID(), st.Caret, st.Anchor)
		}
	}
}

func TestDragExtendsRange(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="1234567890"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "1234567890")
	left, width := runSpan(p, op)
	base := runBase(p, op)

	if err := p.SelectAt(ctx, left+width*0.2, base-4); err != nil {
		t.Fatal(err)
	}

	if err := p.Drag(ctx, left+width*0.8, base-4); err != nil {
		t.Fatal(err)
	}

	st := p.formState().snap()
	start, end := st.Bounds(10)
	if start != 2 || end != 8 || st.Caret != 8 || st.Anchor != 2 {
		t.Fatalf("range [%d,%d) caret %d anchor %d", start, end, st.Caret, st.Anchor)
	}

	if p.FormSelected("t") {
		t.Fatal("partial range reports selected")
	}

	if err := p.Drag(ctx, left+width*0.4, base-4); err != nil {
		t.Fatal(err)
	}

	st = p.formState().snap()
	if start, end := st.Bounds(10); start != 2 || end != 4 {
		t.Fatalf("second drag [%d,%d)", start, end)
	}
}
