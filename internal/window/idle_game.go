package window

import (
	"context"
	"image"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/safego"
	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"github.com/hajimehoshi/ebiten/v2"
)

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
	once sync.Once
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
		shell:  game,
		policy: pacing.New(time.Now()),
		base:   ebiten.TPS(),
		maxFPS: maxFPS,
		noIdle: options.NoIdlePace,
		stop:   make(chan struct{}),
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
	safego.Go("pace timer", func() { pace.timer(ctx) })

	return pace, pace.close
}

func (g *paceGame) close() {
	g.once.Do(func() { close(g.stop) })
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
	// no busy-spin). Every iteration presents once; Updates run hotter via
	// TPS headroom, so the FE shows MaxFPS while the app ticks at
	// MaxFPS+20. TPS/ActualTPS carries the hotter rate, FPS the cap.
	now := time.Now()
	if wait := loopWait(g.maxFPS, g.lastLoop, now); wait > 0 {
		time.Sleep(wait)
		now = time.Now()
	}
	g.lastLoop = now
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
