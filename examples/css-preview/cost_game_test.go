package main

import (
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// costGame wraps the real window and times its draw calls. The window
// publishes its draw time to the page only with the devtools dock open, so
// the harness reads the time here instead.
type costGame struct {
	inner ebiten.Game
	last  time.Duration
	draws int
}

func (g *costGame) Update() error { return g.inner.Update() }

func (g *costGame) Draw(screen *ebiten.Image) {
	start := time.Now()
	g.inner.Draw(screen)
	g.last = time.Since(start)
	g.draws++
}

func (g *costGame) Layout(w, h int) (int, int) { return g.inner.Layout(w, h) }
