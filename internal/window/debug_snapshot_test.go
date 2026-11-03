package window

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
)

func TestDebugDisabledPublishesNothing(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.maybePublishDebug()
	if got := s.debugSnapshot.Load(); got != nil {
		t.Fatalf("disabled debug published a snapshot: %+v", got)
	}
}

func TestDebugPublishCarriesBoxesAndStats(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.debugEnabled = true
	s.maybePublishDebug()
	snap := s.debugSnapshot.Load()
	if snap == nil {
		t.Fatal("enabled debug published nothing")
	}
	if snap.Width != 800 || snap.Height != 600 {
		t.Fatalf("size = %dx%d, want 800x600", snap.Width, snap.Height)
	}
	if len(snap.Boxes) != 1 || snap.Boxes[0].ID != "known" {
		t.Fatalf("boxes = %+v, want the known box", snap.Boxes)
	}
	if snap.Stats == nil {
		t.Fatal("stats missing")
	}
	if snap.Runtime != nil {
		t.Fatalf("runtime = %v, want nil without Perf", snap.Runtime)
	}
	if snap.Frame != nil {
		t.Fatalf("frame = %v, want nil without Perf", snap.Frame)
	}
}

func TestDebugPublishIncludesRuntimeAndFrameWhenPerf(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.debugEnabled = true
	s.perf = true
	s.maybePublishDebug()
	snap := s.debugSnapshot.Load()
	if snap == nil {
		t.Fatal("enabled debug published nothing")
	}
	if _, ok := snap.Runtime["heap_alloc"]; !ok {
		t.Fatalf("runtime = %v, want heap_alloc", snap.Runtime)
	}
	if _, ok := snap.Runtime["goroutines"]; !ok {
		t.Fatalf("runtime = %v, want goroutines", snap.Runtime)
	}
	if _, ok := snap.Frame["avg_ns"]; !ok {
		t.Fatalf("frame = %v, want avg_ns", snap.Frame)
	}
}

func TestDebugPublishThrottles(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.debugEnabled = true
	s.maybePublishDebug()
	first := s.debugSnapshot.Load()
	s.maybePublishDebug()
	if second := s.debugSnapshot.Load(); second != first {
		t.Fatal("second publish inside the gap replaced the snapshot")
	}
	s.debugLast = time.Now().Add(-debugPublishGap).Add(-time.Second)
	s.maybePublishDebug()
	if third := s.debugSnapshot.Load(); third == first {
		t.Fatal("publish past the gap kept the old snapshot")
	}
}

func TestDebugOpCountsMapNames(t *testing.T) {
	display := &layout.Display{
		Ops: []layout.DisplayOp{
			{Kind: layout.DisplayOpFillRect},
			{Kind: layout.DisplayOpStrokeRect},
			{Kind: layout.DisplayOpLine},
			{Kind: layout.DisplayOpText},
			{Kind: layout.DisplayOpBullet},
			{Kind: layout.DisplayOpImage},
			{Kind: layout.DisplayOpGridRun},
			{Kind: layout.DisplayOpPath},
		},
	}
	counts := debugOpCountsMap(display)
	for _, name := range []string{"fill", "stroke", "line", "text", "bullet", "image", "grid", "path"} {
		if counts[name] != 1 {
			t.Fatalf("%s = %d, want 1 in %v", name, counts[name], counts)
		}
	}
	if counts["total"] != len(display.Ops) {
		t.Fatalf("total = %d, want %d", counts["total"], len(display.Ops))
	}
}

func TestSanitizeDebugLine(t *testing.T) {
	if got := sanitizeDebugLine("a\nb\rc"); got != "a_b_c" {
		t.Fatalf("sanitize = %q, want underscores", got)
	}
	long := strings.Repeat("x", 600)
	if got := sanitizeDebugLine(long); len(got) != 500 {
		t.Fatalf("sanitize length = %d, want 500", len(got))
	}
}

// TestDebugSnapshotRace publishes while readers load, for the -race gate:
// the game thread owns the build, the HTTP thread only loads immutable
// copies, so no mutex guards the exchange.
func TestDebugSnapshotRace(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.debugEnabled = true
	s.perf = true
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 25; j++ {
				if snap := s.debugSnapshot.Load(); snap != nil {
					_ = len(snap.Boxes)
				}
			}
		}()
	}
	for i := 0; i < 25; i++ {
		s.debugLast = time.Time{}
		s.maybePublishDebug()
	}
	wg.Wait()
}
