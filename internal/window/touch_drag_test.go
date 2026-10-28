package window

import (
	"context"
	"testing"
)

type captureScreen struct {
	fakeScreen
	claim       bool
	moves, ends int
}

func (f *captureScreen) BeginDrag(context.Context, float64, float64) (bool, error) {
	return f.claim, nil
}
func (f *captureScreen) MoveDrag(context.Context, float64, float64) error { f.moves++; return nil }
func (f *captureScreen) EndDrag(context.Context) error                    { f.ends++; return nil }

func TestTouchCaptureMovesBeforeLiftAndCancelsPinch(t *testing.T) {
	app := &captureScreen{fakeScreen: fakeScreen{width: 200, height: 200}, claim: true}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200, pageZoom: 1}
	finger := touchPos{id: 1, x: 40, y: 40}
	frames := [][]touchPos{{finger}, {{id: 1, x: 60, y: 45}}, {{id: 1, x: 60, y: 45}, {id: 2, x: 100, y: 100}}, nil}
	for i, now := range frames {
		u := s.fingers.frame(now, false)
		if _, err := s.routeTouchDrag(u, now, 200, 200); err != nil {
			t.Fatal(err)
		}
		if i == 1 && app.moves != 1 {
			t.Fatal("movement waits for lift")
		}
	}
	if app.ends != 1 || s.gesture.active || s.scrollY != 0 {
		t.Fatalf("capture ends=%d active=%v scroll=%d", app.ends, s.gesture.active, s.scrollY)
	}
}

func TestUnclaimedTouchRemainsScrollable(t *testing.T) {
	app := &captureScreen{fakeScreen: fakeScreen{width: 200, height: 200}, claim: false}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200, pageZoom: 1}
	p := touchPos{id: 1, x: 40, y: 40}
	u := s.fingers.frame([]touchPos{p}, false)
	handled, err := s.routeTouchDrag(u, []touchPos{p}, 200, 200)
	if err != nil || handled || s.gesture.active {
		t.Fatalf("unclaimed %v %v", handled, err)
	}
}

func TestDevToolsForwardsDragUntilRelease(t *testing.T) {
	app := &captureScreen{fakeScreen: fakeScreen{width: 200, height: 200}, claim: true}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	if err := s.devPick(40, 40, false); err != nil {
		t.Fatal(err)
	}
	if !s.dev.forward || !s.gesture.active {
		t.Fatal("content press did not start the captured drag")
	}
	if err := s.devForwardDrag(60, 50, true, false); err != nil {
		t.Fatal(err)
	}
	if err := s.devForwardDrag(60, 50, false, true); err != nil {
		t.Fatal(err)
	}
	if app.moves != 1 || app.ends != 1 || s.dev.forward || s.gesture.active {
		t.Fatalf("moves=%d ends=%d forward=%v active=%v", app.moves, app.ends, s.dev.forward, s.gesture.active)
	}
}
