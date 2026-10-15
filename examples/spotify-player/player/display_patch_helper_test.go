package player

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe"
)

func TestDisplayPatchHelperPatchOp(t *testing.T) {
	app := mustApp(t)
	ctx := context.Background()
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if app.Page().Display() == nil {
		t.Fatal("no display")
	}

	h := NewDisplayPatchHelper(app)
	if !h.PatchOp(0, func(o *ownframe.DisplayOp) { o.Alpha = 0.5 }) {
		t.Fatal("PatchOp failed")
	}
	report := h.Report()
	if report.Changed != 1 || report.Full {
		t.Fatalf("report = %+v", report)
	}
	rects, ok := app.Page().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("no dirty rects: %v %v", rects, ok)
	}
	if got := app.PatchReport(); got.Changed != 1 {
		t.Fatalf("app patch report = %+v", got)
	}
	if h.PatchOp(1000000, func(*ownframe.DisplayOp) {}) {
		t.Fatal("out-of-range PatchOp reported success")
	}
}

func TestDisplayPatchHelperPatchDisplay(t *testing.T) {
	app := mustApp(t)
	ctx := context.Background()
	if err := app.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	if app.Page().Display() == nil {
		t.Fatal("no display")
	}
	app.Page().TakeDirtyRects()

	h := app.DisplayPatchHelper()
	report := h.PatchDisplay(func(d *ownframe.Display) {
		if len(d.Ops) == 0 {
			t.Fatal("no ops")
		}
		d.Ops[0].X += 4
	})
	if report.Changed != 1 {
		t.Fatalf("report = %+v", report)
	}
	rects, ok := app.Page().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatalf("no dirty rects: %v %v", rects, ok)
	}
}
