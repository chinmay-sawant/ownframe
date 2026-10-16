package eyes

import (
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestHUDHidesAndReturnsAfterInput(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	if a.pm.devBG == nil || a.pm.modeBG == nil {
		t.Fatal("control backgrounds were not bound")
	}
	c.adv(hudIdle)
	tk(t, a, c, time.Second/60)
	if a.pm.dev.Opacity() > hudMin || a.pm.mode.Opacity() > hudMin || a.pm.hint.Opacity() > hudMin ||
		a.pm.devBG.Opacity() > hudMin || a.pm.modeBG.Opacity() > hudMin {
		t.Fatal("idle HUD did not fade out")
	}
	for _, op := range a.pm.hintRuns {
		if op.Text != "" {
			t.Fatalf("idle hint run is still painted: %q", op.Text)
		}
	}
	if err := a.onTap(t.Context(), ownframe.Box{ID: "eye-left"}); err != nil {
		t.Fatal(err)
	}
	tk(t, a, c, time.Second/60)
	if a.pm.dev.Opacity() < .99 || a.pm.mode.Opacity() < .99 || a.pm.hint.Opacity() < .99 ||
		a.pm.devBG.Opacity() < .99 || a.pm.modeBG.Opacity() < .99 {
		t.Fatal("user input did not restore the HUD")
	}
	if a.pm.hint.Text == "" {
		t.Fatal("user input did not restore the navigation hint")
	}
	for _, op := range a.pm.hintRuns[1:] {
		if op.Text != "" {
			t.Fatalf("extra hint run was not cleared: %q", op.Text)
		}
	}
}
