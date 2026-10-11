package player

import "testing"

func TestBackAndForwardRestoreScreenHistory(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	click(t, app, "nav-search")
	click(t, app, "nav-library")
	click(t, app, "avatar")
	if v := app.View(); v.Nav != "profile" || !v.CanBack || v.CanForward {
		t.Fatalf("history at profile = %+v", v)
	}
	click(t, app, "back")
	if v := app.View(); v.Nav != "library" || !v.CanBack || !v.CanForward {
		t.Fatalf("back to library = %+v", v)
	}
	click(t, app, "back")
	if v := app.View(); v.Nav != "search" || !v.CanBack || !v.CanForward {
		t.Fatalf("back to search = %+v", v)
	}
	click(t, app, "forward")
	if v := app.View(); v.Nav != "library" {
		t.Fatalf("forward to library = %+v", v)
	}
	click(t, app, "nav-radio")
	if v := app.View(); v.Nav != "radio" || !v.CanBack || v.CanForward {
		t.Fatalf("new history branch = %+v", v)
	}
	click(t, app, "forward")
	if v := app.View(); v.Nav != "radio" {
		t.Fatalf("forward crossed a discarded branch = %+v", v)
	}
	redraw(t, app)
}
