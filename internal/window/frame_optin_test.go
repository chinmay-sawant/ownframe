package window

import "testing"

// optInScreen is a ticking screen whose frame-dirty flag clears on read, the
// way *page.Page does.
type optInScreen struct {
	*fakeScreen
	ticking bool
	dirty   bool
	reads   int
}

func (o *optInScreen) Ticking() bool { return o.ticking }

func (o *optInScreen) FrameDirty() bool {
	o.reads++
	on := o.dirty
	o.dirty = false

	return on
}

// TestFullReplayCapturesOptInPerTick pins the window-side fix: the opt-in is
// read once per Update, so a Draw that Ebiten issues without an intervening
// Update keeps the buffer path instead of replaying the whole list.
func TestFullReplayCapturesOptInPerTick(t *testing.T) {
	app := &optInScreen{fakeScreen: &fakeScreen{width: 100, height: 100}, ticking: true, dirty: true}
	s := &shell{app: app}

	s.captureFrameOptIn()
	if s.fullReplay() {
		t.Fatal("an opted-in tick must keep the buffer path")
	}

	reads := app.reads
	if s.fullReplay() {
		t.Fatal("a draw without an Update must not fall back to the full replay")
	}
	if app.reads != reads {
		t.Fatalf("FrameDirty read %d times during Draw, want %d", app.reads, reads)
	}

	app.dirty = false
	s.captureFrameOptIn()
	if !s.fullReplay() {
		t.Fatal("a tick without the opt-in must replay in full")
	}
}

func TestFullReplayIgnoresNonTickingScreens(t *testing.T) {
	app := &optInScreen{fakeScreen: &fakeScreen{width: 10, height: 10}}
	s := &shell{app: app}
	s.captureFrameOptIn()

	if s.fullReplay() {
		t.Fatal("a screen without a tick keeps the ordinary partial path")
	}
}
