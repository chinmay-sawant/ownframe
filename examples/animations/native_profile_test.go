//go:build !js

package main

import (
	"context"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

// TestNativeScrollProfile drives the shipped window loop with VSync off.
// Set OWNFRAME_ANIMATIONS_GPU_TEST=1 to retain timings and bottom-page pixels
// under temp/animations-native-profile at the repository root.
func TestNativeScrollProfile(t *testing.T) {
	if os.Getenv("OWNFRAME_ANIMATIONS_GPU_TEST") != "1" {
		t.Skip("set OWNFRAME_ANIMATIONS_GPU_TEST=1 for the native scroll workload")
	}
	dir := filepath.Join("..", "..", "temp", "animations-native-profile")
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatal(err)
	}
	app, err := newApp(true, true)
	if err != nil {
		t.Fatal(err)
	}
	p := app.Page()
	p.SetSize(1908, 999)
	if err := p.Redraw(context.Background()); err != nil {
		t.Fatal(err)
	}
	game := &scrollProfileGame{Game: ownframe.NewGame(context.Background(), p), page: p, dir: dir}
	ebiten.SetWindowSize(1908, 999)
	ebiten.SetWindowTitle("Animation scroll profile")
	ebiten.SetScreenClearedEveryFrame(false)
	ebiten.SetVsyncEnabled(false)
	if err := ebiten.RunGame(game); err != nil {
		t.Fatal(err)
	}
	if game.captureErr != nil {
		t.Fatal(game.captureErr)
	}
	if !game.bottom || !game.top {
		t.Fatalf("scroll traversal incomplete: bottom=%v top=%v", game.bottom, game.top)
	}
	runtime.GC()
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)
	data, err := json.MarshalIndent(struct {
		Frames []scrollProfileFrame
		Memory []scrollProfileMemory
		Heap   uint64
		Total  uint64
	}{Frames: game.frames, Memory: game.memory, Heap: mem.HeapAlloc, Total: mem.TotalAlloc}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "frames.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	runtime.KeepAlive(game)
}

type scrollProfileFrame struct {
	Tick   int
	Y      int
	GapMS  float64
	WorkMS float64
}

type scrollProfileMemory struct {
	Tick int
	Heap uint64
	FPS  float64
	TPS  float64
}

type scrollProfileGame struct {
	ebiten.Game
	page       *ownframe.Page
	dir        string
	tick       int
	maxY       int
	requestY   int
	last       time.Time
	work       time.Duration
	frames     []scrollProfileFrame
	memory     []scrollProfileMemory
	bottom     bool
	top        bool
	captured   bool
	captureErr error
}

func (g *scrollProfileGame) Update() error {
	g.tick++
	if g.tick > 600 {
		return ebiten.Termination
	}
	if g.tick == 60 {
		_, h := g.page.Size()
		for _, b := range g.page.Boxes() {
			g.maxY = max(g.maxY, int(b.Y+b.H)-h)
		}
	}
	if g.tick >= 60 && g.tick%6 == 0 {
		switch {
		case g.tick < 180, g.tick >= 420 && g.tick < 540:
			g.requestY = min(g.maxY, g.requestY+240)
		case g.tick >= 240 && g.tick < 360:
			g.requestY = max(0, g.requestY-240)
		}
		g.page.ScrollTo(0, g.requestY)
	}
	if g.tick%60 == 0 {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)
		g.memory = append(g.memory, scrollProfileMemory{
			Tick: g.tick, Heap: mem.HeapAlloc, FPS: ebiten.ActualFPS(), TPS: ebiten.ActualTPS(),
		})
	}
	start := time.Now()
	err := g.Game.Update()
	g.work += time.Since(start)
	_, y := g.page.Scroll().ScrollOffset()
	g.bottom = g.bottom || g.maxY > 0 && y >= g.maxY
	g.top = g.top || g.bottom && y == 0
	return err
}

func (g *scrollProfileGame) Draw(dst *ebiten.Image) {
	start := time.Now()
	gap := start.Sub(g.last)
	if g.last.IsZero() {
		gap = 0
	}
	g.last = start
	g.Game.Draw(dst)
	_, y := g.page.Scroll().ScrollOffset()
	g.frames = append(g.frames, scrollProfileFrame{
		Tick: g.tick, Y: y,
		GapMS:  float64(gap) / float64(time.Millisecond),
		WorkMS: float64(g.work+time.Since(start)) / float64(time.Millisecond),
	})
	g.work = 0
	if !g.captured && g.tick >= 190 {
		g.captured = true
		bounds := dst.Bounds()
		img := image.NewRGBA(bounds)
		dst.ReadPixels(img.Pix)
		g.captureErr = g.saveImage(img)
		g.last = time.Now()
	}
}

func (g *scrollProfileGame) saveImage(img image.Image) error {
	f, err := os.Create(filepath.Join(g.dir, "bottom.png"))
	if err != nil {
		return err
	}
	if err := png.Encode(f, img); err != nil {
		_ = f.Close()
		return err
	}
	return f.Close()
}
