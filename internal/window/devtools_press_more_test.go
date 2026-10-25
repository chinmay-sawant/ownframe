package window

import "testing"

// TestDevPanelPressHits checks a press lands on the recorded hit rects and
// that empty panel space does nothing.
func TestDevPanelPressHits(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.panel = devRect{X: 100, W: 200, H: 400}
	s.dev.hits = []devHit{{
		rect: devRect{X: 100, Y: 30, W: 200, H: 16},
		act:  devActTab,
		arg:  int(devTabOps),
	}}

	s.devPanelPress(120, 35)
	if s.dev.tab != devTabOps {
		t.Fatalf("tab = %v, want Ops", s.dev.tab)
	}

	s.devPanelPress(120, 300)
	if s.dev.tab != devTabOps {
		t.Fatal("an empty panel click changed the tab")
	}
}

// TestDevCloseHitClosesDock checks the header close key shuts the overlay
// through the page flag.
func TestDevCloseHitClosesDock(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.dev.panel = devRect{X: 100, W: 400, H: 600}
	cls, _ := devHeaderRight(s.dev.panel)
	s.dev.hits = []devHit{{rect: cls, act: devActClose}}

	s.devPanelPress(int(cls.X+cls.W-4), 10)
	if err := s.devSync(); err != nil {
		t.Fatal(err)
	}
	if s.dev.on || d.on {
		t.Fatal("the close key did not shut the dock")
	}
}

// TestDevResizeFollowsThePointer checks the drag width and its clamps.
func TestDevResizeFollowsThePointer(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	s.devResize(s.screenW - 500)
	if s.dev.dockW != 500 {
		t.Fatalf("dock width = %v, want 500", s.dev.dockW)
	}

	s.devResize(s.screenW - 1)
	if s.dev.dockW != devDockMin {
		t.Fatalf("narrow drag = %v, want %v", s.dev.dockW, devDockMin)
	}
}
