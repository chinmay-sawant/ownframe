package page

import (
	"context"
	"testing"
)

func TestSelectWordAt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="hello world"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, "hello world")
	left, width := runSpan(p, op)
	base := runBase(p, op)

	if err := p.SelectWordAt(ctx, left+width*0.1, base-4); err != nil {
		t.Fatal(err)
	}

	st := p.formState().snap()
	start, end := st.Bounds(11)
	if start != 0 || end != 5 {
		t.Fatalf("first word [%d,%d)", start, end)
	}

	if err := p.SelectWordAt(ctx, left+width*0.6, base-4); err != nil {
		t.Fatal(err)
	}

	st = p.formState().snap()
	start, end = st.Bounds(11)
	if start != 6 || end != 11 {
		t.Fatalf("second word [%d,%d)", start, end)
	}
}

func TestSelectLineAt(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<textarea id="t" style="display:block;width:220px;height:60px">one
two</textarea>`)
	bftDraw(t, p, nil)

	op := runOp(p, "two")
	if op == nil {
		t.Fatal("no second line run")
	}

	left, width := runSpan(p, op)
	base := runBase(p, op)

	if err := p.SelectLineAt(ctx, left+width/2, base-4); err != nil {
		t.Fatal(err)
	}

	st := p.formState().snap()
	if start, end := st.Bounds(7); start != 4 || end != 7 {
		t.Fatalf("line [%d,%d)", start, end)
	}
}

func TestRangeClampsAfterRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcdef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	st := p.formState().snap()
	st.Anchor, st.Caret = 2, 5
	c := st.ByID["t"]
	c.Value = "ab"
	st.ByID["t"] = c
	p.formState().store(&st)

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	st = p.formState().snap()
	if st.Caret != 2 || st.Anchor != 2 {
		t.Fatalf("caret %d anchor %d after clamp", st.Caret, st.Anchor)
	}

	if start, end := st.Bounds(2); start != 2 || end != 2 {
		t.Fatalf("range [%d,%d) after clamp", start, end)
	}
}
