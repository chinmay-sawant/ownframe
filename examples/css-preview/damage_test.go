package main

import (
	"context"
	"testing"
)

func TestAnimationTickMarksPartialDamage(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	row := v.Catalog[14] // animation-name, an untriggered keyframe sample
	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing %s", row.Button)
	}
	if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	// Drop the click's own damage so the tick's marks stand alone.
	p.Probe().TakeDirtyRects()

	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if !p.Probe().FrameDirty() {
		t.Fatal("animation tick did not opt into the frame-dirty path")
	}
	rects, ok := p.Probe().TakeDirtyRects()
	if !ok || len(rects) == 0 {
		t.Fatal("animation tick marked no damage")
	}
	for _, rect := range rects {
		if rect.Dx() > 400 || rect.Dy() > 400 {
			t.Fatalf("tick damage %v is the whole page, want the animated card", rect)
		}
	}
}

func TestFinishedTickDropsTheCallback(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	row := v.Catalog[14]
	button, ok := findBox(p.Boxes(), row.Button)
	if !ok {
		t.Fatalf("missing %s", row.Button)
	}
	if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	for _, s := range d.samples {
		s.plan = motion{}
		s.active = true
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}
	if p.Probe().Ticking() {
		t.Fatal("finished tick kept the callback")
	}
}
