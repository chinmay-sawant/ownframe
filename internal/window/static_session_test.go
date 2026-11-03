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

func TestWindowStaticEvents(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, err := page.New(page.Config{Width: 128, Height: 96, HTML: `<body style="margin:0"><div style="width:128px;height:96px;background:{{.}}"></div></body>`})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData("red")
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	sh := newShell(ctx, p)
	defer sh.cancel()
	defer sh.renderer.Close()
	game, stop := paceWindow(ctx, sh, Options{})
	defer stop()
	pace, ok := game.(*paceGame)
	if !ok {
		t.Fatalf("missing pacer: %T", game)
	}
	g := &staticSessionGame{pace: pace, page: p}
	watchdog := &idleSessionGame{}
	g.watchdog = watchdog
	go watchdog.watchdog(ctx)
	ebiten.SetWindowSize(128, 96)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if !g.done {
		t.Fatal("notification did not paint the updated page")
	}
}

type staticSessionGame struct {
	pace         *paceGame
	page         *page.Page
	watchdog     *idleSessionGame
	paints       atomic.Uint64
	quietUpdates atomic.Uint64
	ready        atomic.Bool
	armed, done  bool
	applied      bool
	err          error
}

func (g *staticSessionGame) Layout(w, h int) (int, int) { return g.pace.Layout(w, h) }
func (g *staticSessionGame) Update() error {
	if g.watchdog.expired.Load() {
		return fmt.Errorf("static notification timed out")
	}
	if g.done || g.err != nil {
		return ebiten.Termination
	}
	if g.ready.Load() {
		g.page.SetData("blue")
		if err := g.page.Redraw(context.Background()); err != nil {
			return err
		}
		g.applied = true
	}
	return g.pace.Update()
}
func (g *staticSessionGame) Draw(screen *ebiten.Image) {
	g.pace.Draw(screen)
	g.paints.Store(g.pace.presents)
	if !g.armed && g.pace.idle {
		g.armed = true
		go func() {
			// Let already-posted startup events drain before measuring silence.
			time.Sleep(100 * time.Millisecond)
			before := g.paints.Load()
			time.Sleep(2600 * time.Millisecond)
			g.quietUpdates.Store(g.paints.Load() - before)
			g.ready.Store(true)
			g.page.WantFrameIn(0)
		}()
		return
	}
	if !g.applied {
		return
	}
	if n := g.quietUpdates.Load(); n != 0 {
		g.err = fmt.Errorf("static page ran %d host draws during quiet interval", n)
		return
	}
	if got := idleSessionPixel(screen); got != [4]uint32{0, 0, 0xffff, 0xffff} {
		g.err = fmt.Errorf("notified page did not paint blue: %v", got)
		return
	}
	g.done = true
	//nolint:staticcheck // reason: the completed graphics test must wake once to terminate.
	ebiten.ScheduleFrame()
}
