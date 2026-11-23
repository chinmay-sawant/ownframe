package window

import (
	"context"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/debugtag"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/safego"
	"github.com/chinmay-sawant/ownframe/internal/textrun"
	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"os"
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// chromeChanged reports chrome that differs from the previous set and
// records the new set. An unchanged overlay repaints the same pixels, so it
// does not need the active drawing loop. The buffers swap, allocating nothing.
func (g *paceGame) chromeChanged(restore []image.Rectangle) bool {
	g.chromeNext = append(g.chromeNext[:0], restore...)
	changed := !slices.Equal(g.chromeNext, g.chrome)
	g.chrome, g.chromeNext = g.chromeNext, g.chrome

	return changed
}

// viewMoved reports a focus or screen size change, either of which makes
// the replay paint in full.
func (g *paceGame) viewMoved(screen *ebiten.Image) bool {
	focus := ebiten.IsFocused()
	size := screen.Bounds().Size()
	moved := !g.seen || focus != g.focus || size != g.size
	g.seen, g.focus, g.size = true, focus, size

	return moved
}

// paceDebugEnv turns on one pacing line per second on stderr, so a headless
// CPU run can report the present rate without opening the devtools dock.
// It is dev-only and off unless explicitly set; keep it off in shipped builds.
const paceDebugEnv = "OWNFRAME_PACE_DEBUG"

// paceDebugWriter returns stderr when the debug environment variable is set
// and nil otherwise. It is read once, when the wrapper starts. The variable
// is ignored without -tags ownframe_debug.
func paceDebugWriter() *os.File {
	if !debugtag.Enabled {
		return nil
	}
	if os.Getenv(paceDebugEnv) == "" {
		return nil
	}

	return os.Stderr
}

// debugLine writes presents per second, the skipped total, and the idle flag
// once a second while the debug variable is set. The counts come from the
// publish that just read them under the policy lock.
func (g *paceGame) debugLine(now time.Time, presents, skipped uint64, idle bool) {
	if g.debug == nil || now.Sub(g.debugAt) < time.Second {
		return
	}

	dp := presents - g.debugPresents
	g.debugAt, g.debugPresents = now, presents

	_, _ = fmt.Fprintf(g.debug, "pace: presents=%d/s skipped=%d idle=%v\n", dp, skipped, idle)
}

// frame decides this frame's pacing after the shell has drawn.
func (g *paceGame) frame(screen *ebiten.Image, now time.Time) {
	wasIdle := g.idle
	wasTrimmed := g.trimmed
	clean := g.shell.idleCleanFrame()
	chrome := g.chromeChanged(g.shell.frame.restore)
	dirty := !clean || chrome
	if !dirty {
		dirty = g.viewMoved(screen)
	}

	flinging := g.shell.scrollMotion.vx != 0 || g.shell.scrollMotion.vy != 0
	input := g.input || g.watch.held() || flinging
	g.input = false
	if g.idle && !dirty && !input && !g.trimmed && now.Sub(g.idleSince) >= idleTrimDelay {
		g.shell.renderer.SetIdle()
		textrun.Trim()
		g.trimmed = true
	}

	g.mu.Lock()
	act := g.policy.Step(now, dirty, input, g.frameDeadline())
	g.mu.Unlock()

	switch act {
	case pacing.ToIdle:
		if !g.noIdle {
			g.applyIdle()
			g.mu.Lock()
			g.policy.SetDeadline(g.frameDeadline())
			g.mu.Unlock()
		}
	case pacing.ToFast:
		g.applyFast()
	case pacing.Burst:
		//nolint:staticcheck // reason: ScheduleFrame is the only wake-up for FPSModeVsyncOffMinimum event-wait; SetScreenClearedEveryFrame(false) only skips clearing and keeps the full-rate loop.
		ebiten.ScheduleFrame()
	}

	g.kickTimer()
	g.publish(now)
	if g.shell.debugEnabled && (wasIdle != g.idle || wasTrimmed != g.trimmed) {
		g.shell.publishDebug()
	}
}

// frameDeadline keeps metrics live even when the page asks for a later frame.
func (g *paceGame) frameDeadline() time.Time {
	want := g.want
	cache := &g.shell.dev.panelCache
	if g.shell.dev.on && cache.valid && !cache.next.IsZero() && (want.IsZero() || cache.next.Before(want)) {
		want = cache.next
	}
	if g.idle && g.eventDriven && !g.trimmed {
		trim := g.idleSince.Add(idleTrimDelay)
		if want.IsZero() || trim.Before(want) {
			want = trim
		}
	}
	return want
}

const idleTrimDelay = 5 * time.Second

// applyIdle leaves the active drawing loop: the backend blocks in WaitEvents, and a
// scheduled post or an input event wakes it. TPS syncs with the frame rate
// there, so one wake-up runs one Update. A plain page only blits its tiles
// while parked. A separate deadline trims caches after sustained inactivity,
// including with DevTools open, rather than forcing GC on every short pause.
func (g *paceGame) applyIdle() {
	if g.idle {
		return
	}

	g.idle = true
	g.idleSince = time.Now()
	g.trimmed = false
	//nolint:staticcheck // reason: needs FPSModeVsyncOffMinimum event-wait; SetVsyncEnabled(false) selects Maximum busy loop and would pin CPU.
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.SetTPS(ebiten.SyncWithFPS)
}

// applyFast restores the active drawing loop (still VSync-off maximum at the
// Ebiten level; the present cap is enforced by the Draw limiter, not VSync)
// and the loop TPS: the capped loop rate when a cap is set, else SyncWithFPS so input is sampled once per rendered frame. The pixel-cache containers are pre-sized first, so the
// first post-wake paints do not regrow them under the draw lock.
func (g *paceGame) applyFast() {
	if !g.idle {
		return
	}

	g.idle = false
	if g.shell != nil {
		g.shell.renderer.Wake()
	}
	ebiten.SetVsyncEnabled(false)
	if loop := LoopFPS(g.maxFPS); loop > 0 {
		ebiten.SetTPS(loop)
		return
	}
	ebiten.SetTPS(g.base)
}

// paceGame wraps the shell and moves the window between the active drawing loop and
// the minimum FPS mode. Ebiten runs Update and Draw on its game goroutine,
// where SetFPSMode and SetTPS are safe; the keep-alive timer runs on its
// own goroutine and only calls the concurrent-safe ScheduleFrame.
//
// Present capping (no VSync): heavy presents are capped to maxFPS by pacing
// every loop iteration to the present interval with a monotonic sleep, so
// the FE shows maxFPS while the app ticks faster via TPS headroom. Every
// iteration presents once; the retained screen keeps the last frame between
// keep-alive wakes. maxFPS==0 is uncapped (legacy uncapped drawing).
type paceGame struct {
	shell       *shell
	policy      pacing.Policy
	watch       inputWatch
	input       bool
	base        int
	maxFPS      int
	noIdle      bool
	idle        bool
	idleSince   time.Time
	trimmed     bool
	eventDriven bool
	wake        atomic.Bool
	skipFrame   bool
	focus       bool
	size        image.Point
	seen        bool
	// chrome and chromeNext hold the previous and current frame's chrome
	// rects, so an unchanged overlay reads as no work.
	chrome     []image.Rectangle
	chromeNext []image.Rectangle
	// lastLoop gates loop iterations to the loop rate; lastPresent gates
	// heavy presents to the present cap. presents counts heavy presents for
	// the Frame tab; presentFPS is the 1s sliding present rate shown there.
	lastLoop    time.Time
	lastPresent time.Time
	presents    uint64
	presentFPS  float64
	presentHist []int64
	pacePage
	mu   sync.Mutex
	stop chan struct{}
	// timerDone closes when the timer goroutine returns, so close can join it.
	timerDone chan struct{}
	once      sync.Once
}

// paceWindow wraps the shell when pacing is on. The returned stop releases
// the keep-alive goroutine after the game loop ends.
func paceWindow(ctx context.Context, game *shell, options Options) (ebiten.Game, func()) {
	maxFPS := NormalizeMaxFPS(options.MaxFPS)
	if maxFPS == 0 {
		ebiten.SetTPS(ebiten.SyncWithFPS)
	}
	// Uncapped + idle pacing off stays legacy: no wrapper. Capped always
	// wraps so the present cap holds even with NoIdlePace (which then means
	// "present continuously at the cap", never parking).
	if (options.NoIdlePace || !idlePaceSupported) && maxFPS == 0 {
		return game, func() {}
	}

	pace := &paceGame{
		shell:     game,
		policy:    pacing.New(time.Now()),
		base:      ebiten.TPS(),
		maxFPS:    maxFPS,
		noIdle:    options.NoIdlePace,
		stop:      make(chan struct{}),
		timerDone: make(chan struct{}),
	}
	// Capped active path ticks at the loop rate (present cap + headroom) so
	// the app runs hotter than the FE presents. Uncapped updates follow each rendered frame.
	// g.base keeps SyncWithFPS for the uncapped restore path;
	// the capped path always re-applies LoopFPS in applyFast.
	if loop := LoopFPS(maxFPS); loop > 0 {
		ebiten.SetTPS(loop)
	}
	pace.kick = make(chan struct{}, 1)
	pace.debug, pace.debugAt = paceDebugWriter(), time.Now()
	safego.Go("pace timer", func() {
		defer close(pace.timerDone)
		pace.timer(ctx)
	})

	return pace, pace.close
}

// close stops the timer goroutine and waits for it to return, so no wake
// races the GPU teardown that follows the run loop.
func (g *paceGame) close() {
	g.once.Do(func() { close(g.stop) })
	if g.timerDone != nil {
		<-g.timerDone
	}
}

func (g *paceGame) Update() error {
	if err := g.shell.ctx.Err(); err != nil {
		return err
	}
	input := g.watch.changed()
	g.input = g.input || input
	notified := g.wake.Swap(false) || closeRequested.Load()
	// The desktop backend can wake without page work, even in minimum FPS
	// mode. Keep its retained screen and skip our update and replay stages.
	unchanged := g.shell.built() && g.shell.drawSyncErr == nil && ebiten.IsFocused() == g.focus
	sameSize := g.shell.pendingW == g.shell.screenW && g.shell.pendingH == g.shell.screenH
	g.skipFrame = g.idle && g.eventDriven && !input && !notified &&
		unchanged && sameSize
	if g.skipFrame {
		return nil
	}
	if err := g.shell.Update(); err != nil {
		return err
	}

	g.captureFrameWant()

	g.publish(time.Now())

	return nil
}

func (g *paceGame) Draw(screen *ebiten.Image) {
	if g.skipFrame {
		return
	}
	if g.maxFPS <= 0 {
		g.notePresent(time.Now())
		g.shell.Draw(screen)
		g.frame(screen, time.Now())
		return
	}
	if g.idle {
		now := time.Now()
		g.lastLoop = now
		g.lastPresent = now
		g.notePresent(now)
		g.shell.Draw(screen)
		g.frame(screen, now)
		return
	}
	// Pace the loop to the present cap with a monotonic sleep (no VSync,
	// no busy-spin). The sleep runs on Ebiten's game goroutine, so OS events
	// queue until the next Update polls them; loopWait never exceeds one
	// present interval (33 ms at 30 FPS), which is the same latency any
	// frame-capped loop has, and Update runs at LoopFPS (hotter than the
	// cap) so input is still sampled between presents. Tested in
	// TestLoopWaitNeverExceedsOnePresentInterval. Every iteration presents once; Updates run hotter via
	// TPS headroom, so the FE shows MaxFPS while the app ticks at
	// MaxFPS+20. TPS/ActualTPS carries the hotter rate, FPS the cap.
	now := time.Now()
	if wait := loopWait(g.maxFPS, g.lastLoop, now); wait > 0 {
		time.Sleep(wait)
		now = time.Now()
	}
	g.lastLoop = nextLoopBase(g.maxFPS, g.lastLoop, now)
	g.lastPresent = now
	g.notePresent(now)
	g.shell.Draw(screen)
	g.frame(screen, now)
}

// notePresent records one heavy present and refreshes the 1s sliding
// present rate shown in the Frame tab (the FE number).
func (g *paceGame) notePresent(now time.Time) {
	g.presents++
	n := now.UnixNano()
	g.presentHist = append(g.presentHist, n)
	const (
		keep = 1200
		// minPresentSamples is the minimum history entries needed for a
		// present rate: one sample has no span to divide by.
		minPresentSamples = 2
	)
	if len(g.presentHist) > keep {
		copy(g.presentHist, g.presentHist[len(g.presentHist)-keep:])
		g.presentHist = g.presentHist[:keep]
	}
	if len(g.presentHist) < minPresentSamples {
		return
	}
	cutoff := n - int64(time.Second)
	idx := 0
	for i, t := range g.presentHist {
		if t >= cutoff {
			idx = i
			break
		}
		if i == len(g.presentHist)-1 {
			idx = len(g.presentHist) - 1
		}
	}
	count := len(g.presentHist) - idx
	if count < minPresentSamples && idx > 0 {
		idx--
		count++
	}
	if count < minPresentSamples {
		g.presentFPS = 0
		return
	}
	if span := time.Duration(n - g.presentHist[idx]); span > 0 {
		g.presentFPS = float64(count-1) / span.Seconds()
	}
}

func (g *paceGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.shell.Layout(outsideWidth, outsideHeight)
}

// inputWatch reports whether input arrived since the previous call. Ebiten
// gives each tick its own input snapshot, so it must run in Update. The two
// key and rune buffers are distinct and swap, so the comparison never sees
// the append it is comparing against. Just-pressed and just-released keys
// catch a tap that starts and ends inside one tick, when the held set is
// already empty again.
type inputWatch struct {
	keys   []ebiten.Key
	next   []ebiten.Key
	chars  []rune
	runes  []rune
	tap    []ebiten.Key
	cursor image.Point
	touch  int
	ready  bool
	mouse  [3]bool
}

func (w *inputWatch) changed() bool {
	x, y := ebiten.CursorPosition()
	keys := inpututil.AppendPressedKeys(w.next[:0])
	chars := ebiten.AppendInputChars(w.runes[:0])
	tap := len(inpututil.AppendJustPressedKeys(w.tap[:0])) > 0 ||
		len(inpututil.AppendJustReleasedKeys(w.tap[:0])) > 0
	touch := len(ebiten.AppendTouchIDs(nil))
	mouse := [3]bool{
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight),
		ebiten.IsMouseButtonPressed(ebiten.MouseButtonMiddle),
	}
	changed := !w.ready || x != w.cursor.X || y != w.cursor.Y ||
		!slices.Equal(keys, w.keys) || !slices.Equal(chars, w.chars) ||
		tap || touch != w.touch || mouse != w.mouse || justPressedMouse() || wheelMoved()

	w.next, w.keys = w.keys, keys
	w.runes, w.chars = w.chars, chars
	w.cursor, w.touch, w.ready = image.Pt(x, y), touch, true
	w.mouse = mouse

	return changed
}

