package main

import (
	"context"
	"testing"
)

// TestWalkResetClearsReadout checks that a reset drops the readout the
// click patched into the tree. The after-redraw hook patches it back into
// every rebuilt tree, so a stale copy would keep it on screen.
func TestWalkResetClearsReadout(t *testing.T) {
	p, d, v := fixturePage(t)
	w := &walker{p: p, d: d, v: v}
	d.readout = "opacity = 1\nIt fades."
	v.Readout = d.readout
	w.reset(context.Background())
	if d.readout != "" {
		t.Errorf("demo readout = %q, want it cleared", d.readout)
	}
	if v.Readout != "" {
		t.Errorf("view readout = %q, want it cleared", v.Readout)
	}
}

// TestWalkActiveFalseValue checks that a false value in the environment
// turns the harness off.
func TestWalkActiveFalseValue(t *testing.T) {
	for _, v := range []string{"0", "f", "false", "FALSE", "False"} {
		t.Setenv("CSS_PREVIEW_WALK", v)
		if walkActive() {
			t.Errorf("walkActive() = true for CSS_PREVIEW_WALK=%q, want false", v)
		}
	}
	for _, v := range []string{"1", "t", "true", "yes"} {
		t.Setenv("CSS_PREVIEW_WALK", v)
		if !walkActive() {
			t.Errorf("walkActive() = false for CSS_PREVIEW_WALK=%q, want true", v)
		}
	}
	t.Setenv("CSS_PREVIEW_WALK", "")
	if walkActive() {
		t.Error("walkActive() = true for an empty CSS_PREVIEW_WALK, want false")
	}
}
