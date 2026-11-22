package window

import "testing"

func TestFitInto(t *testing.T) {
	// The 1920 by 1080 monitor left of the primary one, taskbar excluded.
	work := fitRect{x: -1920, y: 320, w: 1920, h: 1032}

	cases := []struct {
		name string
		win  fitRect
		want fitRect
		move bool
	}{
		{"already inside", fitRect{-1700, 400, 480, 676}, fitRect{-1700, 400, 480, 676}, false},
		{"exactly the work area", work, work, false},
		{"hangs off the bottom", fitRect{-1700, 700, 720, 1014}, fitRect{-1700, 338, 720, 1014}, true},
		{"taller than the monitor", fitRect{-1800, 340, 1470, 1134}, fitRect{-1800, 320, 1470, 1032}, true},
		{"header above the work area", fitRect{-1700, 100, 480, 676}, fitRect{-1700, 320, 480, 676}, true},
		{"off the left edge", fitRect{-2200, 400, 480, 676}, fitRect{-1920, 400, 480, 676}, true},
		{"off the right edge", fitRect{-200, 400, 480, 676}, fitRect{-480, 400, 480, 676}, true},
		{"wider and taller", fitRect{-2100, 200, 3000, 2000}, work, true},
	}

	for _, tc := range cases {
		got, moved := fitInto(tc.win, work)
		if got != tc.want || moved != tc.move {
			t.Errorf("%s: fitInto(%+v) = %+v, %t; want %+v, %t", tc.name, tc.win, got, moved, tc.want, tc.move)
		}
	}
}

func TestFitIntoIgnoresEmptyInput(t *testing.T) {
	win := fitRect{10, 10, 100, 100}

	if got, moved := fitInto(win, fitRect{}); got != win || moved {
		t.Fatalf("empty work area changed the window: %+v, %t", got, moved)
	}

	if got, moved := fitInto(fitRect{}, fitRect{0, 0, 800, 600}); got != (fitRect{}) || moved {
		t.Fatalf("empty window changed: %+v, %t", got, moved)
	}
}

func TestFitNeverEnlarges(t *testing.T) {
	got, _ := fitInto(fitRect{5000, 5000, 100, 100}, fitRect{0, 0, 1920, 1080})
	if got.w != 100 || got.h != 100 {
		t.Fatalf("fit changed the size of a window that fits: %+v", got)
	}
}

// fitScript drives windowFit.step with a scripted platform.
type fitScript struct {
	probe   fitProbe
	applied []fitRect
}

func (s *fitScript) hooks() (func() (fitProbe, bool), func(fitRect) bool) {
	return func() (fitProbe, bool) { return s.probe, true },
		func(r fitRect) bool { s.applied = append(s.applied, r); return true }
}

func (f *windowFit) run(n int, dragging bool, s *fitScript) {
	probe, apply := s.hooks()
	for range n {
		f.step(dragging, probe, apply)
	}
}

func TestFitRunsOnceAtStartupAfterItSettles(t *testing.T) {
	s := &fitScript{probe: fitProbe{
		monitor: 1,
		window:  fitRect{-1700, 700, 720, 1014},
		work:    fitRect{-1920, 320, 1920, 1032},
	}}

	var f windowFit

	f.run(fitSettleUpdates-1, false, s)

	if len(s.applied) != 0 {
		t.Fatalf("fit ran before the window settled: %+v", s.applied)
	}

	f.run(1, false, s)

	if len(s.applied) != 1 || s.applied[0] != (fitRect{-1700, 338, 720, 1014}) {
		t.Fatalf("applied = %+v, want one fit to {-1700 338 720 1014}", s.applied)
	}

	f.run(60, false, s)

	if len(s.applied) != 1 {
		t.Fatalf("fit ran again with no monitor change: %+v", s.applied)
	}
}

func TestFitWaitsForTheDragAndForMaximize(t *testing.T) {
	s := &fitScript{probe: fitProbe{
		monitor: 1,
		window:  fitRect{-1700, 700, 720, 1014},
		work:    fitRect{-1920, 320, 1920, 1032},
	}}

	var f windowFit

	f.run(30, true, s)

	if len(s.applied) != 0 {
		t.Fatalf("fit ran during a drag: %+v", s.applied)
	}

	s.probe.skip = true
	f.run(30, false, s)

	if len(s.applied) != 0 {
		t.Fatalf("fit ran on a maximized window: %+v", s.applied)
	}

	s.probe.skip = false
	f.run(fitSettleUpdates, false, s)

	if len(s.applied) != 1 {
		t.Fatalf("fit did not run once the drag and maximize ended: %+v", s.applied)
	}
}

func TestFitRunsAgainOnANewMonitor(t *testing.T) {
	s := &fitScript{probe: fitProbe{
		monitor: 1,
		window:  fitRect{100, 100, 480, 676},
		work:    fitRect{0, 0, 1707, 1019},
	}}

	var f windowFit

	f.run(20, false, s)

	if len(s.applied) != 0 {
		t.Fatalf("a window that fits was moved: %+v", s.applied)
	}

	s.probe = fitProbe{
		monitor: 2,
		window:  fitRect{-1000, 900, 1470, 1134},
		work:    fitRect{-1920, 320, 1920, 1032},
	}

	f.run(fitSettleUpdates, false, s)

	if len(s.applied) != 1 || s.applied[0] != (fitRect{-1470, 320, 1470, 1032}) {
		t.Fatalf("applied = %+v, want one fit on the new monitor", s.applied)
	}
}

func TestFitDoesNothingWithoutAPlatformProbe(t *testing.T) {
	var f windowFit

	called := false

	for range 20 {
		f.step(false, func() (fitProbe, bool) { return fitProbe{}, false }, func(fitRect) bool { called = true; return true })
	}

	if called || f.started {
		t.Fatal("fit acted on a platform with no probe")
	}
}
