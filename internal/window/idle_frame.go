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
	act := g.policy.Step(now, dirty, input, g.frameDeadline())
	g.mu.Unlock()

	switch act {
	case pacing.ToIdle:
		if !g.noIdle {
			g.applyIdle()
		}
	case pacing.ToFast:
		g.applyFast()
	case pacing.Burst:
		//nolint:staticcheck // reason: ScheduleFrame is the only wake-up for FPSModeVsyncOffMinimum event-wait; SetScreenClearedEveryFrame(false) only skips clearing and keeps the full-rate loop.
		ebiten.ScheduleFrame()
	}

	g.kickTimer()
	g.publish(now)
}

// frameDeadline keeps metrics live even when the page asks for a later frame.
func (g *paceGame) frameDeadline() time.Time {
	want := g.want
	cache := &g.shell.dev.panelCache
	if g.shell.dev.on && cache.valid && (want.IsZero() || cache.next.Before(want)) {
		want = cache.next
	}
	return want
}

// applyIdle leaves the active drawing loop: the backend blocks in WaitEvents, and a
// scheduled post or an input event wakes it. TPS syncs with the frame rate
// there, so one wake-up runs one Update. A plain page only blits its tiles
// while parked. DevTools still draws outlines and metrics, so it keeps the
// bounded replay caches and avoids forcing GC on every short input pause.
func (g *paceGame) applyIdle() {
	if g.idle {
		return
	}

	g.idle = true
	if g.shell != nil && !g.shell.dev.on {
		g.shell.renderer.SetIdle()
	}
	//nolint:staticcheck // reason: needs FPSModeVsyncOffMinimum event-wait; SetVsyncEnabled(false) selects Maximum busy loop and would pin CPU.
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.SetTPS(ebiten.SyncWithFPS)
}

// applyFast restores the active drawing loop (still VSync-off maximum at the
// Ebiten level; the present cap is enforced by the Draw limiter, not VSync)
// and the loop TPS: the capped loop rate when a cap is set, else the TPS the
// app started with. The pixel-cache containers are pre-sized first, so the
// first post-wake paints do not regrow them under the draw lock.
func (g *paceGame) applyFast() {
	if !g.idle {
		return
	}

	g.idle = false
	if g.shell != nil {
		g.shell.renderer.Wake()
	}
	ebiten.SetVsyncEnabled(false)
	if loop := LoopFPS(g.maxFPS); loop > 0 {
		ebiten.SetTPS(loop)
		return
	}
	ebiten.SetTPS(g.base)
}
