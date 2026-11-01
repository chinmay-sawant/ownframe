package window

import (
	"testing"
	"time"
)

func TestRestoreRejectsCascadedWorkarea(t *testing.T) {
	want := workarea{x: -1920, y: 320, w: 1920, h: 1032}
	for _, offset := range []int{1, 40, 48, 50} {
		actual := want
		actual.y += offset
		if restoreRectMatches(actual, want) {
			t.Fatalf("accepted %d-pixel top gap", offset)
		}
	}
	if !restoreRectMatches(want, want) {
		t.Fatal("correct saved rectangle rejected")
	}
}

func TestPartialWindowChoosesLargestVisibleArea(t *testing.T) {
	layout := []monitorRect{{name: "left", x: -1920, y: 320, w: 1920, h: 1080}, {name: "right", w: 2560, h: 1600}}
	for _, percent := range []int{10, 20, 40, 50, 60, 80, 90} {
		// A 1000px-wide window with this percentage on the right panel.
		m, ok := monitorForWindow(percent*10-1000, 400, 1000, 600, layout)
		want := "left"
		if percent >= 50 {
			want = "right"
		}
		if !ok || m.name != want {
			t.Fatalf("%d%% on right: got %q, want %q", percent, m.name, want)
		}
	}
	// The center is in the left monitor, but most visible pixels are right.
	m, ok := monitorForWindow(-600, 0, 1000, 600, layout)
	if !ok || m.name != "right" {
		t.Fatalf("vertical offset selected %+v", m)
	}
}

func TestWorkareaProbeWaitsForFinalPosition(t *testing.T) {
	var p workareaProbe
	now := time.Now()
	partial := workarea{x: -1920, y: 360, w: 1920, h: 1032}
	final := partial
	final.y = 320
	if p.observe(partial, now) {
		t.Fatal("accepted first configure")
	}
	if p.observe(final, now.Add(50*time.Millisecond)) {
		t.Fatal("accepted moving workarea")
	}
	if p.observe(final, now.Add(100*time.Millisecond)) {
		t.Fatal("accepted before settlement")
	}
	if !p.observe(final, now.Add(200*time.Millisecond)) {
		t.Fatal("final workarea did not settle")
	}
}

func TestHostWorkareaUsesSelectedMonitor(t *testing.T) {
	screens := []hostWorkarea{
		{Bounds: desktopRect{X: 0, Y: 0, W: 2560, H: 1600}, Work: desktopRect{X: 0, Y: 0, W: 2560, H: 1528}},
		{Bounds: desktopRect{X: -1920, Y: 320, W: 1920, H: 1080}, Work: desktopRect{X: -1920, Y: 320, W: 1920, H: 1032}},
	}
	area, ok := matchHostWorkarea(monitorRect{x: -1920, y: 320, w: 1920, h: 1080}, screens)
	if !ok || area != (workarea{x: -1920, y: 320, w: 1920, h: 1032}) {
		t.Fatalf("left workarea=%+v ok=%v", area, ok)
	}
}
