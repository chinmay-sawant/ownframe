//go:build !js && !android && !ios

package window

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestWindowDevToolsIdle(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("run make graphics-check for a graphics session")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	p, err := page.New(page.Config{
		Width: 1000, Height: 600, DevTools: true,
		HTML: `<body style="margin:0;background:white"><div style="height:600px">Idle dashboard</div></body>`,
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	sh := newShell(ctx, p)
	defer sh.cancel()
	defer sh.renderer.Close()
	defer sh.dev.panelCache.close()
	game, stop := paceWindow(ctx, sh, Options{})
	defer stop()
	pace, ok := game.(*paceGame)
	if !ok {
		t.Fatalf("desktop loop has no pacing wrapper: %T", game)
	}
	g := &devIdleSessionGame{pace: pace}
	var memory runtime.MemStats
	runtime.ReadMemStats(&memory)
	g.forcedGC = memory.NumForcedGC
	watchdog := &idleSessionGame{}
	go watchdog.watchdog(ctx)
	g.watchdog = watchdog
	ebiten.SetWindowSize(1000, 600)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if !g.done {
		t.Fatal("dock did not finish the idle and pixel checks")
	}
}

type devIdleSessionGame struct {
	pace      *paceGame
	watchdog  *idleSessionGame
	idleSince time.Time
	firstNext time.Time
	frames    int
	forcedGC  uint32
	checked   bool
	done      bool
	err       error
}

func (g *devIdleSessionGame) Layout(w, h int) (int, int) { return g.pace.Layout(w, h) }

func (g *devIdleSessionGame) Update() error {
	if g.watchdog.expired.Load() {
		return fmt.Errorf("DevTools prevented idle or stopped scheduled metric frames")
	}
	if g.err != nil || g.done {
		return ebiten.Termination
	}
	return g.pace.Update()
}

func (g *devIdleSessionGame) Draw(screen *ebiten.Image) {
	g.pace.Draw(screen)
	sh := g.pace.shell
	if !g.checked {
		g.err = checkDevPanelPixels(sh, screen)
		g.checked = true
	}
	if g.err != nil {
		return
	}
	if g.idleSince.IsZero() {
		if g.pace.idle {
			var memory runtime.MemStats
			runtime.ReadMemStats(&memory)
			if memory.NumForcedGC != g.forcedGC {
				g.err = fmt.Errorf("opening the dock forced GC when entering idle")
				return
			}
			g.idleSince = time.Now()
			g.firstNext = sh.dev.panelCache.next
		}
		return
	}
	g.frames++
	if !g.pace.idle {
		g.err = fmt.Errorf("metrics refresh returned the idle window to continuous drawing")
		return
	}
	if time.Since(g.idleSince) < 2*devPanelRefreshGap {
		return
	}
	if g.frames > 6 || !sh.dev.panelCache.next.After(g.firstNext) {
		g.err = fmt.Errorf("idle metrics cadence: %d frames, next refresh %v", g.frames, sh.dev.panelCache.next)
		return
	}
	g.done = true
}

func checkDevPanelPixels(sh *shell, screen *ebiten.Image) error {
	bounds := screen.Bounds()
	expected := ebiten.NewImageWithOptions(bounds, &ebiten.NewImageOptions{Unmanaged: true})
	defer expected.Deallocate()
	sh.drawDevPanel(expected)
	panel := sh.dev.panelCache.img.Bounds()
	actualClip, ok := screen.SubImage(panel).(*ebiten.Image)
	if !ok {
		return fmt.Errorf("cannot read cached panel pixels")
	}
	expectedClip, ok := expected.SubImage(panel).(*ebiten.Image)
	if !ok {
		return fmt.Errorf("cannot read reference panel pixels")
	}
	actualPixels := make([]byte, panel.Dx()*panel.Dy()*4)
	expectedPixels := make([]byte, len(actualPixels))
	actualClip.ReadPixels(actualPixels)
	expectedClip.ReadPixels(expectedPixels)
	if !bytes.Equal(actualPixels, expectedPixels) {
		return fmt.Errorf("cached DevTools pixels differ from direct panel rendering")
	}
	return nil
}
