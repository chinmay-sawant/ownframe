package main

import (
	"context"
	"math"
	"testing"
)

// TestPlayAllMovesEverySample checks that the Play all button gives every
// card the same result as its own Start button: the sample moves off its
// starting paint and has not reached its end yet.
func TestPlayAllMovesEverySample(t *testing.T) {
	ctx := context.Background()
	p, d, v := fixturePage(t)
	button, ok := findBox(p.Boxes(), playAllID)
	if !ok {
		t.Fatal("the play all button is missing")
	}
	if err := p.Click(ctx, button.X+button.W/2, button.Y+button.H/2); err != nil {
		t.Fatal(err)
	}
	if err := d.tick(ctx); err != nil {
		t.Fatal(err)
	}

	for _, row := range v.Catalog {
		s := d.samples[row.SampleKey]
		if s == nil || !s.active || len(s.plan.pairs) == 0 {
			tt, _ := targetByKey(row.SampleKey)
			bb, found := findBox(p.Boxes(), tt.id)
			t.Errorf("%s has no running sample: nil=%v active=%v pairs=%d dur=%v base=%d detached=%v found=%v anims=%d trans=%d stats=%+v",
				row.Name, s == nil, s != nil && s.active, len(s.plan.pairs), s.plan.duration,
				len(s.base), s != nil && s.detached, found, len(bb.Animations), len(bb.Transitions), p.Probe().Stats())
			continue
		}
		target, _ := targetByKey(row.SampleKey)
		pair := s.plan.pairs[0]
		base, to := pair.from, pair.to
		at := s.origin.Add(s.plan.delay + s.plan.duration/2)
		s.plan.step(at)
		got := read(pair.op)
		if samePaintValue(got, base) {
			t.Errorf("%s did not move off its starting paint", row.Name)
		}
		if target.kind == "transition" && samePaintValue(got, to) {
			t.Errorf("%s jumped straight to its end paint", row.Name)
		}
	}
}

// samePaintValue reports whether two paints match.
func samePaintValue(a, b paint) bool {
	return math.Abs(a.X-b.X) < .01 && math.Abs(a.Y-b.Y) < .01 &&
		math.Abs(a.W-b.W) < .01 && math.Abs(a.H-b.H) < .01 &&
		math.Abs(a.R-b.R) < .01 && math.Abs(a.G-b.G) < .01 && math.Abs(a.B-b.B) < .01 &&
		math.Abs(a.Opacity-b.Opacity) < .01
}
