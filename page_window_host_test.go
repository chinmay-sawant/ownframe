package ownframe

import "testing"

type fakeWindowHost struct {
	cx, cy          int
	minimized       int
	toggleMaximizes int
	isMaximized     bool
	closed          int
}

func (f *fakeWindowHost) cursorPosition() (int, int) { return f.cx, f.cy }
func (f *fakeWindowHost) minimize()                  { f.minimized++ }
func (f *fakeWindowHost) toggleMaximize() {
	f.toggleMaximizes++
	f.isMaximized = !f.isMaximized
}
func (f *fakeWindowHost) maximized() bool { return f.isMaximized }
func (f *fakeWindowHost) close()          { f.closed++ }

func TestPageWindowHostControlsDetachedAndBound(t *testing.T) {
	// Nil page is a safe no-op.
	var nilPage *Page
	nilPage.Window().Minimize()
	nilPage.Window().ToggleMaximize()
	nilPage.Window().Close()
	if nilPage.Window().Maximized() {
		t.Fatal("nil page reported maximized")
	}
	if x, y := nilPage.Window().CursorPosition(); x != 0 || y != 0 {
		t.Fatalf("nil page cursor = (%d,%d), want (0,0)", x, y)
	}

	// Detached page (before Run/NewGame) does not touch global window state.
	p, err := New(Config{HTML: `<p>hi</p>`, Width: 200, Height: 120})
	if err != nil {
		t.Fatal(err)
	}
	p.Window().Minimize()
	p.Window().ToggleMaximize()
	p.Window().Close()
	if p.Window().Maximized() {
		t.Fatal("detached page reported maximized")
	}
	if x, y := p.Window().CursorPosition(); x != 0 || y != 0 {
		t.Fatalf("detached page cursor = (%d,%d), want (0,0)", x, y)
	}

	// Bound page delegates host actions to its attached host controller.
	fake := &fakeWindowHost{cx: 42, cy: 19}
	p.bindHost(fake)
	defer p.bindHost(nil)

	if x, y := p.Window().CursorPosition(); x != 42 || y != 19 {
		t.Fatalf("bound cursor = (%d,%d), want (42,19)", x, y)
	}
	p.Window().Minimize()
	p.Window().ToggleMaximize()
	if !p.Window().Maximized() {
		t.Fatal("expected maximized after first toggle")
	}
	p.Window().ToggleMaximize()
	if p.Window().Maximized() {
		t.Fatal("expected restored after second toggle")
	}
	p.Window().Close()

	if fake.minimized != 1 || fake.toggleMaximizes != 2 || fake.closed != 1 {
		t.Fatalf("host counts = minimize:%d toggle:%d close:%d, want 1/2/1",
			fake.minimized, fake.toggleMaximizes, fake.closed)
	}
}
