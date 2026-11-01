package window

import "testing"

func TestWSLgLeftMonitorDragAndMaximize(t *testing.T) {
	raw := []monitorRect{
		{name: "XWAYLAND1", x: 0, y: 320, w: 1920, h: 1080},
		{name: "XWAYLAND0", x: 1920, y: 0, w: 2560, h: 1600},
	}
	layout := wslgMonitorLayout(raw)
	minX, minY, maxX, maxY, ok := virtualUnion(layout)
	if !ok {
		t.Fatal("missing monitor bounds")
	}
	// The captured drag was wrongly pinned at -960 by the unshifted union.
	if x, y := clampDragTargetInUnion(-1200, 600, 1120, 760, minX, minY, maxX, maxY); x != -1200 || y != 600 {
		t.Fatalf("left-panel drag clamped to %d,%d", x, y)
	}
	m, ok := monitorForWindow(-1200, 600, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND1" || m.x != -1920 || m.y != 320 {
		t.Fatalf("left maximize target=%+v ok=%v", m, ok)
	}
	m, ok = monitorForWindow(720, 420, 1120, 760, layout)
	if !ok || m.name != "XWAYLAND0" || m.x != 0 || m.y != 0 {
		t.Fatalf("right maximize target=%+v ok=%v", m, ok)
	}
	if raw[1].x != 1920 {
		t.Fatal("source layout mutated")
	}
}
