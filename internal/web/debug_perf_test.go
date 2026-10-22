package web

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDebugRuntimeAbsentByDefault(t *testing.T) {
	srv := &server{app: &fakeDebug{display: &layout.Display{Width: 10, Height: 10}}}
	rec := httptest.NewRecorder()
	srv.debug(rec, httptest.NewRequestWithContext(t.Context(), "GET", "/debug/state", nil))

	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["runtime"]; ok {
		t.Fatalf("runtime present without perf opt-in: %v", body["runtime"])
	}
}

func TestDebugRuntimePresentWhenPerfEnabled(t *testing.T) {

	srv := &server{perf: true, app: &fakeDebug{display: &layout.Display{
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
	if state.Runtime["goroutines"] == nil || state.Runtime["heap_alloc"] == nil {
		t.Fatalf("runtime = %v", state.Runtime)
	}
	if state.Stats == nil || state.Ops["fill"] != 1 {
		t.Fatalf("stats+ops missing: %+v", state)
	}
}