func (w *inputWatch) held() bool {
	return len(w.keys) > 0 || w.mouse != [3]bool{} || w.touch > 0
}

func justPressedMouse() bool {
	return inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonMiddle)
}

func wheelMoved() bool {
	x, y := ebiten.Wheel()

	return x != 0 || y != 0
}

// idleCleanFrame checks the paint surface used by the last Draw. The
// pacing wrapper handles chrome separately. Retained layers stay on the
// frame clock because they may move without damaging the background.
func (s *shell) idleCleanFrame() bool {
	if s.fallback || s.display == nil {
		return false
	}

	switch s.surface {
	case surfaceBackdrop:
		return s.partial.backdropClean && s.bufferOrderSafe()
	case surfaceViewport:
		return s.viewport.blitOnly && (!s.isTicking() || !s.frameOptIn) &&
			!s.renderer.LayersActive(s.display)
	case surfaceTiles:
		return s.partial.mode == repaintBlit && !s.partial.whole && s.partial.buf != nil &&
			!s.renderer.LayersActive(s.display)
	default:
		return false
	}
}

// paceState is the pacer's published state for the devtools Frame tab: the
// next scheduled idle wake, the 60 Hz frames skipped so far, and whether the
// window is parked. The wrapper writes it on the game goroutine after each
// frame; the tab reads it there. It is a snapshot for inspection rather than a scheduling policy.
type paceState struct {
	next        time.Time
	skipped     uint64
	idle        bool
	eventDriven bool
	maxFPS      int
	presentFPS  float64
	presents    uint64
}

