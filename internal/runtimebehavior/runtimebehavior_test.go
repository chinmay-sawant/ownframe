package runtimebehavior

import "testing"

func TestMarker(t *testing.T) {
	m := LoadingMarker{}
	m.Show("slow")
	if !m.Visible || m.Label != "slow" {
		t.Fatal("marker broken")
	}
	m.Hide()
	if m.Visible {
		t.Fatal("hide failed")
	}
	p := SuspensePlaceholder{}
	p.Show("x")
	p.Resolve()
	if !p.Resolved || p.Visible {
		t.Fatal("resolve failed")
	}
}

func TestTransitionAndClip(t *testing.T) {
	tr := Transition{Property: "opacity"}
	tr.Start()
	if !tr.Active {
		t.Fatal("start failed")
	}
	tr.Finish()
	if tr.Active {
		t.Fatal("finish failed")
	}
	c := ClipRect{X: 0, Y: 0, W: 10, H: 10}
	if !c.Contains(5, 5) || c.Contains(15, 5) {
		t.Fatal("clip broken")
	}
}

func TestFocusRestoreAndPromiseError(t *testing.T) {
	f := FocusRestorer{}
	f.Save("input", 3, [2]int{1, 4})
	if !f.Restore("input") || f.Restore("other") {
		t.Fatal("restore broken")
	}
	e := PromiseError{Source: "page", Err: "boom"}
	if e.Error() != "page: boom" {
		t.Fatal("error broken")
	}
}

func TestObserverWatchdogCounter(t *testing.T) {
	seen := ""
	o := Observer{Callback: func(k, id string) { seen = k + id }}
	o.Notify("resize", "box")
	if seen != "resizebox" {
		t.Fatal("observer broken")
	}
	w := Watchdog{}
	if w.Tick(2) {
		t.Fatal("unexpected over first tick")
	}
	w.Frames = 10
	if !w.Tick(2) {
		t.Fatal("expected over")
	}
	d := DevtoolsCounter{}
	d.Add(2, 1)
	if d.Patched != 2 || d.Repainted != 1 {
		t.Fatal("counter broken")
	}
}
