package window

import (
	"image"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// frame decides this frame's pacing after the shell has drawn.
func (g *paceGame) frame(screen *ebiten.Image, now time.Time) {
	clean := g.shell.idleCleanFrame()
	chrome := g.chromeChanged(g.shell.frame.restore)
	dirty := !clean || chrome
	if !dirty {
		dirty = g.viewMoved(screen)
	}

	input := g.input
	g.input = false

	g.mu.Lock()
	act := g.policy.step(now, dirty, input)
	g.mu.Unlock()

	switch act {
	case paceToIdle:
		g.applyIdle()
	case paceToFast:
		g.applyFast()
	case paceBurst:
		ebiten.ScheduleFrame()
	}
}

// chromeChanged reports chrome that differs from the previous set and
// records the new set. An unchanged overlay repaints the same pixels, so it
// does not need the vsync loop. The buffers swap, allocating nothing.
func (g *paceGame) chromeChanged(restore []image.Rectangle) bool {
	g.chromeNext = append(g.chromeNext[:0], restore...)
	changed := !slices.Equal(g.chromeNext, g.chrome)
	g.chrome, g.chromeNext = g.chromeNext, g.chrome

	return changed
}

// viewMoved reports a focus or screen size change, either of which makes
// the replay paint in full.
func (g *paceGame) viewMoved(screen *ebiten.Image) bool {
	focus := ebiten.IsFocused()
	size := screen.Bounds().Size()
	moved := !g.seen || focus != g.focus || size != g.size
	g.seen, g.focus, g.size = true, focus, size

	return moved
}

// applyIdle leaves the vsync loop: the backend blocks in WaitEvents, and a
// keep-alive post or an input event wakes it. TPS syncs with the frame rate
// there, so one wake-up runs one Update.
func (g *paceGame) applyIdle() {
	if g.idle {
		return
	}

	g.idle = true
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.SetTPS(ebiten.SyncWithFPS)
}

// applyFast restores the vsync loop and the TPS the app started with.
func (g *paceGame) applyFast() {
	if !g.idle {
		return
	}

	g.idle = false
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOn)
	ebiten.SetTPS(g.base)
}
