package main

import (
	"context"
	"testing"
)

// TestBatchedPressStartsPreview runs the real press, focus, and click sequence
// in one redraw batch and checks the animation binds to the target.
func TestBatchedPressStartsPreview(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	row := v.Catalog[15]
	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing %q", row.Button)
	}
	x, y := button.X+button.W/2, button.Y+button.H/2

	err := redrawBatch(ctx, p, func() error {
		if err := p.Edit().Press(ctx, x, y); err != nil {
			return err
		}
		if err := p.Testing().SelectAt(ctx, x, y); err != nil {
			return err
		}
		return p.Click(ctx, x, y)
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := p.Form().FocusID(); got != row.Button {
		t.Fatalf("focus = %q, want %q", got, row.Button)
	}
	if !v.Started[row.Button] {
		t.Fatal("the batched click did not record the Start")
	}
	if d.manual[row.SampleKey] != nil {
		t.Fatal("an animation card took the manual path")
	}
	tickPage(t, p)
	if got := xformX(targetOp(t, p, row.TargetID), p.Probe().Display().PointsPerPixel); got == 0 {
		t.Fatalf("%s did not bind after the batched press", row.Name)
	}
}
