package window

import (
	"time"

	"github.com/chinmay-sawant/ownframe/internal/window/pacing"
	"github.com/hajimehoshi/ebiten/v2"
)

// frame decides this frame's pacing after the shell has drawn.
func (g *paceGame) frame(screen *ebiten.Image, now time.Time) {
	wasIdle := g.idle
	wasTrimmed := g.trimmed
	clean := g.shell.idleCleanFrame()
	chrome := g.chromeChanged(g.shell.frame.restore)
	dirty := !clean || chrome
	if !dirty {
		dirty = g.viewMoved(screen)
	}

	flinging := g.shell.scrollMotion.vx != 0 || g.shell.scrollMotion.vy != 0
	input := g.input || g.watch.held() || flinging
	g.input = false
	if g.idle && !dirty && !input && !g.trimmed && now.Sub(g.idleSince) >= idleTrimDelay {
		g.shell.renderer.SetIdle()
		g.trimmed = true
	}

	g.mu.Lock()
	act := g.policy.Step(now, dirty, input, g.frameDeadline())
	g.mu.Unlock()

	switch act {
	case pacing.ToIdle:
		if !g.noIdle {
			g.applyIdle()
			g.mu.Lock()
			g.policy.SetDeadline(g.frameDeadline())
			g.mu.Unlock()
		}
	case pacing.ToFast:
		g.applyFast()
	case pacing.Burst:
		//nolint:staticcheck // reason: ScheduleFrame is the only wake-up for FPSModeVsyncOffMinimum event-wait; SetScreenClearedEveryFrame(false) only skips clearing and keeps the full-rate loop.
		ebiten.ScheduleFrame()
	}

	g.kickTimer()
	g.publish(now)
	if g.shell.debugEnabled && (wasIdle != g.idle || wasTrimmed != g.trimmed) {
		g.shell.publishDebug()
	}
}

// frameDeadline keeps metrics live even when the page asks for a later frame.
func (g *paceGame) frameDeadline() time.Time {
	want := g.want
	cache := &g.shell.dev.panelCache
	if g.shell.dev.on && cache.valid && !cache.next.IsZero() && (want.IsZero() || cache.next.Before(want)) {
		want = cache.next
	}
	if g.idle && g.eventDriven && !g.trimmed {
		trim := g.idleSince.Add(idleTrimDelay)
		if want.IsZero() || trim.Before(want) {
			want = trim
		}
	}
	return want
}

const idleTrimDelay = 5 * time.Second

// applyIdle leaves the active drawing loop: the backend blocks in WaitEvents, and a
// scheduled post or an input event wakes it. TPS syncs with the frame rate
// there, so one wake-up runs one Update. A plain page only blits its tiles
// while parked. A separate deadline trims caches after sustained inactivity,
// including with DevTools open, rather than forcing GC on every short pause.
func (g *paceGame) applyIdle() {
	if g.idle {
		return
	}

	g.idle = true
	g.idleSince = time.Now()
	g.trimmed = false
	//nolint:staticcheck // reason: needs FPSModeVsyncOffMinimum event-wait; SetVsyncEnabled(false) selects Maximum busy loop and would pin CPU.
	ebiten.SetFPSMode(ebiten.FPSModeVsyncOffMinimum)
	ebiten.SetTPS(ebiten.SyncWithFPS)
}

// applyFast restores the active drawing loop (still VSync-off maximum at the
// Ebiten level; the present cap is enforced by the Draw limiter, not VSync)
// and the loop TPS: the capped loop rate when a cap is set, else SyncWithFPS so input is sampled once per rendered frame. The pixel-cache containers are pre-sized first, so the
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
