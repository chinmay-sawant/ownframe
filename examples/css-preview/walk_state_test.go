package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// walkStateFields calls the /state route and returns the fields it wrote.
func walkStateFields(t *testing.T, w *walker) []string {
	t.Helper()
	rec := httptest.NewRecorder()
	w.handleState(rec, httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/state", nil))
	return strings.Fields(rec.Body.String())
}

// TestWalkStateFieldOrder pins the field order of the /state response.
// scripts/css-preview-click-test.py reads the fields by position, so a
// reorder of the route breaks the click script.
func TestWalkStateFieldOrder(t *testing.T) {
	w := &walker{next: 3, trace: walkTrace{
		started: 1, selected: "opacity", last: 2, redraws: 3, parses: 4,
		cascades: 5, layouts: 6, repaints: 7, ops: 8, boxes: 9, tmpl: 10,
		style: 11, display: 12, paint: 13, changed: 14, dirty: 15,
		alloc: 16, startedNS: 17, seq: 18,
	}}
	want := "3 1 opacity 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18"
	if got := strings.Join(walkStateFields(t, w), " "); got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}
