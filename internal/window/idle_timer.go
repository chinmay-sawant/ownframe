package window

import (
	"context"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// timer posts scheduled frames while the policy is idle. ScheduleFrame is
// concurrent-safe: on the desktop backend it writes to the empty-event pipe
// that the blocked WaitEvents polls, so the next frame runs then. The timer
// sleeps until the policy's next wake: a page deadline when the page asked,
// the keep-alive fallback when it did not, and half a keep-alive on the fast
// path, where the vsync loop already supplies every frame. The game loop
// kicks the timer after every step, so a policy change never waits out a
// stale sleep.
func (g *paceGame) timer(ctx context.Context) {
	t := time.NewTimer(g.policy.keepAlive / 2)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-g.stop:
			return
		case <-g.kick:
			resetTimer(t, g.sleep())
		case now := <-t.C:
			g.mu.Lock()
			due := g.policy.due(now)
			g.mu.Unlock()

			if due {
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

	return g.policy.wait(time.Now())
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

	t.Reset(d)
}
