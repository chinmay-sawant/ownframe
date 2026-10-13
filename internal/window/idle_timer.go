package window

import (
	"context"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// timer posts keep-alive frames while the policy is idle. ScheduleFrame is
// concurrent-safe: on the desktop backend it writes to the empty-event pipe
// that the blocked WaitEvents polls, so the next frame runs then. The timer
// runs at half the keep-alive interval, so a post lands within half an
// interval of its due time.
func (g *paceGame) timer(ctx context.Context) {
	t := time.NewTicker(g.policy.keepAlive / 2)
	defer t.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-g.stop:
			return
		case now := <-t.C:
			g.mu.Lock()
			due := g.policy.due(now)
			g.mu.Unlock()

			if due {
				ebiten.ScheduleFrame()
			}
		}
	}
}
