package eyes

import (
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

func TestTpl(t *testing.T) {
	if !strings.Contains(indexHTML, "<!-- stylesheet -->") {
		t.Fatal("marker lost")
	}
	if !strings.Contains(buildHTML(), "<style>") {
		t.Fatal("css not inlined")
	}
}

func TestTickBinds(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	if a.page.Display() == nil {
		t.Fatal("bitmap fallback")
	}
	if a.pm.left == nil || a.pm.right == nil {
		t.Fatal("eyes not bound")
	}
	if a.pm.left.H <= 0 {
		t.Fatal("left eye hidden")
	}
}

func TestMoodTap(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	m0 := a.st.mood
	tk(t, a, c, time.Second/60)
	if err := a.onTap(t.Context(), a.page.Boxes()[0]); err != nil {
		t.Fatal(err)
	}
	if a.st.mood != m0+1 {
		t.Fatal("tap did not cycle mood")
	}
}

func TestDevToggle(t *testing.T) {
	a, c := newT(t)
	tk(t, a, c, time.Second/60)
	if a.pm.dev == nil {
		t.Fatal("dev switch not bound")
	}
	dev := ownframe.Box{ID: "dev-toggle"}
	if err := a.onTap(t.Context(), dev); err != nil {
		t.Fatal(err)
	}
	if !a.st.dev || !a.page.DevTools() {
		t.Fatal("dev tap did not open the dock")
	}
	if err := a.onTap(t.Context(), dev); err != nil {
		t.Fatal(err)
	}
	if a.st.dev || a.page.DevTools() {
		t.Fatal("dev tap did not close the dock")
	}
}
