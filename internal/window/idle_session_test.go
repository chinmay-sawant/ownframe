//go:build !js && !android && !ios

package window

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestWindowIdleDeadline exercises the actual pacing wrapper and timer in a
// running game. The watchdog posts only on failure, so it cannot supply the
// successful deadline wake that this test requires.
func TestWindowIdleDeadline(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for a graphics session")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, err := page.New(page.Config{
		Width: 128, Height: 96,
		HTML: `<body style="margin:0"><div style="width:128px;height:96px;background:{{.}}"></div></body>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData("red")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	sh := newShell(ctx, p)
	defer sh.cancel()
	game, stop := paceWindow(ctx, sh, Options{})
	defer stop()
	pace, ok := game.(*paceGame)
	if !ok {
		t.Fatalf("desktop loop has no pacing wrapper: %T", game)
	}
	g := &idleSessionGame{pace: pace, page: p}
	p.SetTick(g.tick)
	go g.watchdog(ctx)
	ebiten.SetWindowSize(128, 96)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if !g.done {
		t.Fatal("window exited without presenting the deadline update")
	}
}

type idleSessionGame struct {
	pace     *paceGame
	page     *page.Page
	deadline time.Time
	armed    bool
	observed bool
	fired    bool
	done     bool
	expired  atomic.Bool
	err      error
}

func (g *idleSessionGame) Layout(w, h int) (int, int) { return g.pace.Layout(w, h) }

func (g *idleSessionGame) Update() error {
	if g.expired.Load() {
		return fmt.Errorf("idle window failed to present its deadline update before watchdog")
	}
	if g.err != nil || g.done {
		return ebiten.Termination
	}
	return g.pace.Update()
}

func (g *idleSessionGame) tick(ctx context.Context) error {
	if g.deadline.IsZero() || g.fired {
		return nil
	}
	now := time.Now()
	if now.Before(g.deadline) {
		g.page.WantFrameAt(g.deadline)
		g.armed = true
		return nil
	}
	g.fired = true
	g.page.SetData("blue")
	g.page.WantFrameAt(now)
	return g.page.Redraw(ctx)
}

func (g *idleSessionGame) Draw(screen *ebiten.Image) {
	g.pace.Draw(screen)
	if g.err != nil || g.done {
		return
	}
	if g.deadline.IsZero() {
		if g.pace.idle {
			if got := idleSessionPixel(screen); got != [4]uint32{0xffff, 0, 0, 0xffff} {
				g.err = fmt.Errorf("initial idle frame was not red: %v", got)
				return
			}
			g.deadline = time.Now().Add(600 * time.Millisecond)
		}
		return
	}
	if !g.armed {
		return
	}
	if !g.fired {
		//nolint:staticcheck // reason: this test must distinguish the blocking minimum-FPS mode from the unthrottled vsync-off mode.
		if !g.pace.idle || ebiten.FPSMode() != ebiten.FPSModeVsyncOffMinimum {
			g.err = fmt.Errorf("declared deadline did not keep the window in event-wait mode")
			return
		}
		if g.observed {
			g.err = fmt.Errorf("window presented before its declared deadline; possible input interference or keep-alive wake")
		}
		g.observed = true
		return
	}
	if !g.observed || g.pace.idle || ebiten.TPS() != g.pace.base {
		g.err = fmt.Errorf("deadline wake did not restore the fast loop and its original TPS")
		return
	}
	//nolint:staticcheck // reason: verify the actual drawing mode after waking, independently of the bookkeeping flag.
	if ebiten.FPSMode() != ebiten.FPSModeVsyncOffMaximum {
		g.err = fmt.Errorf("deadline wake did not restore uncapped drawing")
		return
	}
	if got := idleSessionPixel(screen); got != [4]uint32{0, 0, 0xffff, 0xffff} {
		g.err = fmt.Errorf("deadline update was not painted blue: %v", got)
		return
	}
	g.done = true
}

func idleSessionPixel(screen *ebiten.Image) [4]uint32 {
	r, g, b, a := screen.At(screen.Bounds().Dx()/2, screen.Bounds().Dy()/2).RGBA()
	return [4]uint32{r, g, b, a}
}

func (g *idleSessionGame) watchdog(ctx context.Context) {
	timer := time.NewTimer(5 * time.Second)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		g.expired.Store(true)
		//nolint:staticcheck // reason: wake a stuck event-wait loop to report failure; successful runs never use this watchdog wake.
		ebiten.ScheduleFrame()
	}
}
