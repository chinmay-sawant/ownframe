package window

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPlacementFitsUnknownMonitorKeepsOldBehavior(t *testing.T) {
	t.Parallel()

	if !placementFits("no-such-monitor-xyz", 2560, 1528) {
		t.Fatal("unknown monitor should not discard the rect")
	}
}

func TestSavePlacementNowSkipsWorkareaDance(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	prevLearning, prevRestoring := manualMax.learning, manualMax.restoring
	manualMax.learning = true
	manualMax.restoring = false
	t.Cleanup(func() {
		manualMax.learning, manualMax.restoring = prevLearning, prevRestoring
	})

	savePlacementNow("dance-guard")

	path := filepath.Join(config, "ownframe", "placement", "dance-guard.json")
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("dance save wrote %q during learning, err=%v", path, err)
	}

	manualMax.learning = false
	manualMax.restoring = true

	savePlacementNow("dance-guard")

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("dance save wrote %q during restoring, err=%v", path, err)
	}
}

func TestRestoreManualForDragClearsActive(t *testing.T) {
	prevActive := manualMax.active
	prevW, prevH := manualMax.sw, manualMax.sh
	t.Cleanup(func() {
		manualMax.active = prevActive
		manualMax.sw, manualMax.sh = prevW, prevH
	})

	manualMax.active = true
	manualMax.sw, manualMax.sh = 1500, 900

	sw, sh, restored := restoreManualForDrag()
	if !restored {
		t.Fatal("drag restore did not trigger while manually maximized")
	}

	if manualMax.active {
		t.Fatal("drag restore left the manual maximize active")
	}

	if sw <= 0 || sh <= 0 {
		t.Fatalf("drag restore size = %dx%d, want the windowed size", sw, sh)
	}

	if _, _, restored := restoreManualForDrag(); restored {
		t.Fatal("second drag restore triggered without a maximize")
	}
}

func TestApplyWorkareaKeysByMonitorName(t *testing.T) {
	prevActive, prevKnown := manualMax.active, manualMax.known
	prevMon := manualMax.learnedMon
	prevAreas := manualMax.areas
	t.Cleanup(func() {
		manualMax.active, manualMax.known = prevActive, prevKnown
		manualMax.learnedMon = prevMon
		manualMax.areas = prevAreas
	})

	manualMax.areas = nil
	manualMax.sx, manualMax.sy = 100, 100
	manualMax.sw, manualMax.sh = 1500, 900

	applyWorkarea("XWAYLAND1", workarea{x: 0, y: 320, w: 1920, h: 1040})

	if !manualMax.active || !manualMax.known {
		t.Fatal("applyWorkarea did not mark the maximize active")
	}

	if manualMax.learnedMon != "XWAYLAND1" {
		t.Fatalf("learned monitor = %q, want %q", manualMax.learnedMon, "XWAYLAND1")
	}

	got, ok := manualMax.areas["XWAYLAND1"]
	if !ok {
		t.Fatal("workarea not cached under the monitor name")
	}

	if got.w != 1920 || got.h != 1040 || got.x != 0 || got.y != 320 {
		t.Fatalf("cached workarea = %+v, want 0,320 1920x1040", got)
	}
}

func TestWorkareaMatchesMonitorUnknownMonitorKeepsOldBehavior(t *testing.T) {
	t.Parallel()

	if !workareaMatchesMonitor("no-such-monitor-xyz", workarea{w: 2560, h: 1528}) {
		t.Fatal("unknown monitor should not discard the rect")
	}
}

func TestClampDragTargetKeepsStripVisible(t *testing.T) {
	t.Parallel()

	// Far off-screen right/bottom pulls back to the bound edge.
	if x, y := clampDragTarget(9000, 9000, 1500, 900, 4480, 2680); x != 4480-dragMinVisibleX || y != 2680-dragMinVisibleY {
		t.Fatalf("clamped = %d,%d, want %d,%d", x, y, 4480-dragMinVisibleX, 2680-dragMinVisibleY)
	}

	// Far off-screen left/top pulls back so a strip stays reachable.
	if x, y := clampDragTarget(-9000, -9000, 1500, 900, 4480, 2680); x != -(1500-dragMinVisibleX) || y != -(900-dragMinVisibleY) {
		t.Fatalf("clamped = %d,%d, want %d,%d", x, y, -(1500 - dragMinVisibleX), -(900 - dragMinVisibleY))
	}

	// A sane on-screen position passes through untouched.
	if x, y := clampDragTarget(2000, 400, 1500, 900, 4480, 2680); x != 2000 || y != 400 {
		t.Fatalf("clamped = %d,%d, want 2000,400", x, y)
	}
}

func TestClampDragTargetSkipsUnknownBound(t *testing.T) {
	t.Parallel()

	if x, y := clampDragTarget(-6368, 9274, 2560, 1528, 0, 0); x != -6368 || y != 9274 {
		t.Fatalf("unknown bound clamped to %d,%d, want untouched", x, y)
	}
}
