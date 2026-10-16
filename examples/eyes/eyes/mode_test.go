package eyes

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestPoseCatalog(t *testing.T) {
	if len(poses) != 30 {
		t.Fatalf("poses=%d", len(poses))
	}
	seen := make(map[string]bool, len(poses))
	for i := range poses {
		if poses[i].name == "" || seen[poses[i].name] || strings.TrimSpace(poses[i].name) != poses[i].name {
			t.Fatalf("invalid or duplicate pose name %q", poses[i].name)
		}
		seen[poses[i].name] = true
		l, r := targetOpen(i)
		if l <= 0 || r <= 0 || l > 1.3 || r > 1.3 {
			t.Fatalf("%s has invalid eye opening %v / %v", poses[i].name, l, r)
		}
		l, r = poses[i].left.height, poses[i].right.height
		if l <= 0 || r <= 0 || l > 1.3 || r > 1.3 {
			t.Fatalf("%s has invalid eye height %v / %v", poses[i].name, l, r)
		}
	}
	for i := 12; i < len(poses); i++ {
		if poses[i].motion == motionNone {
			t.Fatalf("%s has no distinct animation", poses[i].name)
		}
	}
}

func TestModeToggleButton(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	if !a.st.automatic || a.pm.mode == nil || a.pm.mode.Text != "AUTO" {
		t.Fatal("automatic mode is not the visible default")
	}
	button := ownframe.Box{ID: "mode-toggle"}
	if err := a.onTap(t.Context(), button); err != nil {
		t.Fatal(err)
	}
	tk(t, a, c, time.Second/60)
	if a.st.automatic || a.pm.mode.Text != "STATIC" {
		t.Fatal("mode button did not select static mode")
	}
	if err := a.onTap(t.Context(), button); err != nil {
		t.Fatal(err)
	}
	tk(t, a, c, time.Second/60)
	if !a.st.automatic || a.pm.mode.Text != "AUTO" {
		t.Fatal("mode button did not return to automatic mode")
	}
}
