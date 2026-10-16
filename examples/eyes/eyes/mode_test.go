package eyes

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestPoseCatalog(t *testing.T) {
	if len(poses) < 10 {
		t.Fatalf("poses=%d", len(poses))
	}
	for i := range poses {
		l, r := targetOpen(i)
		if l <= 0 || r <= 0 || l > 1.3 || r > 1.3 {
			t.Fatalf("%s has invalid eye opening %v / %v", poses[i].name, l, r)
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
