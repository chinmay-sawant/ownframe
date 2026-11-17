package main

import (
	"bytes"
	"context"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// snapshotRig is a page whose frame loop is driven by the test, with the
// snapshot session installed as the window would install it.
type snapshotRig struct {
	page  *ownframe.Page
	clock *manualLiveClock
	hub   *snapshotHub
	cx    float64
	cy    float64
}

func newSnapshotRig(t *testing.T, tab string, extra ...liveSession) *snapshotRig {
	t.Helper()

	app, err := NewApp(catalogPath())
	if err != nil {
		t.Fatal(err)
	}

	page, err := app.NewPage(context.Background(), tab, 1000, 700)
	if err != nil {
		t.Fatal(err)
	}

	rig := &snapshotRig{page: page, clock: &manualLiveClock{t: time.Unix(1_700_000_000, 0)}, hub: newSnapshotHub()}
	runner := &liveRunner{
		page: page, sessions: append(extra, &snapshotSession{hub: rig.hub}), clock: rig.clock,
		src: liveSource{
			cursor: func() (float64, float64) { return rig.cx, rig.cy },
			click:  func() bool { return false },
			down:   func() bool { return false },
			keys:   func() []string { return nil },
			text:   func() string { return "" },
		},
	}
	runner.attach()

	return rig
}

// step runs one frame of the loop.
func (r *snapshotRig) step(t *testing.T) {
	t.Helper()

	r.clock.advance(50 * time.Millisecond)

	if err := r.page.Testing().Tick(context.Background()); err != nil {
		t.Fatal(err)
	}

}

// snap asks for a picture and runs frames until the loop answers it.
func (r *snapshotRig) snap(t *testing.T) []byte {
	t.Helper()

	type result struct {
		png []byte
		err error
	}

	done := make(chan result, 1)

	go func() {
		png, err := r.hub.Request(context.Background())
		done <- result{png, err}
	}()

	for range 200 {
		r.step(t)

		select {
		case res := <-done:
			if res.err != nil {
				t.Fatal(res.err)
			}

			return res.png
		case <-time.After(5 * time.Millisecond):
		}
	}

	t.Fatal("the frame loop never answered the snapshot request")

	return nil
}

// TestSnapshotFollowsTheWindow: the picture is the window's current page. After
// a click on another tab button the next snapshot is a different picture of
// that tab.
func TestSnapshotFollowsTheWindow(t *testing.T) {
	rig := newSnapshotRig(t, "motion")

	first := rig.snap(t)
	if _, err := png.Decode(bytes.NewReader(first)); err != nil {
		t.Fatalf("first snapshot is not a PNG: %v", err)
	}

	if again := rig.snap(t); !bytes.Equal(first, again) {
		t.Error("two snapshots of an unchanged window differ")
	}

	var x, y float64

	for _, b := range rig.page.Probe().Display().Boxes {
		if b.ID == "tab-btn-svg" {
			x, y = b.X+b.W/2, b.Y+b.H/2
		}
	}

	if x == 0 && y == 0 {
		t.Fatal("no tab-btn-svg box on the page")
	}

	// The window hands the click to the page; the test does the same.
	if err := rig.page.Click(context.Background(), x, y); err != nil {
		t.Fatal(err)
	}

	rig.step(t)

	second := rig.snap(t)
	if bytes.Equal(first, second) {
		t.Error("the snapshot did not change after clicking another tab")
	}
}

// TestSnapshotHandler: GET /frame.png returns the picture, and the viewer page
// is served at /.
func TestSnapshotHandler(t *testing.T) {
	rig := newSnapshotRig(t, "shapes")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Drive the loop in the background, as the window does.
	go func() {
		for ctx.Err() == nil {
			time.Sleep(5 * time.Millisecond)
			_ = rig.page.Testing().Tick(ctx)
			rig.clock.advance(50 * time.Millisecond)
		}
	}()

	hub := rig.hub

	rec := httptest.NewRecorder()
	handlerPNG(hub).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/frame.png", nil))

	if rec.Code != http.StatusOK || rec.Header().Get("Content-Type") != "image/png" {
		t.Fatalf("frame.png: status %d type %q", rec.Code, rec.Header().Get("Content-Type"))
	}

	if _, err := png.Decode(bytes.NewReader(rec.Body.Bytes())); err != nil {
		t.Errorf("frame.png body is not a PNG: %v", err)
	}
}

// TestSnapshotShowsLivePatches: a live session moves the motion box with the
// clock, and the snapshot taken later shows the box somewhere else.
func TestSnapshotShowsLivePatches(t *testing.T) {
	rig := newSnapshotRig(t, "motion", newMotionSession())

	first := rig.snap(t)

	for range 40 { // two seconds of frames
		rig.step(t)
	}

	if later := rig.snap(t); bytes.Equal(first, later) {
		t.Error("the snapshot did not show the motion box moving")
	}
}

// TestSnapshotDoesNotPoll: the session never asks the window for a frame, and
// the viewer page has no timer that reloads the picture.
func TestSnapshotDoesNotPoll(t *testing.T) {
	rig := newSnapshotRig(t, "shapes")

	session := &snapshotSession{hub: rig.hub}

	next, err := session.Frame(context.Background(), rig.page, liveInput{}, rig.clock.Now())
	if err != nil || !next.IsZero() {
		t.Errorf("an idle session asked for a frame: next=%v err=%v", next, err)
	}

	for _, banned := range []string{"setInterval", "setTimeout", "location.reload", `http-equiv="refresh"`} {
		if strings.Contains(snapshotViewer, banned) {
			t.Errorf("the viewer page reloads on its own: it contains %q", banned)
		}
	}
}
