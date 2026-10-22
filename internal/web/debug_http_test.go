package web

import (
	"encoding/json"
	"image"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"

	"github.com/chinmay-sawant/ownframe/internal/host"
)

// fakeDebug is a screen with an inspector and a display list.
type fakeDebug struct {
	host.Screen
	display *layout.Display
}

func (f *fakeDebug) Generation() uint64       { return 3 }
func (f *fakeDebug) Display() *layout.Display { return f.display }
func (f *fakeDebug) Image() image.Image       { return nil }
func (f *fakeDebug) DevTools() bool           { return true }
func (f *fakeDebug) SetDevTools(bool)         {}
func (f *fakeDebug) Stats() host.Stats        { return host.Stats{Redraws: 2, Boxes: 1} }
func (f *fakeDebug) Boxes() []layout.Box {
	return []layout.Box{{ID: "a", Tag: "div", X: 1, Y: 2, W: 3, H: 4}}
}

func TestDebugStateHandler(t *testing.T) {
	t.Parallel()

	srv := &server{app: &fakeDebug{display: &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, W: 10, H: 10},
			{Kind: layout.DisplayOpText, W: 10, H: 10},
			{Kind: layout.DisplayOpNoop},
		},
		Width:  100,
		Height: 50,
	}}}

	rec := httptest.NewRecorder()
	srv.debug(rec, httptest.NewRequestWithContext(t.Context(), "GET", "/debug/state", nil))

	var state debugState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}

	if state.Generation != 3 || state.Width != 100 || state.Height != 50 {
		t.Fatalf("state = %+v", state)
	}

	if state.Stats == nil || state.Stats.Redraws != 2 {
		t.Fatalf("stats = %+v", state.Stats)
	}

	if state.Ops["fill"] != 1 || state.Ops["text"] != 1 || state.Ops["total"] != 2 {
		t.Fatalf("ops = %v", state.Ops)
	}

	if len(state.Boxes) != 1 || state.Boxes[0].ID != "a" {
		t.Fatalf("boxes = %+v", state.Boxes)
	}
}
