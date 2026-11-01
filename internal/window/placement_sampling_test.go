package window

import "testing"

func TestPlacementSamplingThrottlesUnchangedGeometry(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	s := &shell{placementID: "sampling-test", lastPlacementValid: true, lastPlacement: currentPlacement()}
	s.syncPlacement()
	sampled := s.lastPlacementSave
	if sampled.IsZero() {
		t.Fatal("unchanged geometry did not advance the sample timer")
	}
	s.syncPlacement()
	if !s.lastPlacementSave.Equal(sampled) {
		t.Fatal("sampled geometry again before the interval elapsed")
	}
}

func BenchmarkPlacementSamplingUnchanged(b *testing.B) {
	b.Setenv("XDG_CONFIG_HOME", b.TempDir())
	s := &shell{placementID: "sampling-bench", lastPlacementValid: true, lastPlacement: currentPlacement()}
	s.syncPlacement()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.syncPlacement()
	}
}
