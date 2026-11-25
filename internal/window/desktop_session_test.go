//go:build !js && !android && !ios

package window

import (
	"bytes"
	"context"
	"fmt"
	"github.com/chinmay-sawant/ownframe/internal/page"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"os"
	"runtime"
	"sync/atomic"
	"testing"
	"time"
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
	wantTPS := g.pace.base
	if loop := LoopFPS(g.pace.maxFPS); loop > 0 {
		wantTPS = loop
	}
	if !g.observed || g.pace.idle || ebiten.TPS() != wantTPS {
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

func TestWindowSharedHeaderPixels(t *testing.T) {
	if os.Getenv("OWNFRAME_FRAME_GPU_TEST") != "1" {
		t.Skip("graphics session required")
	}
	ctx := context.Background()
	p, err := page.New(page.Config{Width: 320, Height: 80, HTML: `<body style="margin:0;background:#ff0000"><div style="height:80px"></div></body>`})
	if err != nil {
		t.Fatal(err)
	}
	if err = p.Redraw(ctx); err != nil {
		t.Fatal(err)
	}
	sh := newShell(ctx, p)
	defer sh.cancel()
	face, err := newHeaderFont()
	if err != nil {
		t.Fatal(err)
	}
	sh.titlebar = titlebarState{enabled: true, style: headerStyle{height: 48, font: face, background: color.NRGBA{B: 255, A: 255}}}
	defer func() {
		if sh.titlebar.canvas != nil {
			sh.titlebar.canvas.Deallocate()
		}
	}()
	game := &headerPixelGame{shell: sh}
	ebiten.SetWindowSize(320, 128)
	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}
	if game.err != nil {
		t.Fatal(game.err)
	}
}

type headerPixelGame struct {
	shell *shell
	done  bool
	err   error
}

func (g *headerPixelGame) Layout(w, h int) (int, int) { return g.shell.Layout(w, h) }
func (g *headerPixelGame) Update() error {
	if g.done {
		return ebiten.Termination
	}
	return g.shell.Update()
}
func (g *headerPixelGame) Draw(screen *ebiten.Image) {
	g.shell.Draw(screen)
	hr, hg, hb, ha := screen.At(2, 2).RGBA()
	cr, cg, cb, ca := screen.At(2, 49).RGBA()
	if hr != 0 || hg != 0 || hb != 0xffff || ha != 0xffff || cr != 0xffff || cg != 0 || cb != 0 || ca != 0xffff {
		g.err = fmt.Errorf("header=(%d,%d,%d,%d) content=(%d,%d,%d,%d); expected blue header above red page", hr, hg, hb, ha, cr, cg, cb, ca)
	}
	g.done = true
}
