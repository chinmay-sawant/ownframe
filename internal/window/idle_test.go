package window

import (
	"context"
	"github.com/chinmay-sawant/blinkless/layout"
	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"strings"
	"testing"
	"time"
)

func TestPaceChromeUnchanged(t *testing.T) {
	g := &paceGame{}
	bar := []image.Rectangle{image.Rect(1, 1, 9, 60)}

	if g.chromeChanged(nil) {
		t.Fatal("no chrome twice must read unchanged")
	}

	if !g.chromeChanged(bar) {
		t.Fatal("a new scrollbar must read changed")
	}

	if g.chromeChanged(bar) {
		t.Fatal("a still scrollbar must read unchanged")
	}

	moved := []image.Rectangle{image.Rect(1, 20, 9, 79)}
	if !g.chromeChanged(moved) {
		t.Fatal("a moved scrollbar must read changed")
	}

	if !g.chromeChanged(nil) {
		t.Fatal("a dropped scrollbar must read changed")
	}

	if g.chromeChanged(nil) {
		t.Fatal("no chrome again must read unchanged")
	}
}

func TestPaceWindowOption(t *testing.T) {
	saved := ebiten.TPS()
	defer ebiten.SetTPS(saved)
	sh := &shell{ctx: context.Background()}

	out, stop := paceWindow(context.Background(), sh, Options{NoIdlePace: true, MaxFPS: -1})
	stop()
	if ebiten.TPS() != ebiten.SyncWithFPS {
		t.Fatal("uncapped input updates must follow rendered frames")
	}

	if out != sh {
		t.Fatal("NoIdlePace uncapped must not wrap the game")
	}

	out, stop = paceWindow(context.Background(), sh, Options{})
	defer stop()

	if _, ok := out.(*paceGame); !ok {
		t.Fatalf("default pacing wrapped %T, want *paceGame", out)
	}
}

func TestPaceWindowCappedNoIdleStillWraps(t *testing.T) {
	sh := &shell{ctx: context.Background()}

	// paceWindow with a cap mutates the global TPS; restore it so later
	// tests in this package see the TPS they started with.
	saved := ebiten.TPS()
	defer ebiten.SetTPS(saved)

	out, stop := paceWindow(context.Background(), sh, Options{NoIdlePace: true, MaxFPS: 60})
	defer stop()

	pg, ok := out.(*paceGame)
	if !ok {
		t.Fatalf("capped NoIdlePace wrapped %T, want *paceGame", out)
	}
	if pg.maxFPS != 60 {
		t.Fatalf("capped maxFPS = %d, want 60", pg.maxFPS)
	}
	if !pg.noIdle {
		t.Fatal("capped NoIdlePace must set noIdle")
	}
}

func TestIdleCleanFrameProbe(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.idleCleanFrame() {
		t.Fatal("a buffer rebuild must read dirty")
	}

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if !s.idleCleanFrame() {
		t.Fatal("a blit-only frame must read clean")
	}

	screen.click(image.Rect(10, 10, 30, 30))
	s.drawReplayPartial(dst)
	s.frameCommit()

	if s.idleCleanFrame() {
		t.Fatal("a repainted frame must read dirty")
	}

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	s.dev.on = true
	screen.width, screen.height = s.viewSize()

	if !s.idleCleanFrame() {
		t.Fatal("a stable page with open devtools must still idle")
	}

	s.dev.on = false
	s.fallback = true

	if s.idleCleanFrame() {
		t.Fatal("a bitmap fallback must read dirty")
	}
}

func TestIdleCleanFrameTicking(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.gen = 1
	screen.ticking = true
	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s := newDirtyShell(screen)
	s.frameOptIn = true
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)
	s.frameCommit()

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if !s.idleCleanFrame() {
		t.Fatal("a ticking page that opted in must read clean")
	}

	s.frameOptIn = false
	s.drawReplayPartial(dst)
	s.frameCommit()

	if !s.idleCleanFrame() {
		t.Fatal("a ticking page with no damage must still idle")
	}
}

func TestFilteredFrameIdleProbeTracksBackdropChanges(t *testing.T) {
	t.Parallel()

	screen := &dirtyScreen{fakeScreen: &fakeScreen{width: 100, height: 100}}
	screen.display = partialDisplay(100, 100)
	screen.display.Ops[1].SetBlendGroup(&layout.DisplayGroup{BackdropFilter: "blur(12px)"})
	screen.gen = 1
	screen.ok = true
	screen.rects = []image.Rectangle{image.Rect(10, 10, 30, 30)}
	s := newDirtyShell(screen)
	dst := partialBuf(100, 100)

	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("first filtered full replay must read dirty")
	}

	screen.ok = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("stable filtered frame must be eligible to idle")
	}

	screen.click(image.Rect(20, 20, 40, 40))
	s.frameOptIn = true
	s.drawReplayPartial(dst)
	s.frameCommit()
	if s.idleCleanFrame() {
		t.Fatal("a filtered frame changed by the ticker must read dirty")
	}

	screen.ok = false
	s.frameOptIn = false
	s.drawReplayPartial(dst)
	s.frameCommit()
	if !s.idleCleanFrame() {
		t.Fatal("filtered frame should become idle eligible after it stabilizes")
	}
}

// TestDevFramePacingRows pins the scheduler state in the Frame tab.
func TestDevFramePacingRows(t *testing.T) {
	s := newDevShell(newDevScreen())
	s.pace.idle = true
	s.pace.skipped = 41
	s.pace.next = time.Now().Add(750 * time.Millisecond)

	rows := s.devFrameRows(320)

	if got := devFrameValue(rows, "Idle"); got != "true" {
		t.Fatalf("Idle row = %q, want true", got)
	}
	if got := devFrameValue(rows, "Pace skip"); got != "41" {
		t.Fatalf("Pace skip row = %q, want 41", got)
	}
	if got := devFrameValue(rows, "Next"); !strings.HasSuffix(got, "ms") {
		t.Fatalf("Next row = %q, want a millisecond value", got)
	}
}

