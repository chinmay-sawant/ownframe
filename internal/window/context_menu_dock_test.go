package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// TestMenuDockGeometryClicksRow checks that with the devtools dock open the
// menu draws and hit-tests in the same coordinates: a left click at the
// drawn centre of a row acts, and a click outside closes and consumes.
func TestMenuDockGeometryClicksRow(t *testing.T) {
	t.Parallel()

	app := &menuForScreen{
		menuScreen: &menuScreen{
			fakeScreen: &fakeScreen{width: 940, height: 800},
			items: []host.MenuItem{
				{ID: "do-thing", Label: "Do thing", Enabled: true},
				{ID: "do-other", Label: "Do other", Enabled: true},
				{ID: "do-more", Label: "Do more", Enabled: true},
			},
		},
		rows: map[string][]host.MenuItem{},
	}
	const screenW, screenH = 1280, 800
	s := &shell{app: app, ctx: context.Background(), screenW: screenW, screenH: screenH}
	s.dev.on = true
	s.dev.panel = devDockRect(screenW, screenH, s.devDockWidth())

	viewW, _ := s.viewSize()
	if s.dev.panel.X != float64(viewW) {
		t.Fatalf("panel X = %v, view width = %d", s.dev.panel.X, viewW)
	}

	// Right-click near the right and bottom edges of the content strip so
	// the menu must clamp inside the shrunken viewport.
	handled, err := s.devMenuPointer(viewW-10, 780, true, false)
	if err != nil || !handled || !s.menu.open {
		t.Fatalf("right click: handled=%v err=%v menu=%+v", handled, err, s.menu)
	}

	mx, my, mw, mh := s.menuRect()
	if mx+mw > float64(viewW) || my+mh > float64(screenH) {
		t.Fatalf("menu rect (%v, %v, %v, %v) escapes the content strip", mx, my, mw, mh)
	}

	for i := range s.menu.items {
		cx := int(mx + mw/2)
		cy := int(my + float64(i)*menuRowH + menuRowH/2)
		if got := s.menuIndex(cx, cy); got != i {
			t.Fatalf("row %d centre index = %d", i, got)
		}
	}

	cx := int(mx + mw/2)
	cy := int(my + menuRowH + menuRowH/2)
	handled, err = s.devMenuPointer(cx, cy, false, true)
	if err != nil || !handled {
		t.Fatalf("menu click: handled=%v err=%v", handled, err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after a row click")
	}

	if app.acted != "do-other" {
		t.Fatalf("acted = %q, want do-other", app.acted)
	}

	s.openMenu(10, 20)

	handled, err = s.devMenuPointer(viewW-2, screenH-2, false, true)
	if err != nil || !handled {
		t.Fatalf("outside click: handled=%v err=%v", handled, err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after an outside click")
	}

	if app.acted != "do-other" {
		t.Fatalf("outside click acted: %q", app.acted)
	}
}
