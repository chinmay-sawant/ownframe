package page

import (
	"context"
	"testing"
)

func TestPatchDisplayRecordsDirty(t *testing.T) {
	p, err := New(Config{HTML: `<div id=x>hello</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()

	patch := p.PatchDisplay(func(d *Display) {
		if len(d.Ops) == 0 {
			t.Fatal("no ops")
		}
		d.Ops[0].X += 10
	})

	if patch.Changed != 1 || patch.Full {
		t.Fatalf("patch = %+v", patch)
	}
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("no dirty rects: %v %v", rects, ok)
	}
}

func TestPatchOpRecordsDirty(t *testing.T) {
	p, err := New(Config{HTML: `<div id=x>hello</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}

	if !p.PatchOp(0, func(o *DisplayOp) { o.Alpha = 0.5 }) {
		t.Fatal("patch failed")
	}
	if p.LastPatch().Changed != 1 {
		t.Fatalf("patch = %+v", p.LastPatch())
	}
	rects, ok := p.TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("no dirty rects: %v %v", rects, ok)
	}
}

func TestPatchDisplayIdentityNoDirty(t *testing.T) {
	p, err := New(Config{HTML: `<div id=x>hello</div>`, Width: 320, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	p.TakeDirtyRects()

	patch := p.PatchDisplay(func(*Display) {})
	if patch.Changed != 0 || patch.Full {
		t.Fatalf("patch = %+v", patch)
	}
	rects, ok := p.TakeDirtyRects()
	if ok || len(rects) != 0 {
		t.Fatalf("unexpected dirty: %v %v", rects, ok)
	}
}
