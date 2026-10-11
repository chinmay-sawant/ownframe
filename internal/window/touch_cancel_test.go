package window

import (
	"context"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

type cancelScreen struct {
	touchPressScreen
	cancel bool
}

func (f *cancelScreen) TakeTouchCancel() bool { want := f.cancel; f.cancel = false; return want }

func TestCancelledTouchCannotTapOrRestartWhileNativeIDsRemain(t *testing.T) {
	app := &cancelScreen{touchPressScreen: touchPressScreen{fakeScreen: fakeScreen{width: 200, height: 200}}}
	s := &shell{app: app, ctx: context.Background(), screenW: 200, screenH: 200}
	if err := s.applyTouches([]touchPos{{id: 1, x: 20, y: 20}}, false, 200, 200); err != nil {
		t.Fatal(err)
	}
	app.cancel = true
	for _, ids := range [][]ebiten.TouchID{{1}, {1}, nil} {
		handled, err := s.consumeTouchCancel(ids)
		if err != nil || !handled {
			t.Fatalf("cancel %v %v", handled, err)
		}
	}
	if handled, _ := s.consumeTouchCancel(nil); handled {
		t.Fatal("cancellation never cleared")
	}
	if app.clicks != 0 || app.releases != 1 || s.hold.armed || len(s.fingers.fingers) != 0 {
		t.Fatal("cancel delivered tap or retained press")
	}
}
