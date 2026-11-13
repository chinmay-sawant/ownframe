package telegram_test

import (
	"context"
	"testing"
)

func TestWebModePinsFixedChrome(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	app := newApp(t, ctx)
	app.SetWebMode(true)

	click(t, ctx, app, "", "open-anna")

	for _, off := range []int{0, 250, 800} {
		if changed := app.Pin(off, 800); changed {
			t.Errorf("Pin(%d, 800) = true, want false in web mode", off)
		}

		view := app.View()
		if view.BarTop != 0 || view.BottomTop != 0 || view.PadTop != 0 {
			t.Errorf(
				"Pin(%d, 800) left BarTop=%d BottomTop=%d PadTop=%d, want zeros",
				off,
				view.BarTop,
				view.BottomTop,
				view.PadTop,
			)
		}
	}

	if !app.View().Web {
		t.Error("View.Web = false, want true after SetWebMode(true)")
	}
}
