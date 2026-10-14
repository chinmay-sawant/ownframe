package window

import (
	"testing"
	"time"
)

// wantScreen is a screen that reports a next-frame deadline, the way
// *page.Page does. It clears the deadline when it is read.
type wantScreen struct {
	*fakeScreen
	at    time.Time
	reads int
}

func (w *wantScreen) NextFrameAt() time.Time {
	w.reads++
	at := w.at
	w.at = time.Time{}

	return at
}

// TestPaceCaptureWithoutWanter pins that a custom screen without the
// interface keeps the keep-alive fallback: no deadline is captured.
func TestPaceCaptureWithoutWanter(t *testing.T) {
	g := &paceGame{shell: &shell{app: &fakeScreen{width: 10, height: 10}}}
	g.captureFrameWant()

	if !g.want.IsZero() {
		t.Fatal("a screen without host.FrameWanter must not request a frame")
	}
}

// TestPaceCaptureReadsWanterOnce pins that the deadline is read once per tick
// and clears on read.
func TestPaceCaptureReadsWanterOnce(t *testing.T) {
	want := time.Unix(10000, 0)
	app := &wantScreen{fakeScreen: &fakeScreen{width: 10, height: 10}, at: want}
	g := &paceGame{shell: &shell{app: app}}

	g.captureFrameWant()
	if !g.want.Equal(want) {
		t.Fatalf("want = %v, want %v", g.want, want)
	}
	if app.reads != 1 {
		t.Fatalf("NextFrameAt read %d times, want 1", app.reads)
	}

	g.captureFrameWant()
	if !g.want.IsZero() {
		t.Fatal("a consumed deadline must not repeat")
	}
}
