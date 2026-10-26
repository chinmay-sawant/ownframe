package web

import (
	"context"
	"net/http"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

type wheelCall struct {
	x, y, dx, dy float64
}

// wheelScreen records wheel events the browser forwards.
type wheelScreen struct {
	host.Screen
	calls []wheelCall
}

func (f *wheelScreen) Wheel(_ context.Context, x, y, dx, dy float64) (bool, error) {
	f.calls = append(f.calls, wheelCall{x: x, y: y, dx: dx, dy: dy})

	return true, nil
}

var _ host.WheelHandler = (*wheelScreen)(nil)

func TestWheelEndpoint(t *testing.T) {
	t.Parallel()

	app := &wheelScreen{}
	srv := &server{app: app}
	origin := "http://example.com"

	rec := formRec(t, srv.wheel, "/wheel", "x=12.5&y=8&dx=-1&dy=2.25", origin)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("wheel: %d %q", rec.Code, rec.Body.String())
	}
	if len(app.calls) != 1 {
		t.Fatalf("calls = %v", app.calls)
	}
	got := app.calls[0]
	if got.x != 12.5 || got.y != 8 || got.dx != -1 || got.dy != 2.25 {
		t.Fatalf("call = %+v", got)
	}

	if rec := formRec(t, srv.wheel, "/wheel", "x=0&y=0&dx=NaN&dy=1", origin); rec.Code != http.StatusBadRequest {
		t.Fatalf("NaN: %d", rec.Code)
	}
	if rec := formRec(t, srv.wheel, "/wheel", "x=0&y=0&dx=1", origin); rec.Code != http.StatusBadRequest {
		t.Fatalf("missing dy: %d", rec.Code)
	}
	if rec := formRec(t, srv.wheel, "/wheel", "x=0&y=0&dx=1&dy=1", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("missing origin: %d", rec.Code)
	}
}
