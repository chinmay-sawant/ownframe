package window

import (
	"context"
	"testing"
)

// windowScreen is a fakeScreen that observes scroll offsets.
type windowScreen struct {
	*fakeScreen
	ox, oy int
	on     bool
}

func (f *windowScreen) SetScrollOffset(x, y int) { f.ox, f.oy = x, y }
func (f *windowScreen) ScrollOffset() (int, int) { return f.ox, f.oy }
func (f *windowScreen) Windowing() bool          { return f.on }

func TestSyncScrollWindow(t *testing.T) {
	t.Parallel()

	app := &windowScreen{fakeScreen: &fakeScreen{}, on: true}
	s := &shell{app: app, ctx: context.Background(), scrollX: 10, scrollY: 200}

	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.ox != 10 || app.oy != 200 {
		t.Fatalf("offset = %d, %d, want 10, 200", app.ox, app.oy)
	}
	if app.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", app.redraws)
	}

	app.redraws = 0
	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.redraws != 0 {
		t.Fatalf("redraws = %d on unchanged offset, want 0", app.redraws)
	}

	app.on = false
	s.scrollY = 300
	if err := s.syncScrollWindow(); err != nil {
		t.Fatal(err)
	}
	if app.oy != 300 {
		t.Fatalf("offset y = %d, want 300", app.oy)
	}
	if app.redraws != 0 {
		t.Fatalf("redraws = %d when opted out, want 0", app.redraws)
	}
}

func TestOversizedWithoutADeviceLimit(t *testing.T) {
	t.Parallel()

	if oversized(480, 640) {
		t.Fatal("normal canvas oversized without a device limit")
	}
}
