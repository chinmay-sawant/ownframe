package page

import (
	"context"
	"testing"
)

// TestClickRepaintsOneBox clicks a counter whose handler declares the changed
// id, then checks that every operation the click changed lies inside the
// dirty rect. Operations outside it must be byte-identical to the previous
// display list.
func TestClickRepaintsOneBox(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	n := 0

	p, err := New(Config{
		HTML:   `<body style="margin:0"><div id="count"><span id="num">{{.N}}</span></div><div id="inc" data-action="add">+</div></body>`,
		Width:  320,
		Height: 200,
	})
	if err != nil {
		t.Fatal(err)
	}

	p.SetData(struct{ N int }{0})
	p.Handle(Handlers{Click: func(_ context.Context, box Box) error {
		if box.Action == "add" {
			n++
			p.SetData(struct{ N int }{n})
			p.Invalidate("count")
		}

		return nil
	}})

	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}

	p.TakeDirty()
	before := p.draw.Display

	x, y := clickCenterOf(t, p, "inc")
	if err := p.Click(ctx, x, y); err != nil {
		t.Fatal(err)
	}

	rect, ok := p.TakeDirty()
	if !ok {
		t.Fatal("no dirty rect after the click")
	}

	after := p.draw.Display
	if before == nil || after == nil || len(before.Ops) != len(after.Ops) {
		t.Fatalf("op counts = %d then %d", opCount(before), opCount(after))
	}

	for i := range after.Ops {
		if sameOp(&before.Ops[i], &after.Ops[i]) {
			continue
		}

		if opBounds(&after.Ops[i], after).Intersect(rect).Empty() {
			t.Fatalf("op %d changed outside the dirty rect %v", i, rect)
		}
	}
}

func opCount(d *Display) int {
	if d == nil {
		return -1
	}

	return len(d.Ops)
}

// clickCenterOf returns the center of one box for a click.
func clickCenterOf(t *testing.T, p *Page, id string) (float64, float64) {
	t.Helper()

	for _, box := range p.boxes {
		if box.ID == id {
			return box.X + box.W/2, box.Y + box.H/2
		}
	}

	t.Fatalf("box %q not found", id)

	return 0, 0
}
