package window

import (
	"context"
	"image"
	"testing"
	"time"

	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestViewportQuietTickParks(t *testing.T) {
	t.Parallel()

	// Desktop Run turns clearing off. A host that clears every frame marks
	// the blit stale, so this test matches that desktop setting.
	ebiten.SetScreenClearedEveryFrame(false)

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 22000)
	screen.gen = 1
	screen.ticking = true
	screen.ok = false
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)
	s.frameCommit()
	settleViewportWarmup(s, dst)

	const frames = 10
	policy := pacing.New(time.Now())
	at := time.Now()
	idled := false

	for i := 0; i < frames; i++ {
		s.frameOptIn = false
		screen.ok = false
		screen.rects = nil
		s.drawReplayPartial(dst)
		s.frameCommit()

		if !s.viewport.blitOnly {
			t.Fatalf("quiet tick frame %d was not blit-only", i)
		}
		if !s.idleCleanFrame() {
			t.Fatalf("quiet tick frame %d must read clean", i)
		}

		at = at.Add(16 * time.Millisecond)
		if act := policy.Step(at, false, false, time.Time{}); act == pacing.ToIdle {
			idled = true
		}
	}

	if !idled {
		t.Fatal("a quiet tick never reached the idle park")
	}
	if !policy.Snapshot().Idle {
		t.Fatal("a quiet tick left the policy on the frame clock")
	}
}

func TestViewportDamagingTickStaysOnClock(t *testing.T) {
	t.Parallel()

	ebiten.SetScreenClearedEveryFrame(false)

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 1920, height: 1080}}
	screen.display = partialDisplay(1920, 22000)
	screen.gen = 1
	screen.ticking = true
	screen.ok = false
	s := newDirtyShell(screen)
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)
	s.frameCommit()
	settleViewportWarmup(s, dst)

	// A tick that set FrameDirty stays dirty even with no dirty rect.
	s.frameOptIn = true
	screen.ok = false
	screen.rects = nil
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a ticking viewport that set FrameDirty must read dirty")
	}

	// A tick that reported a dirty rect stays dirty.
	s.frameOptIn = false
	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a ticking viewport with damage must read dirty")
	}

	// Sustained damage never idles the pacing policy.
	policy := pacing.New(time.Now())
	at := time.Now()
	for i := 0; i < 600; i++ {
		at = at.Add(16 * time.Millisecond)
		if act := policy.Step(at, true, false, time.Time{}); act == pacing.ToIdle {
			t.Fatalf("damaging tick idled at frame %d", i)
		}
	}
	if policy.Snapshot().Idle {
		t.Fatal("a damaging tick left the fast path")
	}
}

// tickNoMarker ticks but cannot report FrameDirty. It keeps the viewport on
// the frame clock: when in doubt the probe stays dirty.
type tickNoMarker struct {
	*fakeScreen
	display *layout.Display
	gen     uint64
	rects   []image.Rectangle
	ok      bool
}

func (d *tickNoMarker) Display() *layout.Display { return d.display }

func (d *tickNoMarker) Generation() uint64 { return d.gen }

func (d *tickNoMarker) Ticking() bool { return true }

func (d *tickNoMarker) TakeDirtyRects() ([]image.Rectangle, bool) { return d.rects, d.ok }

func TestViewportTickWithoutMarkerStaysDirty(t *testing.T) {
	t.Parallel()

	ebiten.SetScreenClearedEveryFrame(false)

	screen := &tickNoMarker{
		fakeScreen: &fakeScreen{width: 1920, height: 1080},
		display:    partialDisplay(1920, 22000),
		gen:        1,
	}
	s := &shell{
		app: screen, ctx: context.Background(),
		display: screen.display, seq: screen.gen,
		screenW: screen.width, screenH: screen.height,
	}
	dst := partialBuf(1920, 1080)

	s.drawReplayPartial(dst)
	s.frameCommit()
	settleViewportWarmup(s, dst)

	// A marker-less ticker stays on the full-replay path, so the probe
	// reads dirty even when the viewport would otherwise blit.
	if s.idleCleanFrame() {
		t.Fatal("a ticking viewport without FrameDirty must read dirty")
	}
}
