package web

import (
	"encoding/json"
	"image"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// stateScreen is a screen whose picture is the frame size. Its boxes carry
// text that /state must never expose.
type stateScreen struct {
	host.Screen
	generation uint64
	width      int
	height     int
}

func (f *stateScreen) Generation() uint64       { return f.generation }
func (f *stateScreen) Size() (int, int)         { return f.width, f.height }
func (f *stateScreen) Display() *layout.Display { return nil }
func (f *stateScreen) Image() image.Image       { return nil }
func (f *stateScreen) Boxes() []layout.Box {
	return []layout.Box{{ID: "secret-field", Text: "hunter2"}}
}

func TestStateEndpoint(t *testing.T) {
	t.Parallel()

	srv := &server{
		app:     &stateScreen{generation: 9, width: 320, height: 200},
		ticking: true,
	}
	rec := httptest.NewRecorder()
	srv.state(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/state", nil))

	var state frameState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Generation != 9 || state.Width != 320 || state.Height != 200 || !state.Ticking {
		t.Fatalf("state = %+v", state)
	}

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 4 {
		t.Fatalf("body keys = %v", body)
	}
	for _, secret := range []string{"secret-field", "hunter2", "boxes"} {
		if strings.Contains(rec.Body.String(), secret) {
			t.Fatalf("state leaked %q:\n%s", secret, rec.Body.String())
		}
	}
	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Fatalf("Cache-Control = %q", got)
	}
}
