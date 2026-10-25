package window

import (
	"context"
	"image"
	"sync"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/safego"
	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"github.com/hajimehoshi/ebiten/v2"
)

// paceGame wraps the shell and moves the window between the active drawing loop and
// the minimum FPS mode. Ebiten runs Update and Draw on its game goroutine,
// where SetFPSMode and SetTPS are safe; the keep-alive timer runs on its
// own goroutine and only calls the concurrent-safe ScheduleFrame.
type paceGame struct {
	shell  *shell
	policy pacing.Policy
	watch  inputWatch
	input  bool
	base   int
	idle   bool
	focus  bool
	size   image.Point
	seen   bool
	// chrome and chromeNext hold the previous and current frame's chrome
	// rects, so an unchanged overlay reads as no work.
	chrome     []image.Rectangle
	chromeNext []image.Rectangle
	pacePage
	mu   sync.Mutex
	stop chan struct{}
	once sync.Once
}

// paceWindow wraps the shell when pacing is on. The returned stop releases
// the keep-alive goroutine after the game loop ends.
func paceWindow(ctx context.Context, game *shell, options Options) (ebiten.Game, func()) {
	if options.NoIdlePace || !idlePaceSupported {
		return game, func() {}
	}

	pace := &paceGame{
		shell:  game,
		policy: pacing.New(time.Now()),
		base:   ebiten.TPS(),
		stop:   make(chan struct{}),
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
	if err := g.shell.Update(); err != nil {
		return err
	}

	g.captureFrameWant()

	if g.watch.changed() {
		g.input = true
	}

	g.publish(time.Now())

	return nil
}

func (g *paceGame) Draw(screen *ebiten.Image) {
	g.shell.Draw(screen)
	g.frame(screen, time.Now())
}

func (g *paceGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.shell.Layout(outsideWidth, outsideHeight)
}
