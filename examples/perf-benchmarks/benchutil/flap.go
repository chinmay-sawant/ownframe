package benchutil

import (
	"context"
	"time"

	"github.com/chinmay-sawant/ownframe"
)

// flapPipe is one top/bottom pair moving left.
type flapPipe struct {
	x      float64
	gapY   float64
	scored bool
}

// Flap is the Benchmark C game: a tick animation over retained ops. Three
// bound slots paint at most three live pipes; the rest hide.
type Flap struct {
	page   *ownframe.Page
	last   time.Time
	y      float64
	vy     float64
	score  int
	pipes  []flapPipe
	next   float64
	spawns int
	bound  uint64
	bird   *ownframe.DisplayOp
	tops   [3]*ownframe.DisplayOp
	bots   [3]*ownframe.DisplayOp
	text   *ownframe.DisplayOp
	shown  int
}

// NewFlap parses the scene page and registers flap input and the tick.
func NewFlap() (*Flap, error) {
	page, err := ownframe.New(ownframe.Config{DevTools: true,
		Title: "Flappy", HTML: flapHTML,
		Width: 480, Height: 720, MinWidth: 320, MinHeight: 480,
		Perf: true,
	})
	if err != nil {
		return nil, err
	}

	f := &Flap{page: page, shown: -1}
	f.reset()
	page.Handle(ownframe.Handlers{KeyDown: f.onKey, Click: f.onClick})
	page.SetTick(f.Tick)

	return f, nil
}

// reset puts the bird mid-scene with two visible pipes. Both boxes are
// inside the scene, so the clipper keeps their fills animatable: a fully
// clipped box lays out as a noop op the tick cannot draw.
func (f *Flap) reset() {
	f.y, f.vy, f.score = 348, 0, 0
	f.pipes = []flapPipe{{x: 210, gapY: 300}, {x: 390, gapY: 220}}
	f.next, f.spawns = 100, 0
}

// flap gives the bird its upward velocity.
func (f *Flap) flap() { f.vy = flapLift }

// Page returns the page Run and Serve display.
func (f *Flap) Page() *ownframe.Page { return f.page }

// Redraw renders the scene.
func (f *Flap) Redraw(ctx context.Context) error { return f.page.Redraw(ctx) }

// Score returns the current score.
func (f *Flap) Score() int { return f.score }
