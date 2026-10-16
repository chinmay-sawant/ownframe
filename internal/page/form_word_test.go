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

	if start, end := p.formState().bounds(11); start != 0 || end != 5 {
		t.Fatalf("first word [%d,%d)", start, end)
	}

	if err := p.SelectWordAt(ctx, left+width*0.6, base-4); err != nil {
		t.Fatal(err)
	}

	if start, end := p.formState().bounds(11); start != 6 || end != 11 {
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

	if start, end := p.formState().bounds(7); start != 4 || end != 7 {
		t.Fatalf("line [%d,%d)", start, end)
	}
}

func TestRangeClampsAfterRedraw(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	p := bftPage(t, `<input id="t" type="text" value="abcdef"`+bedWide+`>`)
	bftDraw(t, p, nil)
	bftClick(t, p, "t")
	p.formState().anchor, p.formState().caret = 2, 5

	c := p.formState().byID["t"]
	c.Value = "ab"
	p.formState().byID["t"] = c

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	if p.formState().caret != 2 || p.formState().anchor != 2 {
		t.Fatalf("caret %d anchor %d after clamp", p.formState().caret, p.formState().anchor)
	}

	if start, end := p.formState().bounds(2); start != 2 || end != 2 {
		t.Fatalf("range [%d,%d) after clamp", start, end)
	}
}
