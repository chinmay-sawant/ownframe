package window

import (
	"context"
	"testing"
	"time"
)

type touchPressScreen struct {
	fakeScreen
	presses, releases, clicks int
}

func (f *touchPressScreen) Press(context.Context, float64, float64) error { f.presses++; return nil }
func (f *touchPressScreen) Release(context.Context) error                 { f.releases++; return nil }
func (f *touchPressScreen) Click(context.Context, float64, float64) error { f.clicks++; return nil }

func TestTouchTapHasOnePressAndClick(t *testing.T) {
	f := &touchPressScreen{fakeScreen: fakeScreen{width: 200, height: 200}}
	s := &shell{app: f, ctx: context.Background(), screenW: 200, screenH: 200}
	for _, touches := range [][]touchPos{{{id: 1, x: 40, y: 40}}, nil} {
		if err := s.applyTouches(touches, false, 200, 200); err != nil {
			t.Fatal(err)
		}
	}
	if f.presses != 1 || f.clicks != 1 || f.releases != 1 {
		t.Fatalf("events: %+v", f)
	}
}

func TestPinchCancelsActivePressImmediately(t *testing.T) {
	f := &touchPressScreen{fakeScreen: fakeScreen{width: 200, height: 200}}
	s := &shell{app: f, ctx: context.Background(), screenW: 200, screenH: 200}
	for _, touches := range [][]touchPos{{{id: 1, x: 40, y: 40}}, {{id: 1, x: 40, y: 40}, {id: 2, x: 80, y: 40}}, nil} {
		if err := s.applyTouches(touches, false, 200, 200); err != nil {
			t.Fatal(err)
		}
	}
	if f.presses != 1 || f.clicks != 0 || s.hold.due(time.Now().Add(time.Second)) {
		t.Fatalf("events: %+v", f)
	}
}
