package motion

import "testing"

// TestSetPromotedLayersReusesMap pins PERF-MOT-02: a steady promoted set
// refreshes without allocating a new map each frame.
func TestSetPromotedLayersReusesMap(t *testing.T) {
	r := New()
	ids := []string{"a", "b", "c"}
	r.SetPromotedLayers(ids)
	if allocs := testing.AllocsPerRun(100, func() { r.SetPromotedLayers(ids) }); allocs != 0 {
		t.Fatalf("SetPromotedLayers allocs/op = %v, want 0", allocs)
	}
	r.SetPromotedLayers([]string{"z"})
	if len(r.promoted) != 1 {
		t.Fatalf("promoted = %v, want only z", r.promoted)
	}
}
