//go:build !js

package main

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"os"
	"testing"
	"time"

	"github.com/chinmay-sawant/ownframe"
	"github.com/hajimehoshi/ebiten/v2"
)

func TestNativeDelayAndFillReplay(t *testing.T) {
	if os.Getenv("OWNFRAME_ANIMATIONS_GPU_TEST") != "1" {
		t.Skip("set OWNFRAME_ANIMATIONS_GPU_TEST=1 for native replay pixels")
	}
	app, err := newApp(true, false)
	if err != nil {
		t.Fatal(err)
	}
	const dir = "../../temp/animations-card-replay"
	if err := os.MkdirAll(dir, 0750); err != nil {
		t.Fatal(err)
	}
	g := &cardReplayGame{
		Game: ownframe.NewGame(context.Background(), app.Page()),
		page: app.Page(), t: t, capture: scrollProfileGame{dir: dir},
	}
	ebiten.SetWindowSize(defaultWidth, defaultHeight)
	ebiten.SetScreenClearedEveryFrame(false)
	if err := ebiten.RunGame(g); err != nil {
		t.Fatal(err)
	}
	if g.err != nil {
		t.Fatal(g.err)
	}
	if g.step != 8 {
		t.Fatalf("completed %d replay captures, want 8", g.step)
	}
}

type cardReplayGame struct {
	warmup time.Time
	ready  bool
	ebiten.Game
	page    *ownframe.Page
	t       *testing.T
	capture scrollProfileGame
	step    int
	started time.Time
	initial *image.RGBA
	region  image.Rectangle
	err     error
}

func (g *cardReplayGame) Update() error {
	if !g.ready || g.step == 0 && g.started.IsZero() && time.Since(g.warmup) < 3*time.Second {
		return g.Game.Update()
	}
	if g.err != nil || g.step == 8 {
		return ebiten.Termination
	}
	if g.started.IsZero() {
		ids := []string{"show-delay-positive", "show-fill-none", "show-fill-forwards", "show-fill-both"}
		id := ids[g.step/2]
		button := boxByID(g.t, g.page, "play-"+id)
		g.page.Probe().BeginRedrawBatch()
		g.err = g.page.Click(context.Background(), button.X+button.W/2, button.Y+button.H/2)
		if err := g.page.Probe().EndRedrawBatch(context.Background()); g.err == nil {
			g.err = err
		}
		card := boxByID(g.t, g.page, "card-"+id)
		const top = 120
		g.page.ScrollTo(0, int(card.Y)-top)
		g.region = image.Rect(int(card.X)+15, top+15, int(card.X+card.W)-15, top+103)
		g.started = time.Now()
		g.initial = nil
	}
	if g.err != nil {
		return ebiten.Termination
	}
	return g.Game.Update()
}

func (g *cardReplayGame) Draw(dst *ebiten.Image) {
	g.Game.Draw(dst)
	if !g.ready {
		g.ready = true
		g.warmup = time.Now()
		return
	}
	if g.err != nil || g.step == 8 || g.started.IsZero() {
		return
	}
	elapsed := time.Since(g.started)
	if g.initial != nil && elapsed < time.Second {
		return
	}
	frame := image.NewRGBA(dst.Bounds())
	dst.ReadPixels(frame.Pix)
	if g.initial == nil {
		g.initial = frame
		g.err = g.capture.saveImage(frame, fmt.Sprintf("card-%02d-start.png", g.step))
		return
	}
	changed := false
	for y := g.region.Min.Y; y < g.region.Max.Y; y++ {
		start := frame.PixOffset(g.region.Min.X, y)
		end := frame.PixOffset(g.region.Max.X, y)
		if !bytes.Equal(frame.Pix[start:end], g.initial.Pix[start:end]) {
			changed = true
			break
		}
	}
	if !changed {
		g.err = fmt.Errorf("card replay %d changed its label but not its stage pixels", g.step)
		return
	}
	g.err = g.capture.saveImage(frame, fmt.Sprintf("card-%02d-moving.png", g.step))
	g.step++
	g.started = time.Time{}
}