// devPaceNext renders the time until the next scheduled idle wake, or "-"
// when the pacer is on the active drawing loop and nothing is scheduled.
func devPaceNext(next time.Time) string {
	if next.IsZero() {
		return "-"
	}

	if d := time.Until(next); d > 0 {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	return "0ms"
}

// timer posts scheduled frames while the policy is idle. ScheduleFrame is
// concurrent-safe: on the desktop backend it writes to the empty-event pipe
// that the blocked WaitEvents polls, so the next frame runs then. The timer
// sleeps until the policy's next wake: a page deadline when the page asked,
// the keep-alive fallback when it did not, and half a keep-alive on the fast
// path, where the active drawing loop already supplies every frame. The game loop
// kicks the timer after every step, so a policy change never waits out a
// stale sleep.
func (g *paceGame) timer(ctx context.Context) {
	t := time.NewTimer(time.Hour)
	defer t.Stop()
	resetTimer(t, g.sleep())
	var events <-chan struct{}
	if scheduler, ok := g.shell.app.(host.FrameScheduler); ok {
		events = scheduler.FrameEvents()
	}

	for {
		select {
		case <-ctx.Done():
			//nolint:staticcheck // reason: cancellation must release an indefinitely parked desktop loop.
			ebiten.ScheduleFrame()
			return
		case <-g.stop:
			return
		case <-events:
			g.wake.Store(true)
			//nolint:staticcheck // reason: a page notification must wake the desktop event wait.
			ebiten.ScheduleFrame()
		case <-g.kick:
			resetTimer(t, g.sleep())
		case now := <-t.C:
			g.mu.Lock()
			due := g.policy.Due(now)
			g.mu.Unlock()

			if due {
				g.wake.Store(true)
				//nolint:staticcheck // reason: timer wake needs ScheduleFrame to post the empty event; SetScreenClearedEveryFrame(false) cannot wake event-wait.
				ebiten.ScheduleFrame()
			}

			resetTimer(t, g.sleep())
		}
	}
}

// sleep reads the next timer sleep under the policy lock.
func (g *paceGame) sleep() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.policy.Wait(time.Now())
}

