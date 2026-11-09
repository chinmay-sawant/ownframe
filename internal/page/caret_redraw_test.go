package page

import (
	"context"
	"testing"
)

func TestCaretClickMovesPaintedBar(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const value = "1234567890"
	p := bftPage(t, `<input id="t" type="text" value="`+value+`"`+bedWide+`>`)
	bftDraw(t, p, nil)

	op := runOp(p, value)
	if op == nil {
		t.Fatal("missing text run")
	}

	left, width := runSpan(p, op)
	base := runBase(p, op)
	box := p.boxByID("t")
	padX := box.X + box.W - 4
	if padX <= left+width {
		t.Fatalf("field has no empty padding: text ends at %.1f, box ends at %.1f", left+width, box.X+box.W)
	}

	if err := p.SelectAt(ctx, padX, base-4); err != nil {
		t.Fatal(err)
	}

	st := p.formState()
	n := runeLen(value)
	if p.FocusID() != "t" || st.Caret != n || st.Anchor != n {
		t.Fatalf("padding focus %q caret %d anchor %d", p.FocusID(), st.Caret, st.Anchor)
	}

	endX := paintedCaretX(t, p)
	focused := runOp(p, value)
	if focused == nil {
		t.Fatal("missing focused text run")
	}

	runEnd := focused.X + focused.W
	tol := 2 * p.draw.Display.PointsPerPixel
	delta := endX - runEnd
	if delta < 0 {
		delta = -delta
	}
	if delta > tol {
		t.Fatalf("padding caret x %.2f, run ends at %.2f", endX, runEnd)
	}

	if err := p.SelectAt(ctx, left+width*0.2, base-4); err != nil {
		t.Fatal(err)
	}

	st = p.formState()
	if p.FocusID() != "t" || st.Caret != 2 || st.Anchor != 2 {
		t.Fatalf("glyph focus %q caret %d anchor %d", p.FocusID(), st.Caret, st.Anchor)
	}

	glyphX := paintedCaretX(t, p)
	if glyphX >= endX {
		t.Fatalf("painted caret x %.2f did not move left of %.2f", glyphX, endX)
	}
}

func paintedCaretX(t *testing.T, p *Page) float64 {
	t.Helper()

	d := p.draw.Display
	if d == nil || len(d.Ops) == 0 {
		t.Fatal("no display")
	}

	op := d.Ops[len(d.Ops)-1]
	if op.Kind != DisplayOpFillRect || op.W > d.PointsPerPixel {
		t.Fatalf("caret kind %v width %.2f", op.Kind, op.W)
	}

	return op.X
}
