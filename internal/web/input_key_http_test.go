package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// keyScreen records the keys and focus moves a browser request delivers.
type keyScreen struct {
	host.Screen
	downs   []string
	ups     []string
	nexts   int
	prevs   int
	submits int
}

func (f *keyScreen) KeyDown(_ context.Context, key string) error {
	f.downs = append(f.downs, key)

	return nil
}

func (f *keyScreen) KeyUp(_ context.Context, key string) error {
	f.ups = append(f.ups, key)

	return nil
}

func (f *keyScreen) FocusNext(context.Context) error {
	f.nexts++

	return nil
}

func (f *keyScreen) FocusPrev(context.Context) error {
	f.prevs++

	return nil
}

func (f *keyScreen) Focus(context.Context, string) error { return nil }
func (f *keyScreen) FocusID() string                     { return "field" }

func (f *keyScreen) Submit(context.Context) error {
	f.submits++

	return nil
}

var _ host.Focuser = (*keyScreen)(nil)

// formRec runs one urlencoded POST with a matching Origin.
func formRec(t *testing.T, handler http.HandlerFunc, path, body, origin string) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "http://example.com"+path, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	rec := httptest.NewRecorder()
	handler(rec, r)

	return rec
}

func TestKeyEndpoint(t *testing.T) {
	t.Parallel()

	app := &keyScreen{}
	srv := &server{app: app}
	origin := "http://example.com"

	if rec := formRec(t, srv.key, "/key", "key=ArrowDown&down=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("key down: %d %q", rec.Code, rec.Body.String())
	}
	if rec := formRec(t, srv.key, "/key", "key=arrowdown&down=0", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("key up: %d %q", rec.Code, rec.Body.String())
	}
	if len(app.downs) != 1 || app.downs[0] != "arrowdown" {
		t.Fatalf("downs = %v", app.downs)
	}
	if len(app.ups) != 1 || app.ups[0] != "arrowdown" {
		t.Fatalf("ups = %v", app.ups)
	}

	if rec := formRec(t, srv.key, "/key", "key=tab&down=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("tab: %d %q", rec.Code, rec.Body.String())
	}
	if rec := formRec(t, srv.key, "/key", "key=tab&down=1&shift=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("shift tab: %d %q", rec.Code, rec.Body.String())
	}
	if rec := formRec(t, srv.key, "/key", "key=tab&down=0", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("tab release: %d %q", rec.Code, rec.Body.String())
	}
	if app.nexts != 1 || app.prevs != 1 {
		t.Fatalf("nexts = %d, prevs = %d", app.nexts, app.prevs)
	}
	if len(app.downs) != 1 {
		t.Fatalf("tab reached KeyDown: %v", app.downs)
	}

	if rec := formRec(t, srv.key, "/key", "key=s&down=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("ctrl+s key: %d %q", rec.Code, rec.Body.String())
	}
	if len(app.downs) != 2 || app.downs[1] != "s" {
		t.Fatalf("printable shortcut not forwarded: %v", app.downs)
	}

	if rec := formRec(t, srv.key, "/key", "key=enter&down=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("enter: %d %q", rec.Code, rec.Body.String())
	}
	if app.submits != 1 {
		t.Fatalf("submits = %d, want one", app.submits)
	}

	if rec := formRec(t, srv.key, "/key", "key=bogus&down=1", origin); rec.Code != http.StatusBadRequest {
		t.Fatalf("unknown key: %d", rec.Code)
	}
	if rec := formRec(t, srv.key, "/key", "key="+strings.Repeat("a", maxKeyName+1)+"&down=1", origin); rec.Code != http.StatusBadRequest {
		t.Fatalf("oversized key: %d", rec.Code)
	}
	if rec := formRec(t, srv.key, "/key", "key=arrowdown&down=yes", origin); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad down: %d", rec.Code)
	}
	if rec := formRec(t, srv.key, "/key", "key=arrowdown&down=1", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("missing origin: %d", rec.Code)
	}
}
