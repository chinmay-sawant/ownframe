package window

import "testing"

func TestPositionVisibleInBound(t *testing.T) {
	t.Parallel()

	// 1920 + 2560 side by side, heights summed generously.
	const bw, bh = 4480, 2680

	if !positionVisibleInBound(89, 537, 1120, 760, bw, bh) {
		t.Fatal("sane ui-builder origin reported off-screen")
	}
	if !positionVisibleInBound(2000, 400, 1120, 760, bw, bh) {
		t.Fatal("visible right-monitor origin reported off-screen")
	}
	if positionVisibleInBound(9000, 9000, 1120, 760, bw, bh) {
		t.Fatal("far right origin reported visible")
	}
	if positionVisibleInBound(-9000, -9000, 1120, 760, bw, bh) {
		t.Fatal("far left origin reported visible")
	}
	if !positionVisibleInBound(9000, 9000, 1120, 760, 0, 0) {
		t.Fatal("unknown bound should keep the old behavior")
	}
}

func TestExpandedResolutionMustFillMonitor(t *testing.T) {
	t.Parallel()

	// A 1920 workarea is not an expanded resolution on the 2560 panel.
	if areaMatchesMonitorSize(2560, 1600, workarea{w: 1920, h: 1040}) {
		t.Fatal("1920 workarea matched the 2560 panel")
	}
	// A 2560 workarea is not an expanded resolution on the 1920 panel.
	if areaMatchesMonitorSize(1920, 1080, workarea{w: 2560, h: 1528}) {
		t.Fatal("2560 workarea matched the 1920 panel")
	}
	// The right workarea still matches its own panel, minus the taskbar gap.
	if !areaMatchesMonitorSize(2560, 1600, workarea{w: 2560, h: 1528}) {
		t.Fatal("2560 workarea did not match the 2560 panel")
	}
	if !areaMatchesMonitorSize(1920, 1080, workarea{w: 1920, h: 1040}) {
		t.Fatal("1920 workarea did not match the 1920 panel")
	}
}

func TestLiveMonitorResolutionNeverPanics(t *testing.T) {
	t.Parallel()

	w, h, ok := liveMonitorResolution()
	if !ok {
		return
	}
	if w <= 0 || h <= 0 {
		t.Fatalf("live resolution = %dx%d, want positive", w, h)
	}
}
