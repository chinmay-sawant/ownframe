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
