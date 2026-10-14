package player

import (
	"context"
	"testing"
)

func TestNavRetentionHelperRoundTrip(t *testing.T) {
	app := mustApp(t)
	redraw(t, app)
	p := app.Page()
	p.SetScrollOffset(0, 42)
	_ = p.Focus(context.Background(), "q")
	redraw(t, app)

	st := app.saveNavRetention()
	if st.Focus != "q" || st.ScrollY != 42 {
		t.Fatalf("saved = %+v", st)
	}

	delete(app.navFocusStates, app.view.Nav)
	app.restoreNavRetention("home")
	if got := app.retainedNavState("home").ScrollY; got != 42 {
		t.Fatalf("retained home scroll = %d, want 42", got)
	}

	p.SetScrollOffset(0, 7)
	app.saveNavRetention()
	if got := app.retainedNavState("home").ScrollY; got != 7 {
		t.Fatalf("retained home scroll = %d, want 7", got)
	}
}

func TestNavRetentionHelperNilSafe(t *testing.T) {
	var app *App
	app.saveNavRetention()
	app.restoreNavRetention("")
	app.touchNavRetention("")
	if got := app.retainedNavState(""); got != (navFocusState{}) {
		t.Fatalf("unexpected state: %+v", got)
	}
}
