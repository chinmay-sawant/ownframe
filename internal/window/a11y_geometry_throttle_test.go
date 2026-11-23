package window

import "testing"

// TestGeometryThrottleSkipsUnchangedWindow pins PERF-UIA-01: consecutive
// ticks with an unchanged window skip the geometry query until the periodic
// refresh, and any change queries immediately.
func TestGeometryThrottleSkipsUnchangedWindow(t *testing.T) {
	var g geometryThrottle
	sig := geometrySig{x: 10, y: 20, w: 800, h: 600, scale: 1, inset: 32}
	if !g.due(sig) {
		t.Fatal("first tick must query")
	}
	queries := 0
	for range uiaGeometryRefresh {
		if g.due(sig) {
			queries++
		}
	}
	if queries != 0 {
		t.Fatalf("unchanged window queried %d times within the refresh window, want 0", queries)
	}
	if !g.due(sig) {
		t.Fatal("periodic refresh did not query")
	}
	moved := sig
	moved.x++
	if !g.due(moved) {
		t.Fatal("moved window did not query")
	}
	scaled := moved
	scaled.scale = 1.5
	if !g.due(scaled) {
		t.Fatal("scale change did not query")
	}
	g.forget()
	if !g.due(scaled) {
		t.Fatal("forget did not force a query")
	}
}
