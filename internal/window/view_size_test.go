package window

import "testing"

// TestViewSizeShrinksForDock checks the page viewport follows the dock.
func TestViewSizeShrinksForDock(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)
	s.screenW, s.screenH = 800, 600
	s.dev.on = false

	if w, h := s.viewSize(); w != 800 || h != 600 {
		t.Fatalf("closed dock view = %dx%d, want 800x600", w, h)
	}

	s.dev.on = true
	// An 800 px window gives the panel 224 px, not the full 340.
	if w, h := s.viewSize(); w != 576 || h != 600 {
		t.Fatalf("open dock view = %dx%d, want 576x600", w, h)
	}
}

// TestResizeShrinksPageForDock checks opening and closing the dock commits
// the smaller and then the full page size.
func TestResizeShrinksPageForDock(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	d.width, d.height = 800, 600
	s := newDevShell(d)
	s.pendingW, s.pendingH = 800, 600
	s.screenW, s.screenH = 800, 600
	s.dev.on = true

	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if w, h := d.Size(); w != 576 || h != 600 {
		t.Fatalf("open dock page = %dx%d, want 576x600", w, h)
	}

	s.dev.on = false
	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if w, h := d.Size(); w != 800 || h != 600 {
		t.Fatalf("closed dock page = %dx%d, want 800x600", w, h)
	}
}

// TestMinClampedPageScrolls checks a page that cannot shrink to the viewport
// is not scaled, so a pointer stays on its pixel.
func TestMinClampedPageScrolls(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	d.width, d.height = 1440, 960
	d.minW = 1120
	s := newDevShell(d)
	s.pendingW, s.pendingH = 1440, 960
	s.screenW, s.screenH = 1440, 960
	s.dev.on = true

	if err := s.resize(); err != nil {
		t.Fatal(err)
	}

	if s.stretched() {
		t.Fatal("a min-clamped page must scroll, not stretch")
	}

	frameW, frameH := s.frameSize()
	if x, _ := s.contentAt(100, 50, frameW, frameH); x != 100 {
		t.Fatalf("pointer x = %v, want 100", x)
	}
}
