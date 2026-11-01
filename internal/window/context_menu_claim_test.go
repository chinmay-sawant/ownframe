package window

import (
	"context"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// TestMenuClaimActsOnRowPress gives the open menu first refusal on a left
// press. A custom-chrome draggable strip runs before the menu in pointer(),
// so without this a press on a menu row inside the strip starts a window
// drag: the row never acts and the release lands a stray page click.
func TestMenuClaimActsOnRowPress(t *testing.T) {
	t.Parallel()

	app := &menuForScreen{
		menuScreen: &menuScreen{
			fakeScreen: &fakeScreen{width: 300, height: 300},
			items: []host.MenuItem{
				{ID: "do-thing", Label: "Do thing", Enabled: true},
			},
		},
		rows: map[string][]host.MenuItem{},
	}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)
	if !s.menu.open {
		t.Fatal("menu did not open")
	}

	mx, my, mw, _ := s.menuRect()
	cx, cy := int(mx+mw/2), int(my+menuRowH/2)
	if got := s.menuIndex(cx, cy); got != 0 {
		t.Fatalf("row centre index = %d, want 0", got)
	}

	handled, err := s.menuClaim(cx, cy, true, true)
	if err != nil || !handled {
		t.Fatalf("claim = %v, %v", handled, err)
	}

	if app.acted != "do-thing" {
		t.Fatalf("acted = %q, want do-thing", app.acted)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after a row press")
	}

	if !s.mouseDown {
		t.Fatal("menu claim left mouseDown false, so the press would fire twice")
	}
}

// TestMenuClaimOutsideDismisses checks that a left press outside the open
// menu still dismisses it through the same first-refusal path, instead of
// reaching the window drag or the page.
func TestMenuClaimOutsideDismisses(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "copy", Label: "Copy", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}
	s.openMenu(10, 20)

	handled, err := s.menuClaim(250, 250, true, true)
	if err != nil || !handled {
		t.Fatalf("claim = %v, %v", handled, err)
	}

	if s.menu.open {
		t.Fatal("menu stayed open after an outside press")
	}
}

// TestMenuClaimIgnoresWithoutPress checks the claim stays quiet when the
// menu is closed or the press is not a fresh left edge, so normal pointer
// flow is untouched.
func TestMenuClaimIgnoresWithoutPress(t *testing.T) {
	t.Parallel()

	app := &menuScreen{fakeScreen: &fakeScreen{}, items: []host.MenuItem{
		{ID: "copy", Label: "Copy", Enabled: true},
	}}
	s := &shell{app: app, ctx: context.Background(), screenW: 300, screenH: 300}

	if handled, err := s.menuClaim(15, 25, true, true); err != nil || handled {
		t.Fatalf("closed menu claim = %v, %v", handled, err)
	}

	s.openMenu(10, 20)

	if handled, err := s.menuClaim(15, 25, false, true); err != nil || handled {
		t.Fatalf("held level claim = %v, %v", handled, err)
	}

	if !s.menu.open {
		t.Fatal("claim disturbed the menu without a fresh press")
	}
}
