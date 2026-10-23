package page

import (
	"context"
	"testing"
)

func tickBarPage(t *testing.T) *Page {
	t.Helper()

	p, err := New(Config{
		HTML:   `<body style="margin:0"><div id="bar" style="width:40px;height:12px;background:#224466"></div></body>`,
		Width:  80,
		Height: 40,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirty()
	if p.FrameDirty() {
		t.Fatal("Redraw left FrameDirty set")
	}

	return p
}

func TestTickFillPatchSetsFrameDirty(t *testing.T) {
	t.Parallel()

	p := tickBarPage(t)
	patched := false
	p.SetTick(func(context.Context) error {
		d := p.Display()
		if d == nil {
			return nil
		}
		bar := boxRect(p.boxByID("bar"))
		for i := range d.Ops {
			op := &d.Ops[i]
			if op.Kind != DisplayOpFillRect || op.W <= 1 || op.H <= 1 {
				continue
			}
			if opBounds(op, d).Overlaps(bar) {
				op.R = 1
				patched = true
				break
			}
		}

		return nil
	})
	if err := p.Tick(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !patched {
		t.Fatal("tick did not patch a fill")
	}
	if !p.FrameDirty() {
		t.Fatal("patched fill must set FrameDirty")
	}
	r, ok := p.TakeDirty()
	if !ok || r.Empty() {
		t.Fatalf("TakeDirty %v ok=%v", r, ok)
	}
	if r.Intersect(boxRect(p.boxByID("bar"))).Empty() {
		t.Fatalf("dirty %v misses bar", r)
	}
}
