package window

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
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
	act := g.policy.Step(now, dirty, input, g.want)
	g.mu.Unlock()

	switch act {
	case pacing.ToIdle:
		g.applyIdle()
	case pacing.ToFast:
		g.applyFast()
	case pacing.Burst:
		//nolint:staticcheck // reason: ScheduleFrame is the only wake-up for FPSModeVsyncOffMinimum event-wait; SetScreenClearedEveryFrame(false) only skips clearing and keeps the full-rate loop.
		ebiten.ScheduleFrame()
	}

	g.kickTimer()
	g.publish(now)
}

// applyIdle leaves the active drawing loop: the backend blocks in WaitEvents, and a
// scheduled post or an input event wakes it. TPS syncs with the frame rate
// there, so one wake-up runs one Update. The replay pixel caches are released
// first: parked keep-alive frames only blit the tile buffers, so the bakes
// would sit resident until the next dirty paint.
func (g *paceGame) applyIdle() {
	if g.idle {
		return
	}

	g.idle = true
	if g.shell != nil {
		g.shell.renderer.SetIdle()
	}
	//nolint:staticcheck // reason: needs FPSModeVsyncOffMinimum event-wait; SetVsyncEnabled(false) selects Maximum busy loop and would pin CPU.
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.SetTPS(ebiten.SyncWithFPS)
}

// applyFast restores uncapped drawing and the TPS the app started with. The
// pixel-cache containers are pre-sized first, so the first post-wake paints
// do not regrow them under the draw lock.
func (g *paceGame) applyFast() {
	if !g.idle {
		return
	}

	g.idle = false
	if g.shell != nil {
		g.shell.renderer.Wake()
	}
	ebiten.SetVsyncEnabled(false)
	ebiten.SetTPS(g.base)
}