// resetTimer re-arms t, draining a value that fired before the stop. Only the
// timer goroutine touches t.
func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}

	if d > 0 {
		t.Reset(d)
	}
}

// pacePage is the wrapper's page-driven state: the deadline captured from the
// page, the timer kick channel, and the debug counters. The pacing wrapper embeds it.
type pacePage struct {
	want          time.Time
	kick          chan struct{}
	debug         *os.File
	debugAt       time.Time
	debugPresents uint64
}

// captureFrameWant reads the page's next-frame request once per tick, the way
// captureFrameOptIn reads the frame-dirty opt-in. The page clears the request
// when it is read, so a page that wants another frame asks again on every
// frame. A screen without host.FrameWanter never asks, and the window keeps
// its keep-alive fallback.
func (g *paceGame) captureFrameWant() {
	g.want = time.Time{}
	polling := false
	if scheduler, ok := g.shell.app.(host.FrameScheduler); ok {
		polling = scheduler.FramePolling()
		g.eventDriven = !polling
		g.mu.Lock()
		g.policy.SetEventDriven(g.eventDriven)
		g.mu.Unlock()
	}

	if wanter, ok := g.shell.app.(host.FrameWanter); ok {
		g.want = wanter.NextFrameAt()
	}
	if polling && g.want.IsZero() {
		g.want = time.Now().Add(compatibilityFrameGap)
	}
}

const compatibilityFrameGap = 125 * time.Millisecond

// publish copies the pacer's state to the shell for the devtools Frame tab.
// The game goroutine writes it; the timer goroutine owns the policy lock.
func (g *paceGame) publish(now time.Time) {
	g.mu.Lock()
	state := g.policy.Snapshot()
	g.mu.Unlock()

	g.shell.pace.next, g.shell.pace.skipped, g.shell.pace.idle = state.Next, state.Skipped, state.Idle
	g.shell.pace.maxFPS, g.shell.pace.presentFPS, g.shell.pace.presents = g.maxFPS, g.presentFPS, g.presents
	g.shell.pace.eventDriven = g.eventDriven
	g.debugLine(now, state.Presents, state.Skipped, state.Idle)
}

// kickTimer asks the timer goroutine to recompute its sleep after the policy
// moved. The channel holds one pending kick, so the game loop never blocks.
func (g *paceGame) kickTimer() {
	select {
	case g.kick <- struct{}{}:
	default:
	}
}
