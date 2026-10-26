package window

import "testing"

// TestDevDockOnTheRight checks the dock covers the right edge at full
// height and clamps a width wider than the window.
func TestDevDockOnTheRight(t *testing.T) {
	t.Parallel()

	if got := devDockRect(800, 600, 340); got != (devRect{X: 460, Y: 0, W: 340, H: 600}) {
		t.Fatalf("dock = %+v, want the right 340 columns", got)
	}

	if got := devDockRect(300, 200, 340); got != (devRect{X: 0, Y: 0, W: 300, H: 200}) {
		t.Fatalf("clamped dock = %+v, want the whole screen", got)
	}
}

// TestDevBadgeClearsTheDock checks the fallback badge shifts left of the
// dock instead of hiding under it.
func TestDevBadgeClearsTheDock(t *testing.T) {
	t.Parallel()

	dock := devDockRect(800, 600, devDockWidth)
	x, y, w, h := badgeRectInset(800, badgeLabel, dock.W)
	badge := devRect{X: x, Y: y, W: w, H: h}

	if badge.X+badge.W > dock.X {
		t.Fatalf("badge %+v overlaps dock %+v", badge, dock)
	}
}

// TestDevDockResponsiveWidth checks the default dock takes its full 340 px
// only on a roomy screen, gives a narrow window a share instead, and leaves
// the page at least devDockPageMin wide wherever the screen has room.
func TestDevDockResponsiveWidth(t *testing.T) {
	t.Parallel()

	cases := []struct {
		screenW int
		dock    float64
	}{
		{1500, devDockWidth},
		{1180, devDockWidth},
		{1100, 308},
		{1000, 280},
		{800, 224},
		{600, devDockMin},
		{420, devDockMin},
	}

	for _, c := range cases {
		s := newDevShell(newDevScreen())
		s.screenW, s.screenH = c.screenW, 900
		s.dev.on = true

		got := s.devDockWidth()
		if got != c.dock {
			t.Fatalf("%d px screen: dock = %v, want %v", c.screenW, got, c.dock)
		}

		if got < devDockMin && c.screenW >= int(devDockMin) {
			t.Fatalf("%d px screen: dock = %v, under the %v minimum", c.screenW, got, devDockMin)
		}

		if got > float64(c.screenW) {
			t.Fatalf("%d px screen: dock = %v, wider than the screen", c.screenW, got)
		}

		pageW, pageH := s.viewSize()
		if pageW != c.screenW-int(got) || pageH != 900 {
			t.Fatalf("%d px screen: page = %dx%d, want %dx900",
				c.screenW, pageW, pageH, c.screenW-int(got))
		}

		if c.screenW >= int(devDockMin+devDockPageMin) && pageW < int(devDockPageMin) {
			t.Fatalf("%d px screen: page = %d, want at least %v",
				c.screenW, pageW, devDockPageMin)
		}

		panel := devDockRect(s.screenW, s.screenH, got)
		if panel.X < 0 || panel.W != got || panel.X+panel.W != float64(c.screenW) {
			t.Fatalf("%d px screen: panel = %+v", c.screenW, panel)
		}
	}
}

// TestDevDockWidthClamps checks the stored drag width never goes under the
// minimum or squeezes the page under 320 px on a screen with room for it.
func TestDevDockWidthClamps(t *testing.T) {
	t.Parallel()

	d := newDevScreen()
	s := newDevShell(d)

	if got := s.devDockWidth(); got != devDockWidth {
		t.Fatalf("default width = %v, want %v", got, devDockWidth)
	}

	s.dev.dockW = 100
	if got := s.devDockWidth(); got != devDockMin {
		t.Fatalf("narrow width = %v, want %v", got, devDockMin)
	}

	s.dev.dockW = 9000
	if got := s.devDockWidth(); got != float64(s.screenW)-devDockPageMin {
		t.Fatalf("wide width = %v, want the page at %v", got, devDockPageMin)
	}

	s.screenW = 800
	s.dev.dockW = 9000
	if got := s.devDockWidth(); got != float64(s.screenW)-devDockPageMin {
		t.Fatalf("small screen drag = %v, want the page at %v", got, devDockPageMin)
	}

	s.dev.dockW = 0
	s.screenW = 200
	if got := s.devDockWidth(); got != float64(s.screenW) {
		t.Fatalf("screen under the minimum = %v, want it filled", got)
	}
}
