package window

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatePlacementID(t *testing.T) {
	t.Parallel()

	for _, id := range []string{"code-view", "app_1", "A-0_z"} {
		if err := validatePlacementID(id); err != nil {
			t.Fatalf("validatePlacementID(%q) = %v, want nil", id, err)
		}
	}

	for _, id := range []string{"", "  ", "has space", "a/b", "a\\b", "..", "a..b", "tab\there", "semi;colon", "dot.name", "slash/name"} {
		if err := validatePlacementID(id); err == nil {
			t.Fatalf("validatePlacementID(%q) = nil, want error", id)
		}
	}

	long := make([]byte, placementMaxIDLen+1)
	for i := range long {
		long[i] = 'a'
	}

	if err := validatePlacementID(string(long)); err == nil {
		t.Fatal("overlong placement id accepted")
	}
}

func TestPlacementRoundTrip(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	want := windowPlacement{X: 12, Y: 24, W: 1500, H: 900, Maximized: true, MaxX: 0, MaxY: 0, MaxW: 1920, MaxH: 1032, Monitor: "XWAYLAND1"}

	if err := savePlacement("code-view", want); err != nil {
		t.Fatalf("savePlacement = %v", err)
	}

	got, ok := loadPlacement("code-view")
	if !ok {
		t.Fatal("loadPlacement missed a saved file")
	}

	want.Version = placementVersion
	if got != want {
		t.Fatalf("placement = %+v, want %+v", got, want)
	}

	path := filepath.Join(config, "ownframe", "placement", "code-view.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat placement file = %v", err)
	}

	if info.Mode().Perm() != placementFileMode {
		t.Fatalf("placement mode = %o, want %o", info.Mode().Perm(), placementFileMode)
	}
}

func TestWSLMaximizePlan(t *testing.T) {
	t.Parallel()

	applyNow, retry := wslMaximizePlan(false, true)
	if applyNow || !retry {
		t.Fatalf("unknown monitor: apply=%v retry=%v, want deferred", applyNow, retry)
	}

	applyNow, retry = wslMaximizePlan(true, true)
	if !applyNow || !retry {
		t.Fatalf("fitting monitor: apply=%v retry=%v, want both", applyNow, retry)
	}

	applyNow, retry = wslMaximizePlan(true, false)
	if applyNow || retry {
		t.Fatalf("oversized rect: apply=%v retry=%v, want neither", applyNow, retry)
	}
}

func TestPlacementLoadMissingFallsBack(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	if _, ok := loadPlacement("nothing-saved"); ok {
		t.Fatal("missing placement loaded")
	}

	if _, err := placementPath("bad/id"); err == nil {
		t.Fatal("bad placement id resolved a path")
	}
}

func TestPlacementRejectsBadRect(t *testing.T) {
	config := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", config)

	if err := savePlacement("code-view", windowPlacement{W: 0, H: 900}); err != nil {
		t.Fatalf("savePlacement = %v", err)
	}

	if _, ok := loadPlacement("code-view"); ok {
		t.Fatal("zero-width placement loaded")
	}
}
