package window

import (
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
	act := g.policy.step(now, dirty, input, g.want)
	g.mu.Unlock()

	switch act {
	case paceToIdle:
		g.applyIdle()
	case paceToFast:
		g.applyFast()
	case paceBurst:
		ebiten.ScheduleFrame()
	}

	g.kickTimer()
	g.publish(now)
}

// applyIdle leaves the vsync loop: the backend blocks in WaitEvents, and a
// scheduled post or an input event wakes it. TPS syncs with the frame rate
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
