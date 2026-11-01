package window

import "testing"

func TestParseXrandrMonitors(t *testing.T) {
	t.Parallel()

	out := "Monitors: 2\n" +
		" 0: +XWAYLAND0 2560/677x1600/423+1920+0  XWAYLAND0\n" +
		" 1: +XWAYLAND1 1920/527x1080/296+0+320  XWAYLAND1\n"
	rects := parseXrandrMonitors(out)
	if len(rects) != 2 {
		t.Fatalf("parsed %d monitors, want 2", len(rects))
	}
	if rects[0] != (monitorRect{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600}) {
		t.Fatalf("first rect = %+v", rects[0])
	}
	if rects[1] != (monitorRect{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080}) {
		t.Fatalf("second rect = %+v", rects[1])
	}
}

func TestMonitorForWindowPicksHeaderMonitor(t *testing.T) {
	t.Parallel()

	layout := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	// Double-tap header at the center of the 1120x760 frame on the 1920 panel.
	m, ok := monitorForWindow(100, 400, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND1" {
		t.Fatalf("1920 window attributed to %+v, ok=%v", m, ok)
	}
	// Same frame moved to the 2560 panel.
	m, ok = monitorForWindow(2500, 400, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND0" {
		t.Fatalf("2560 window attributed to %+v, ok=%v", m, ok)
	}
	// The stuck strip at x=-960 keeps 160px on the 1920 panel: it overlaps
	// that panel, but less than half its area, so saves must skip it.
	m, ok = monitorForWindow(-960, 425, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND1" {
		t.Fatalf("strip attributed to %+v, ok=%v", m, ok)
	}
	if substantiallyVisibleInLayout(-960, 425, 1120, 760, layout) {
		t.Fatal("off-screen strip reported substantially visible")
	}
	if !substantiallyVisibleInLayout(100, 400, 1120, 760, layout) {
		t.Fatal("sane window reported insubstantial")
	}
}

func TestClampDragTargetInUnion(t *testing.T) {
	t.Parallel()

	minX, minY, maxX, maxY := 0, 0, 4480, 1600
	// Sane position passes through.
	if x, y := clampDragTargetInUnion(2500, 400, 1120, 760, minX, minY, maxX, maxY); x != 2500 || y != 400 {
		t.Fatalf("visible clamped to %d,%d", x, y)
	}
	// Far right pulls back to keep a grabbable strip.
	if x, _ := clampDragTargetInUnion(9000, 400, 1120, 760, minX, minY, maxX, maxY); x != maxX-dragMinVisibleX {
		t.Fatalf("right clamped to %d", x)
	}
	// Below the union (summed-height bound allowed 2648) now clamps.
	if _, y := clampDragTargetInUnion(200, 9000, 1120, 760, minX, minY, maxX, maxY); y != maxY-dragMinVisibleY {
		t.Fatalf("bottom clamped to %d", y)
	}
}

func TestPlacementSubstantiallyVisible(t *testing.T) {
	t.Parallel()

	if !substantiallyVisibleInLayout(100, 400, 1120, 760, []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}) {
		t.Fatal("sane window reported insubstantial")
	}
}

func TestOriginMustBeInsideUnion(t *testing.T) {
	t.Parallel()

	layout := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	// Fresh cascade at y=149 sits in the dead zone above the 1920 panel:
	// area overlaps, but the header is unreachable.
	if originInLayout(128, 149, layout) {
		t.Fatal("dead-zone origin reported visible")
	}
	if !originInLayout(0, 320, layout) {
		t.Fatal("1920 top-left reported invisible")
	}
	if !originInLayout(1920, 0, layout) {
		t.Fatal("2560 top-left reported invisible")
	}
}

func TestTargetMonitorStartsAtTopLeft(t *testing.T) {
	t.Parallel()

	layout := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	// Smallest panel that fits 1120x760 is the 1920 one.
	saved := "XWAYLAND1"
	if target, ok := targetMonitorInLayout(layout, saved, 1120, 760); !ok || target.name != "XWAYLAND1" || target.x != 0 || target.y != 320 {
		t.Fatalf("target = %+v, ok=%v", target, ok)
	}
}
