//go:build linux && !android && !ios && !js

package window

import (
	"context"
	"fmt"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/host"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestWindowAccessibilityActionWake(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for a graphics session")
	}
	t.Setenv("AT_SPI_BUS_ADDRESS", "disabled")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	p, err := page.New(page.Config{Width: 128, Height: 96,
		HTML: `<body style="margin:0"><button id="go" aria-label="Activate" style="width:128px;height:96px;border:0;background:{{.}}"></button></body>`})
	if err != nil {
		t.Fatal(err)
	}
	p.SetData("red")
	p.Handle(page.Handlers{Click: func(_ context.Context, _ page.Box) error { p.SetData("blue"); return nil }})
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	sh := newShell(ctx, p)
	defer sh.cancel()
	defer sh.closeA11y()
	game, stop := paceWindow(ctx, sh, Options{})
	defer stop()
	pace, ok := game.(*paceGame)
	if !ok {
		t.Fatal("missing desktop pacing wrapper")
	}
	g := &a11yWakeGame{pace: pace, page: p, ctx: ctx, result: make(chan error, 1)}
	p.SetTick(g.tick)
	watchdog := time.AfterFunc(5*time.Second, func() {
		g.expired.Store(true)
		//nolint:staticcheck // reason: watchdog wakes an otherwise stuck test only to report failure.
		ebiten.ScheduleFrame()
	})
	defer watchdog.Stop()
	ebiten.SetWindowSize(128, 96)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil || !g.done {
		t.Fatalf("idle action did not paint: %v", g.err)
	}
	select {
	case err := <-g.result:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("painted action did not complete its reply")
	}
}

type a11yWakeGame struct {
	pace      *paceGame
	page      *page.Page
	ctx       context.Context
	deadline  time.Time
	armed     bool
	submitted bool
	done      bool
	expired   atomic.Bool
	result    chan error
	err       error
}

func (g *a11yWakeGame) Layout(w, h int) (int, int) { return g.pace.Layout(w, h) }

func (g *a11yWakeGame) Update() error {
	if g.expired.Load() {
		return fmt.Errorf("accessibility action failed to wake idle window")
	}
	if g.err != nil || g.done {
		return ebiten.Termination
	}
	return g.pace.Update()
}

func (g *a11yWakeGame) tick(context.Context) error {
	if !g.deadline.IsZero() {
		g.page.WantFrameAt(g.deadline)
		g.armed = true
	}
	return nil
}

func (g *a11yWakeGame) Draw(screen *ebiten.Image) {
	g.pace.Draw(screen)
	if g.done || g.err != nil {
		return
	}
	if g.deadline.IsZero() {
		if g.pace.idle {
			g.deadline = time.Now().Add(2 * time.Second)
		}
		return
	}
	if g.armed && !g.submitted {
		if !g.pace.idle || idleSessionPixel(screen) != [4]uint32{0xffff, 0, 0, 0xffff} {
			g.err = fmt.Errorf("action fixture was not red and idle before submission")
			return
		}
		g.submitted = true
		tree, err := g.page.AccessibilityTree()
		if err != nil {
			g.err = err
			return
		}
		go func() {
			timer := time.NewTimer(100 * time.Millisecond)
			defer timer.Stop()
			select {
			case <-g.ctx.Done():
				return
			case <-timer.C:
			}
			g.result <- g.pace.shell.a11y.actions.Request(g.ctx, host.AccessibilityRequest{
				Document: tree.Document, Node: "id:go", Kind: host.AccessibilityActivate,
			})
		}()
		return
	}
	if g.submitted && idleSessionPixel(screen) == [4]uint32{0, 0, 0xffff, 0xffff} {
		if !time.Now().Before(g.deadline) {
			g.err = fmt.Errorf("action relied on the later page deadline to wake")
			return
		}
		g.done = true
	}
}