// TestDevPaceNextEmpty pins the fast-path rendering.
func TestDevPaceNextEmpty(t *testing.T) {
	if got := devPaceNext(time.Time{}); got != "-" {
		t.Fatalf("devPaceNext(zero) = %q, want -", got)
	}
}

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

// wantScreen is a screen that reports a next-frame deadline, the way
// *page.Page does. It clears the deadline when it is read.
type wantScreen struct {
	*fakeScreen
	at    time.Time
	reads int
}

func (w *wantScreen) NextFrameAt() time.Time {
	w.reads++
	at := w.at
	w.at = time.Time{}

	return at
}

// TestPaceCaptureWithoutWanter pins that a custom screen without the
// interface keeps the keep-alive fallback: no deadline is captured.
func TestPaceCaptureWithoutWanter(t *testing.T) {
	g := &paceGame{shell: &shell{app: &fakeScreen{width: 10, height: 10}}}
	g.captureFrameWant()

	if !g.want.IsZero() {
		t.Fatal("a screen without host.FrameWanter must not request a frame")
	}
}

// TestPaceCaptureReadsWanterOnce pins that the deadline is read once per tick
// and clears on read.
func TestPaceCaptureReadsWanterOnce(t *testing.T) {
	want := time.Unix(10000, 0)
	app := &wantScreen{fakeScreen: &fakeScreen{width: 10, height: 10}, at: want}
	g := &paceGame{shell: &shell{app: app}}

	g.captureFrameWant()
	if !g.want.Equal(want) {
		t.Fatalf("want = %v, want %v", g.want, want)
	}
	if app.reads != 1 {
		t.Fatalf("NextFrameAt read %d times, want 1", app.reads)
	}

	g.captureFrameWant()
	if !g.want.IsZero() {
		t.Fatal("a consumed deadline must not repeat")
	}
}

func TestNextLoopBaseAbsorbsOvershootAndResyncsOnStall(t *testing.T) {
	t.Parallel()

	t0 := time.Unix(1000, 0)
	interval := LoopInterval(60)

	// First frame anchors to now.
	base := nextLoopBase(60, time.Time{}, t0)
	if !base.Equal(t0) {
		t.Fatalf("first base = %v, want %v", base, t0)
	}

	// Simulate 100 frames where time.Sleep oversleeps by 2ms each frame:
	// the pacing anchor must advance by exactly 100*interval with zero drift.
	overshoot := 2 * time.Millisecond
	for range 100 {
		woke := base.Add(interval).Add(overshoot)
		base = nextLoopBase(60, base, woke)
	}
	want := t0.Add(100 * interval)
	if !base.Equal(want) {
		t.Fatalf("drifted base = %v, want exact %v (drift=%v)", base, want, base.Sub(want))
	}

	// A stall >= 1 interval past target re-anchors to now so no catch-up burst occurs.
	stalled := base.Add(3 * interval)
	resynced := nextLoopBase(60, base, stalled)
	if !resynced.Equal(stalled) {
		t.Fatalf("stalled base = %v, want resynced %v", resynced, stalled)
	}
}

// TestPaceGameCloseJoinsTimerGoroutine pins GPU-LIFE-02: close returns only
// after the timer goroutine has exited.
func TestPaceGameCloseJoinsTimerGoroutine(t *testing.T) {
	sh := &shell{ctx: context.Background()}
	saved := ebiten.TPS()
	defer ebiten.SetTPS(saved)

	out, stop := paceWindow(context.Background(), sh, Options{NoIdlePace: true, MaxFPS: 60})
	pg, ok := out.(*paceGame)
	if !ok {
		t.Fatalf("wrapped %T, want *paceGame", out)
	}
	stop()
	select {
	case <-pg.timerDone:
	default:
		t.Fatal("close returned before the timer goroutine exited")
	}
}

// TestDisposeSurfacesIsIdempotentOnBareShell pins that the run-exit disposal
// tolerates a shell that never built any surface, and runs twice.
func TestDisposeSurfacesIsIdempotentOnBareShell(t *testing.T) {
	sh := &shell{ctx: context.Background()}
	sh.disposeSurfaces()
	sh.disposeSurfaces()
	if sh.img != nil || sh.scaleBuf != nil || sh.partial.tiles != nil || sh.bitmapView.img != nil || sh.titlebar.canvas != nil {
		t.Fatal("surfaces remain after disposeSurfaces")
	}
}

// TestLoopWaitNeverExceedsOnePresentInterval pins GPU-LIFE-03: the sleep in
// paceGame.Draw is bounded by one present interval for any anchor, so a
// MaxFPS cap of 30 delays OS event pumping by at most ~33 ms.
func TestLoopWaitNeverExceedsOnePresentInterval(t *testing.T) {
	for _, fps := range []int{10, 24, 30, 60} {
		interval := LoopInterval(fps)
		now := time.Now()
		for _, ago := range []time.Duration{0, interval / 4, interval / 2, interval, 10 * interval} {
			if wait := loopWait(fps, now.Add(-ago), now); wait < 0 || wait > interval {
				t.Fatalf("fps %d anchor %v ago: wait %v outside [0,%v]", fps, ago, wait, interval)
			}
		}
	}
	if got := LoopInterval(30); got > 34*time.Millisecond {
		t.Fatalf("30 FPS interval %v exceeds the 34 ms input budget", got)
	}
}
