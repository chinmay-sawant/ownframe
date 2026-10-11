package page

import (
	"context"
	"testing"
)

func TestDragAcrossWrappedLines(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	value := "aaaa bbbb cccc dddd eeee ffff gggg hhhh iiii jjjj"
	p := bftPage(t, `<input id="t" type="text" value="`+value+`" style="display:block;width:120px;height:80px">`)
	bftDraw(t, p, nil)

	box := p.boxByID("t")
	pt := p.display.PointsPerPixel

	var first, last *DisplayOp
	for i := range p.display.Ops {
		op := &p.display.Ops[i]
		if op.Kind != DisplayOpText || op.Text == "" || op.Font == nil {
			continue
		}

		if op.Y < box.Y*pt || op.Y > (box.Y+box.H)*pt {
			continue
		}

		if first == nil || op.Y < first.Y {
			first = op
		}

		if last == nil || op.Y > last.Y {
			last = op
		}
	}

	if first == nil || last == nil || first == last {
		t.Fatal("value did not wrap")
	}

	if err := p.SelectAt(ctx, first.X/pt, first.Y/pt-4); err != nil {
		t.Fatal(err)
	}

	if err := p.Drag(ctx, (last.X+last.W)/pt+4, last.Y/pt-4); err != nil {
		t.Fatal(err)
	}

	start, end := p.form.bounds(runeLen(value))
	if start != 0 || end != runeLen(value) {
		t.Fatalf("wrapped range [%d,%d) of %d", start, end, runeLen(value))
	}

	if !p.FormSelected("t") {
		t.Fatal("whole wrapped value not selected")
	}
}
