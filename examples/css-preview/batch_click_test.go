package main

import (
	"context"
	"testing"
)

// TestBatchedPressStartsPreview runs the real press, focus, and click sequence
// in one redraw batch and checks the animation still binds to the target.
func TestBatchedPressStartsPreview(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	row := v.Catalog[14]
	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing %q", row.Button)
	}
	x, y := button.X+button.W/2, button.Y+button.H/2

	err := redrawBatch(ctx, p, func() error {
		if err := p.Testing().Press(ctx, x, y); err != nil {
			return err
		}
		if err := p.SelectAt(ctx, x, y); err != nil {
			return err
		}
		return p.Testing().Click(ctx, x, y)
	})
	if err != nil {
		t.Fatal(err)
	}

	if got := p.FocusID(); got != row.Button {
		t.Fatalf("focus = %q, want %q", got, row.Button)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	sample := d.samples[row.SampleKey]
	if sample == nil || len(sample.plan.pairs) == 0 {
		t.Fatalf("%s did not bind after the batched press", row.Name)
	}
}
