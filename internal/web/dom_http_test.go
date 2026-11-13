package web

import (
	"context"
	"encoding/json"
	"image"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// domScreen is a screen with an executed source and a writable control.
type domScreen struct {
	host.Screen
	generation uint64
	width      int
	height     int
	source     string
	theme      string
	focused    string
	values     map[string]string
	checked    map[string]bool
	placed     map[string]int
	redraws    int
	clicks     [][2]float64
}

func (f *domScreen) Generation() uint64       { return f.generation }
func (f *domScreen) Size() (int, int)         { return f.width, f.height }
func (f *domScreen) Display() *layout.Display { return nil }
func (f *domScreen) Image() image.Image       { return nil }
func (f *domScreen) ExecutedSource() string   { return f.source }
func (f *domScreen) ThemeSource() string      { return f.theme }
func (f *domScreen) Boxes() []layout.Box {
	return []layout.Box{{ID: "login", Action: "login", X: 10, Y: 20, W: 100, H: 30}}
}
func (f *domScreen) SelectAt(_ context.Context, _, _ float64) error { return nil }
func (f *domScreen) Click(_ context.Context, x, y float64) error {
	f.clicks = append(f.clicks, [2]float64{x, y})
	return nil
}
func (f *domScreen) Focus(_ context.Context, id string) error {
	f.focused = id
	return nil
}
func (f *domScreen) SetFormValue(id, value string) {
	if f.values == nil {
		f.values = map[string]string{}
	}
	f.values[id] = value
}
func (f *domScreen) PlaceCaret(id string, caret int) {
	if f.placed == nil {
		f.placed = map[string]int{}
	}
	f.placed[id] = caret
}
func (f *domScreen) SetFormChecked(id string, checked bool) {
	if f.checked == nil {
		f.checked = map[string]bool{}
	}
	f.checked[id] = checked
}
func (f *domScreen) Redraw(_ context.Context) error {
	f.redraws++
	return nil
}

func TestDomEndpointSnapshot(t *testing.T) {
	t.Parallel()

	app := &domScreen{
		generation: 4,
		width:      480,
		height:     640,
		source:     "<html><body><p id=\"m\">hi</p></body></html>",
		theme:      ":root{}",
	}
	srv := &server{app: app}
	rec := httptest.NewRecorder()
	srv.dom(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/dom", nil))

	var state domState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Generation != 4 || state.Width != 480 || state.Height != 640 {
		t.Fatalf("state = %+v", state)
	}
	if state.Fallback || !strings.Contains(state.HTML, "<p id=\"m\">hi</p>") || state.Theme != ":root{}" {
		t.Fatalf("state = %+v", state)
	}
	if len(state.Selects) != 0 {
		t.Fatalf("selects = %v", state.Selects)
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
}

func TestDomEndpointFallback(t *testing.T) {
	t.Parallel()

	srv := &server{app: &stateScreen{generation: 3, width: 100, height: 50}}
	rec := httptest.NewRecorder()
	srv.dom(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/dom", nil))

	var state domState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if !state.Fallback || state.HTML != "" {
		t.Fatalf("state = %+v", state)
	}
}

func TestChangeEndpoint(t *testing.T) {
	t.Parallel()

	origin := "http://example.com"

	app := &domScreen{values: map[string]string{"email": "a"}}
	srv := &server{app: app}

	// Focus only: no value means the field keeps its text.
	if rec := formRec(t, srv.change, "/change", "id=email", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("focus: %d %q", rec.Code, rec.Body.String())
	}
	if app.focused != "email" || app.values["email"] != "a" || app.redraws != 1 {
		t.Fatalf("focused=%q values=%v redraws=%d", app.focused, app.values, app.redraws)
	}

	if rec := formRec(t, srv.change, "/change", "id=email&value=b", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("value: %d %q", rec.Code, rec.Body.String())
	}
	if app.values["email"] != "b" || app.redraws != 2 {
		t.Fatalf("values=%v redraws=%d", app.values, app.redraws)
	}

	if rec := formRec(t, srv.change, "/change", "id=remember&checked=1", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("checked: %d %q", rec.Code, rec.Body.String())
	}
	if !app.checked["remember"] {
		t.Fatalf("checked=%v", app.checked)
	}

	// A caret without a value only aligns, so it does not redraw.
	redraws := app.redraws
	if rec := formRec(t, srv.change, "/change", "id=email&caret=2", origin); rec.Code != http.StatusNoContent {
		t.Fatalf("caret: %d %q", rec.Code, rec.Body.String())
	}
	if app.placed["email"] != 2 || app.redraws != redraws {
		t.Fatalf("placed=%v redraws=%d", app.placed, app.redraws)
	}

	for _, bad := range []string{"id=email&caret=x", "id=email&caret=-1", "id=email&caret=12345678901"} {
		if rec := formRec(t, srv.change, "/change", bad, origin); rec.Code != http.StatusBadRequest {
			t.Fatalf("%s: %d %q", bad, rec.Code, rec.Body.String())
		}
	}

	if rec := formRec(t, srv.change, "/change", "id=<bad>", origin); rec.Code != http.StatusBadRequest {
		t.Fatalf("bad id: %d %q", rec.Code, rec.Body.String())
	}
	if rec := formRec(t, srv.change, "/change", "id=email&value=b", ""); rec.Code != http.StatusForbidden {
		t.Fatalf("missing origin: %d %q", rec.Code, rec.Body.String())
	}
}

func TestClickByID(t *testing.T) {
	t.Parallel()

	origin := "http://example.com"
	app := &domScreen{}
	srv := &server{app: app}

	if rec := formRec(t, srv.click, "/click", "id=login", origin); rec.Code != http.StatusSeeOther {
		t.Fatalf("click: %d %q", rec.Code, rec.Body.String())
	}
	if len(app.clicks) != 1 || app.clicks[0][0] != 60 || app.clicks[0][1] != 35 {
		t.Fatalf("clicks = %v", app.clicks)
	}

	if rec := formRec(t, srv.click, "/click", "action=login", origin); rec.Code != http.StatusSeeOther {
		t.Fatalf("action click: %d %q", rec.Code, rec.Body.String())
	}
	if len(app.clicks) != 2 {
		t.Fatalf("clicks = %v", app.clicks)
	}

	// Proportional offsets land inside the box instead of the center.
	if rec := formRec(t, srv.click, "/click", "id=login&ox=0&oy=0&w=200&h=60", origin); rec.Code != http.StatusSeeOther {
		t.Fatalf("offset click: %d %q", rec.Code, rec.Body.String())
	}
	if last := app.clicks[2]; last[0] != 10 || last[1] != 20 {
		t.Fatalf("clicks = %v", app.clicks)
	}

	// Out-of-range offsets fall back to the center.
	if rec := formRec(t, srv.click, "/click", "id=login&ox=999&oy=0&w=200&h=60", origin); rec.Code != http.StatusSeeOther {
		t.Fatalf("bad offset click: %d %q", rec.Code, rec.Body.String())
	}
	if last := app.clicks[3]; last[0] != 60 || last[1] != 35 {
		t.Fatalf("clicks = %v", app.clicks)
	}
}
