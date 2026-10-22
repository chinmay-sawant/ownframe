package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDebugRuntimeAbsentByDefault(t *testing.T) {
	old := GetRuntimeSnapshot()
	SetRuntimeSnapshot(nil)
	defer func() { SetRuntimeSnapshot(old) }()

	srv := &server{app: &fakeDebug{display: &layout.Display{Width: 10, Height: 10}}}
	rec := httptest.NewRecorder()
	srv.debug(rec, httptest.NewRequestWithContext(t.Context(), "GET", "/debug/state", nil))

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["runtime"]; ok {
		t.Fatalf("runtime present without hook: %v", body["runtime"])
	}
}

func TestDebugRuntimePresentWhenHookSet(t *testing.T) {
	old := GetRuntimeSnapshot()
	SetRuntimeSnapshot(func() map[string]any {
		return map[string]any{"goroutines": 7}
	})
	defer func() { SetRuntimeSnapshot(old) }()

	srv := &server{app: &fakeDebug{display: &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect, W: 1, H: 1},
		},
	}}}
	rec := httptest.NewRecorder()
	srv.debug(rec, httptest.NewRequestWithContext(t.Context(), "GET", "/debug/state", nil))

	var state debugState
	if err := json.Unmarshal(rec.Body.Bytes(), &state); err != nil {
		t.Fatal(err)
	}
	if state.Runtime["goroutines"] != float64(7) {
		t.Fatalf("runtime = %v", state.Runtime)
	}
	if state.Stats == nil || state.Ops["fill"] != 1 {
		t.Fatalf("stats+ops missing: %+v", state)
	}
}
