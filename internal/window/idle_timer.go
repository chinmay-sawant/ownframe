package window

import (
	"context"
	"github.com/chinmay-sawant/ownframe/internal/host"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// timer posts scheduled frames while the policy is idle. ScheduleFrame is
// concurrent-safe: on the desktop backend it writes to the empty-event pipe
// that the blocked WaitEvents polls, so the next frame runs then. The timer
// sleeps until the policy's next wake: a page deadline when the page asked,
// the keep-alive fallback when it did not, and half a keep-alive on the fast
// path, where the active drawing loop already supplies every frame. The game loop
// kicks the timer after every step, so a policy change never waits out a
// stale sleep.
func (g *paceGame) timer(ctx context.Context) {
	t := time.NewTimer(time.Hour)
	defer t.Stop()
	resetTimer(t, g.sleep())
	var events <-chan struct{}
	if scheduler, ok := g.shell.app.(host.FrameScheduler); ok {
		events = scheduler.FrameEvents()
	}

	for {
		select {
		case <-ctx.Done():
			//nolint:staticcheck // reason: cancellation must release an indefinitely parked desktop loop.
			ebiten.ScheduleFrame()
			return
		case <-g.stop:
			return
		case <-events:
			g.wake.Store(true)
			//nolint:staticcheck // reason: a page notification must wake the desktop event wait.
			ebiten.ScheduleFrame()
		case <-g.kick:
			resetTimer(t, g.sleep())
		case now := <-t.C:
			g.mu.Lock()
			due := g.policy.Due(now)
			g.mu.Unlock()

			if due {
				g.wake.Store(true)
				//nolint:staticcheck // reason: timer wake needs ScheduleFrame to post the empty event; SetScreenClearedEveryFrame(false) cannot wake event-wait.
				ebiten.ScheduleFrame()
			}

			resetTimer(t, g.sleep())
		}
	}
}

// sleep reads the next timer sleep under the policy lock.
func (g *paceGame) sleep() time.Duration {
	g.mu.Lock()
	defer g.mu.Unlock()

	return g.policy.Wait(time.Now())
}

// resetTimer re-arms t, draining a value that fired before the stop. Only the
// timer goroutine touches t.
func resetTimer(t *time.Timer, d time.Duration) {
	if !t.Stop() {
		select {
		case <-t.C:
		default:
		}
	}

	if d > 0 {
		t.Reset(d)
	}
}
