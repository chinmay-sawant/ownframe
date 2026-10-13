package runtimebehavior

import "testing"

func TestTransitionCoverage(t *testing.T) {
	tr := Transition{}
	tr.Begin("hidden", "visible", 100)
	if !tr.Active || tr.Phase != "running" {
		t.Fatal("begin broken")
	}
	tr.Cancel()
	if tr.Active {
		t.Fatal("cancel broken")
	}
	pt := PhaseTransition{}
	pt.StartPhase("mount")
	if !pt.Active || pt.PhaseName() != "mount" {
		t.Fatal("phase transition broken")
	}
	pt.End()
	if pt.Active {
		t.Fatal("end broken")
	}
	c := ClipRect{X: 0, Y: 0, W: 10, H: 10}
	c.ScrollBy(2, 3)
	x, y := c.Offset()
	if x != 2 || y != 3 {
		t.Fatal("scroll broken")
	}
	c.ResetScroll()
	if c.ScrollX != 0 || c.ScrollY != 0 {
		t.Fatal("reset scroll broken")
	}
}
