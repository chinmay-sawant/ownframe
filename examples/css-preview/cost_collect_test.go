package main

import (
	"context"
	"time"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/chinmay-sawant/ownframe"
)

// costCollector records what one frame of the real window cost. It runs
// the demo's own tick, so the animations step the way they do on screen,
// and reads the window's draw time from the game.
type costCollector struct {
	p         *ownframe.Page
	d         *demo
	v         *view
	g         *costGame
	warm      int
	seen      int
	lastTick  time.Time
	lastAlloc uint64
	frames    []costFrame
}

// costTick runs one frame and records it. The page asks for the next frame
// so the window keeps drawing, and the run ends when the frames are done.
func (c *costCollector) tick(ctx context.Context) error {
	c.seen++
	now := costAlloc()
	start := time.Now()
	if err := c.d.tick(ctx); err != nil {
		return err
	}
	elapsed := time.Since(start)
	frame := costFrame{
		draw:   float64(c.g.last.Microseconds()) / 1000,
		tick:   float64(elapsed.Microseconds()) / 1000,
		alloc:  float64(now-c.lastAlloc) / 1024,
		active: costActive(c.d),
	}
	if !c.lastTick.IsZero() {
		frame.gap = float64(start.Sub(c.lastTick).Microseconds()) / 1000
	}
	c.lastTick = start
	c.lastAlloc = costAlloc()
	c.frames = append(c.frames, frame)
	c.p.SetTick(c.tick)
	c.p.WantFrameIn(0)
	if c.seen >= costWindowFrames {
		return ebiten.Termination
	}

	return nil
}

// costPick returns the frames that had at least target samples running,
// dropping the warm-up frames.
func (c *costCollector) costPick(target int) []costFrame {
	out := []costFrame{}
	for i, f := range c.frames {
		if i < c.warm || f.active < target {
			continue
		}
		out = append(out, f)
	}

	return out
}
